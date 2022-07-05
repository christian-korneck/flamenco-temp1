# SPDX-License-Identifier: GPL-3.0-or-later
# <pep8 compliant>

import datetime
import logging
from pathlib import Path, PurePosixPath
from typing import Optional, TYPE_CHECKING
from urllib3.exceptions import HTTPError, MaxRetryError

import bpy

from . import job_types, job_submission, preferences
from .job_types_propgroup import JobTypePropertyGroup

if TYPE_CHECKING:
    from .bat.interface import (
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

        scene = context.scene
        old_job_type_name = getattr(scene, "flamenco_job_type", "")

        try:
            job_types.fetch_available_job_types(api_client, scene)
        except ApiException as ex:
            self.report({"ERROR"}, "Error getting job types: %s" % ex)
            return {"CANCELLED"}
        except MaxRetryError as ex:
            # This is the common error, when for example the port number is
            # incorrect and nothing is listening.
            self.report({"ERROR"}, "Unable to reach Manager")
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
        from flamenco.manager.models import FlamencoVersion, ManagerConfiguration

        context.window_manager.flamenco_status_ping = "..."

        meta_api = MetaApi(api_client)
        try:
            version: FlamencoVersion = meta_api.get_version()
            config: ManagerConfiguration = meta_api.get_configuration()
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
            report = "%s version %s found" % (version.name, version.version)
            level = "INFO"

            # Store whether this Manager supports the Shaman API.
            prefs = preferences.get(context)
            prefs.is_shaman_enabled = config.shaman_enabled
            prefs.job_storage = config.storage_location

        self.report({level}, report)
        context.window_manager.flamenco_status_ping = report
        return {"FINISHED"}


class FLAMENCO_OT_eval_setting(FlamencoOpMixin, bpy.types.Operator):
    bl_idname = "flamenco.eval_setting"
    bl_label = "Flamenco: Evalutate Setting Value"
    bl_description = "Automatically determine a suitable value"
    bl_options = {"REGISTER", "INTERNAL", "UNDO"}

    setting_key: bpy.props.StringProperty(name="Setting Key")  # type: ignore
    setting_eval: bpy.props.StringProperty(name="Python Expression")  # type: ignore

    def execute(self, context: bpy.types.Context) -> set[str]:
        job = job_submission.job_for_scene(context.scene)
        if job is None:
            self.report({"ERROR"}, "This Scene has no Flamenco job")
            return {"CANCELLED"}

        propgroup: JobTypePropertyGroup = context.scene.flamenco_job_settings
        propgroup.eval_and_assign(context, job, self.setting_key, self.setting_eval)
        return {"FINISHED"}


class FLAMENCO_OT_submit_job(FlamencoOpMixin, bpy.types.Operator):
    bl_idname = "flamenco.submit_job"
    bl_label = "Flamenco: Submit Job"
    bl_description = "Pack the current blend file and send it to Flamenco"
    bl_options = {"REGISTER"}  # No UNDO.

    blendfile_on_farm: Optional[PurePosixPath] = None
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
        from .bat import interface as bat_interface

        if bat_interface.is_packing():
            self.report({"ERROR"}, "Another packing operation is running")
            self._quit(context)
            return {"CANCELLED"}

        prefs = preferences.get(context)
        if prefs.is_shaman_enabled:
            # self.blendfile_on_farm will be set when BAT created the checkout,
            # see _on_bat_pack_msg() below.
            self.blendfile_on_farm = None
            self._bat_pack_shaman(context, blendfile)
        else:
            self.blendfile_on_farm = self._bat_pack_filesystem(context, blendfile)

        context.window_manager.modal_handler_add(self)
        wm = context.window_manager
        self.timer = wm.event_timer_add(0.25, window=context.window)

        return {"RUNNING_MODAL"}

    def _bat_pack_filesystem(
        self, context: bpy.types.Context, blendfile: Path
    ) -> PurePosixPath:
        """Use BAT to store the pack on the filesystem.

        :return: the path of the blend file, for use in the job definition.
        """
        from .bat import interface as bat_interface

        # TODO: get project path from addon preferences / project definition on Manager.
        project_path = blendfile.parent
        try:
            project_path = Path(bpy.path.abspath(str(project_path))).resolve()
        except FileNotFoundError:
            # Path.resolve() will raise a FileNotFoundError if the project path doesn't exist.
            self.report({"ERROR"}, "Project path %s does not exist" % project_path)
            raise  # TODO: handle this properly.

        # Determine where the blend file will be stored.
        unique_dir = "%s-%s" % (
            datetime.datetime.now().isoformat("-").replace(":", ""),
            self.job_name,
        )
        prefs = preferences.get(context)
        pack_target_dir = Path(prefs.job_storage) / unique_dir

        # TODO: this should take the blendfile location relative to the project path into account.
        pack_target_file = pack_target_dir / blendfile.name
        self.log.info("Will store blend file at %s", pack_target_file)

        self.packthread = bat_interface.copy(
            base_blendfile=blendfile,
            project=project_path,
            target=str(pack_target_dir),
            exclusion_filter="",  # TODO: get from GUI.
            relative_only=True,  # TODO: get from GUI.
        )

        return PurePosixPath(pack_target_file.as_posix())

    def _bat_pack_shaman(
        self, context: bpy.types.Context, blendfile: Path
    ) -> None:
        """Use the Manager's Shaman API to submit the BAT pack.

        :return: the filesystem path of the blend file, for in the render job definition.
        """
        from .bat import (
            interface as bat_interface,
            shaman as bat_shaman,
        )

        assert self.job is not None
        self.log.info("Sending BAT pack to Shaman")

        # TODO: get project name from preferences/GUI and insert that here too.
        checkout_root = PurePosixPath(f"{self.job.name}")

        self.packthread = bat_interface.copy(
            base_blendfile=blendfile,
            project=blendfile.parent,  # TODO: get from preferences/GUI.
            target="/",  # Target directory irrelevant for Shaman transfers.
            exclusion_filter="",  # TODO: get from GUI.
            relative_only=True,  # TODO: get from GUI.
            packer_class=bat_shaman.Packer,
            packer_kwargs=dict(
                api_client=self.get_api_client(context),
                checkout_path=checkout_root,
            ),
        )

        # We cannot assume the blendfile location is known until the Shaman
        # checkout has actually been created.

    def _on_bat_pack_msg(self, context: bpy.types.Context, msg: _Message) -> set[str]:
        from .bat import interface as bat_interface

        if isinstance(msg, bat_interface.MsgDone):
            if self.blendfile_on_farm is None:
                # Adjust the blendfile to match the Shaman checkout path. Shaman
                # may have checked out at a different location than we
                # requested.
                #
                # Manager automatically creates a variable "jobs" that will
                # resolve to the job storage directory.
                self.blendfile_on_farm = PurePosixPath("{jobs}") / msg.output_path

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
        assert self.blendfile_on_farm is not None

        api_client = self.get_api_client(context)

        propgroup = getattr(context.scene, "flamenco_job_settings", None)
        assert isinstance(propgroup, JobTypePropertyGroup), "did not expect %s" % (
            type(propgroup)
        )
        propgroup.eval_hidden_settings_of_job(context, self.job)

        job_submission.set_blend_file(
            propgroup.job_type, self.job, self.blendfile_on_farm
        )

        submitted_job = job_submission.submit_job(self.job, api_client)
        self.report({"INFO"}, "Job %s submitted" % submitted_job.name)

    def _quit(self, context: bpy.types.Context) -> set[str]:
        """Stop any timer and return a 'FINISHED' status.

        Does neither check nor abort the BAT pack thread.
        """
        if self.timer is not None:
            context.window_manager.event_timer_remove(self.timer)
            self.timer = None
        return {"FINISHED"}


class FLAMENCO3_OT_explore_file_path(bpy.types.Operator):
    """Opens the given path in a file explorer.

    If the path cannot be found, this operator tries to open its parent.
    """

    bl_idname = "flamenco3.explore_file_path"
    bl_label = "Open in file explorer"
    bl_description = __doc__.rstrip(".")

    path: bpy.props.StringProperty(  # type: ignore
        name="Path", description="Path to explore", subtype="DIR_PATH"
    )

    def execute(self, context):
        import platform
        import pathlib

        # Possibly open a parent of the path
        to_open = pathlib.Path(self.path)
        while to_open.parent != to_open:  # while we're not at the root
            if to_open.exists():
                break
            to_open = to_open.parent
        else:
            self.report(
                {"ERROR"}, "Unable to open %s or any of its parents." % self.path
            )
            return {"CANCELLED"}

        if platform.system() == "Windows":
            import os

            # Ignore the mypy error here, as os.startfile() only exists on Windows.
            os.startfile(str(to_open))  # type: ignore

        elif platform.system() == "Darwin":
            import subprocess

            subprocess.Popen(["open", str(to_open)])

        else:
            import subprocess

            subprocess.Popen(["xdg-open", str(to_open)])

        return {"FINISHED"}


classes = (
    FLAMENCO_OT_fetch_job_types,
    FLAMENCO_OT_ping_manager,
    FLAMENCO_OT_eval_setting,
    FLAMENCO_OT_submit_job,
    FLAMENCO3_OT_explore_file_path,
)
register, unregister = bpy.utils.register_classes_factory(classes)
