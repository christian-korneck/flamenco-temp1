# SPDX-License-Identifier: GPL-3.0-or-later

# <pep8 compliant>

import bpy


class FLAMENCO_OT_fetch_job_types(bpy.types.Operator):
    bl_idname = "flamenco.fetch_job_types"
    bl_label = "Fetch Job Types"
    bl_description = "Query Flamenco Manager to obtain the available job types."

    def execute(self, context: bpy.types.Context) -> set[str]:
        from . import comms, job_types

        # Getting the client also loads the dependencies, so we can only import
        # API stuff after it.
        api_client = comms.flamenco_api_client()

        from flamenco.manager import ApiException

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


classes = (FLAMENCO_OT_fetch_job_types,)
register, unregister = bpy.utils.register_classes_factory(classes)
