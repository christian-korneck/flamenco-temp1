# SPDX-License-Identifier: GPL-3.0-or-later
# <pep8 compliant>

import logging
from pathlib import Path
from typing import Optional, TYPE_CHECKING

import bpy

from . import job_types, job_submission

if TYPE_CHECKING:
    from .bat_interface import (
        PackThread as _PackThread,
        Message as _Message,
    )
    from .manager.models import SubmittedJob as _SubmittedJob
else:
    _PackThread = object
    _Message = object
    _SubmittedJob = object

_log = logging.getLogger(__name__)


class FlamencoOpMixin:
    @staticmethod
    def get_api_client(context):
        """Get a Flamenco API client to talk to the Manager.

        Getting the client also loads the dependencies, so only import things
        from `flamenco.manager` after calling this function.
        """
        from . import comms, preferences

        manager_url = preferences.manager_url(context)
        api_client = comms.flamenco_api_client(manager_url)
        return api_client


class FLAMENCO_OT_fetch_job_types(FlamencoOpMixin, bpy.types.Operator):
    bl_idname = "flamenco.fetch_job_types"
    bl_label = "Fetch Job Types"
    bl_description = "Query Flamenco Manager to obtain the available job types"

    def execute(self, context: bpy.types.Context) -> set[str]:
        api_client = self.get_api_client(context)

        from flamenco.manager import ApiException
        from . import job_types

        scene = context.scene
        old_job_type_name = getattr(scene, "flamenco_job_type", "")

        try:
            job_types.fetch_available_job_types(api_client, scene)
        except ApiException as ex:
            self.report({"ERROR"}, "Error getting job types: %s" % ex)
            return {"CANCELLED"}

        if old_job_type_name:
            # TODO: handle cases where the old job type no longer exists.
            scene.flamenco_job_type = old_job_type_name

        job_types.update_job_type_properties(scene)
        return {"FINISHED"}


class FLAMENCO_OT_ping_manager(FlamencoOpMixin, bpy.types.Operator):
    bl_idname = "flamenco.ping_manager"
    bl_label = "Flamenco: Ping Manager"
    bl_description = "Attempt to connect to the Manager"
    bl_options = {"REGISTER"}  # No UNDO.

    def execute(self, context: bpy.types.Context) -> set[str]:
        api_client = self.get_api_client(context)

        from flamenco.manager import ApiException
        from flamenco.manager.apis import MetaApi
        from flamenco.manager.models import FlamencoVersion
        from urllib3.exceptions import HTTPError, MaxRetryError

        context.window_manager.flamenco_status_ping = "..."

        meta_api = MetaApi(api_client)
        try:
            response: FlamencoVersion = meta_api.get_version()
        except ApiException as ex:
            report = "Manager cannot be reached: %s" % ex
            level = "ERROR"
        except MaxRetryError as ex:
            # This is the common error, when for example the port number is
            # incorrect and nothing is listening.
            report = "Manager cannot be reached"
            level = "WARNING"
        except HTTPError as ex:
            report = "Manager cannot be reached: %s" % ex
            level = "ERROR"
        else:
            report = "%s version %s found" % (response.name, response.version)
            level = "INFO"

        self.report({level}, report)
        context.window_manager.flamenco_status_ping = report
        return {"FINISHED"}


