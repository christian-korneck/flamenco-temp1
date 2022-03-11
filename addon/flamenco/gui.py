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


classes = (FLAMENCO_PT_job_submission,)
register, unregister = bpy.utils.register_classes_factory(classes)
