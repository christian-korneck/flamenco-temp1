# SPDX-License-Identifier: GPL-3.0-or-later
# <pep8 compliant>

import bpy


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

        old_job_type_name = getattr(context.window_manager, "flamenco_job_type", "")

        try:
            job_types.fetch_available_job_types(api_client)
        except ApiException as ex:
            self.report({"ERROR"}, "Error getting job types: %s" % ex)
            return {"CANCELLED"}

        if old_job_type_name:
            # TODO: handle cases where the old job type no longer exists.
            context.window_manager.flamenco_job_type = old_job_type_name

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

        context.window_manager.flamenco_status_ping = "..."

        meta_api = MetaApi(api_client)
        try:
            response: FlamencoVersion = meta_api.get_version()
        except ApiException as ex:
            report = "Manager cannot be reached: %s" % ex
            level = "ERROR"
        else:
            report = "%s version %s found" % (response.name, response.version)
            level = "INFO"

        self.report({level}, report)
        context.window_manager.flamenco_status_ping = report
        return {"FINISHED"}


classes = (
    FLAMENCO_OT_fetch_job_types,
    FLAMENCO_OT_ping_manager,
)
register, unregister = bpy.utils.register_classes_factory(classes)