class FLAMENCO_OT_submit_job(FlamencoOpMixin, bpy.types.Operator):
    bl_idname = "flamenco.submit_job"
    bl_label = "Flamenco: Submit Job"
    bl_description = "Pack the current blend file and send it to Flamenco"
    bl_options = {"REGISTER"}  # No UNDO.

    blendfile_on_farm: Optional[Path] = None
    job_name: bpy.props.StringProperty(name="Job Name")  # type: ignore
    job: Optional[_SubmittedJob] = None

    timer: Optional[bpy.types.Timer] = None
    packthread: Optional[_PackThread] = None

    log = _log.getChild(bl_idname)

    def invoke(self, context: bpy.types.Context, event: bpy.types.Event) -> set[str]:
        filepath = self._save_blendfile(context)

        # Construct the Job locally before trying to pack. If any validations fail, better fail early.
        self.job = job_submission.job_for_scene(context.scene)
        if self.job is None:
            self.report({"ERROR"}, "Unable to create job")
            return {"CANCELLED"}

        return self._bat_pack(context, filepath)

    def modal(self, context: bpy.types.Context, event: bpy.types.Event) -> set[str]:
        # This function is called for TIMER events to poll the BAT pack thread.
        if event.type != "TIMER":
            return {"PASS_THROUGH"}

        if self.packthread is None:
            # If there is no pack thread running, there isn't much we can do.
            return self._quit(context)

        msg = self.packthread.poll()
        if not msg:
            return {"RUNNING_MODAL"}

        return self._on_bat_pack_msg(context, msg)

    def _save_blendfile(self, context):
        """Save to a different file, specifically for Flamenco.

        We shouldn't overwrite the artist's file.
        We can compress, since this file won't be managed by SVN and doesn't need diffability.
        """
        render = context.scene.render

        # Remember settings we need to restore after saving.
        old_use_file_extension = render.use_file_extension
        old_use_overwrite = render.use_overwrite
        old_use_placeholder = render.use_placeholder

        # TODO: see about disabling the denoiser (like the old Blender Cloud addon did).

        try:
            # The file extension should be determined by the render settings, not necessarily
            # by the setttings in the output panel.
            render.use_file_extension = True

            # Rescheduling should not overwrite existing frames.
            render.use_overwrite = False
            render.use_placeholder = False

            filepath = Path(context.blend_data.filepath).with_suffix(".flamenco.blend")
            self.log.info("Saving copy to temporary file %s", filepath)
            bpy.ops.wm.save_as_mainfile(
                filepath=str(filepath), compress=True, copy=True
            )
        finally:
            # Restore the settings we changed, even after an exception.
            render.use_file_extension = old_use_file_extension
            render.use_overwrite = old_use_overwrite
            render.use_placeholder = old_use_placeholder

        return filepath

    def _bat_pack(self, context: bpy.types.Context, blendfile: Path) -> set[str]:
        from . import bat_interface

        if bat_interface.is_packing():
            self.report({"ERROR"}, "Another packing operation is running")
            self._quit(context)
            return {"CANCELLED"}

        # TODO: get project path from addon preferences / project definition on Manager.
        project_path = blendfile.parent
        try:
            project_path = Path(bpy.path.abspath(str(project_path))).resolve()
        except FileNotFoundError:
            # Path.resolve() will raise a FileNotFoundError if the project path doesn't exist.
            self.report({"ERROR"}, "Project path %s does not exist" % project_path)
            return {"CANCELLED"}

        # Determine where the render output will be stored.
        pack_target_dir = Path("/render/_flamenco/tests/renders") / self.job_name

        # TODO: this should take the blendfile location relative to the project path into account.
        pack_target_file = pack_target_dir / blendfile.name
        self.log.info("Will store blend file at %s", pack_target_file)
        self.blendfile_on_farm = pack_target_file

        self.packthread = bat_interface.copy(
            base_blendfile=blendfile,
            project=project_path,
            target=pack_target_dir,
            exclusion_filter="",  # TODO: get from GUI.
            relative_only=True,  # TODO: get from GUI.
        )

        context.window_manager.modal_handler_add(self)
        wm = context.window_manager
        self.timer = wm.event_timer_add(0.25, window=context.window)

        return {"RUNNING_MODAL"}

    def _on_bat_pack_msg(self, context: bpy.types.Context, msg: _Message) -> set[str]:
        from . import bat_interface

        if isinstance(msg, bat_interface.MsgDone):
            self._submit_job(context)
            return self._quit(context)

        if isinstance(msg, bat_interface.MsgException):
            self.log.error("Error performing BAT pack: %s", msg.ex)
            self.report({"ERROR"}, "Error performing BAT pack")

            # This was an exception caught at the top level of the thread, so
            # the packing thread itself has stopped.
            return self._quit(context)

        if isinstance(msg, bat_interface.MsgSetWMAttribute):
            wm = context.window_manager
            setattr(wm, msg.attribute_name, msg.value)

        return {"RUNNING_MODAL"}

    def _submit_job(self, context: bpy.types.Context) -> None:
        """Use the Flamenco API to submit the new Job."""
        assert self.job is not None

        job_type = job_types.active_job_type(context.scene)
        job_submission.set_blend_file(job_type, self.job, self.blendfile_on_farm)

        api_client = self.get_api_client(context)
        job = job_submission.submit_job(self.job, api_client)

        self.report({"INFO"}, "Job %s submitted" % job.name)

    def _quit(self, context: bpy.types.Context) -> set[str]:
        """Stop any timer and return a 'FINISHED' status.

        Does neither check nor abort the BAT pack thread.
        """
        if self.timer is not None:
            context.window_manager.event_timer_remove(self.timer)
            self.timer = None
        return {"FINISHED"}


classes = (
    FLAMENCO_OT_fetch_job_types,
    FLAMENCO_OT_ping_manager,
    FLAMENCO_OT_submit_job,
)
register, unregister = bpy.utils.register_classes_factory(classes)
