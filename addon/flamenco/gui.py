# SPDX-License-Identifier: GPL-3.0-or-later

# <pep8 compliant>
import bpy


class FLAMENCO_PT_job_submission(bpy.types.Panel):
    bl_space_type = "PROPERTIES"
    bl_region_type = "WINDOW"
    bl_context = "output"
    bl_label = "Flamenco 3"

    def draw(self, context: bpy.types.Context) -> None:
        from . import job_types

        layout = self.layout
        col = layout.column(align=True)

        if not job_types.are_job_types_available():
            col.operator("flamenco.fetch_job_types", icon="FILE_REFRESH")
            return

        row = col.row(align=True)
        row.prop(context.scene, "flamenco_job_type", text="")
        row.operator("flamenco.fetch_job_types", text="", icon="FILE_REFRESH")
        self.draw_job_settings(context, layout)

        layout.separator()
        col = layout.column(align=True)
        col.prop(context.scene, "flamenco_job_name", text="Job Name")

        self.draw_flamenco_status(context, layout)

    def draw_job_settings(
        self, context: bpy.types.Context, layout: bpy.types.UILayout
    ) -> None:
        from . import job_types

        job_type = job_types.active_job_type(context.scene)
        if job_type is None:
            return

        propgroup = getattr(context.scene, "flamenco_job_settings", None)
        if propgroup is None:
            return

        for setting in job_type.settings:
            if not setting.get("visible", True):
                continue
            layout.prop(propgroup, setting.key)

    def draw_flamenco_status(
        self, context: bpy.types.Context, layout: bpy.types.UILayout
    ) -> None:
        # Show current status of Flamenco.
        flamenco_status = context.window_manager.flamenco_bat_status
        if flamenco_status in {"IDLE", "ABORTED", "DONE"}:
            ui = layout
            props = ui.operator(
                "flamenco.submit_job",
                text="Submit to Flamenco",
                icon="RENDER_ANIMATION",
            )
            props.job_name = context.scene.flamenco_job_name
        elif flamenco_status == "INVESTIGATING":
            row = layout.row(align=True)
            row.label(text="Investigating your files")
            # row.operator(FLAMENCO_OT_abort.bl_idname, text="", icon="CANCEL")
        elif flamenco_status == "COMMUNICATING":
            layout.label(text="Communicating with Flamenco Server")
        elif flamenco_status == "ABORTING":
            row = layout.row(align=True)
            row.label(text="Aborting, please wait.")
            # row.operator(FLAMENCO_OT_abort.bl_idname, text="", icon="CANCEL")
        if flamenco_status == "TRANSFERRING":
            row = layout.row(align=True)
            row.prop(
                context.window_manager,
                "flamenco_bat_progress",
                text=context.window_manager.flamenco_bat_status_txt,
            )
            # row.operator(FLAMENCO_OT_abort.bl_idname, text="", icon="CANCEL")
        elif (
            flamenco_status != "IDLE" and context.window_manager.flamenco_bat_status_txt
        ):
            layout.label(text=context.window_manager.flamenco_bat_status_txt)


classes = (FLAMENCO_PT_job_submission,)
register, unregister = bpy.utils.register_classes_factory(classes)
