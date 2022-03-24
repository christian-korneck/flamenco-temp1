# SPDX-License-Identifier: GPL-3.0-or-later
# <pep8 compliant>

import bpy


def discard_flamenco_client(prefs, context):
    """Discard any cached Flamenco client after the Manager URL changes."""
    from . import comms

    comms.discard_flamenco_data()
    context.window_manager.flamenco_status_ping = ""


def _update_default_job_storage(
    prefs: "FlamencoPreferences", context: bpy.types.Context
) -> None:
    _unregister_rna_props()
    _register_rna_props(prefs)


class FlamencoPreferences(bpy.types.AddonPreferences):
    bl_idname = "flamenco"

    manager_url: bpy.props.StringProperty(  # type: ignore
        name="Manager URL",
        description="Location of the Manager",
        default="http://localhost:8080/",
        update=discard_flamenco_client,
    )

    is_shaman_enabled: bpy.props.BoolProperty(  # type: ignore
        name="Shaman Enabled",
        description="Whether this Manager has the Shaman protocol enabled",
        default=False,
    )

    job_storage: bpy.props.StringProperty(  # type: ignore
        name="Job Storage Directory",
        subtype="DIR_PATH",
        default="",
        description="Directory where blend files are stored, when submitting them to Flamenco",
        update=_update_default_job_storage,
    )

    def draw(self, context: bpy.types.Context) -> None:
        layout = self.layout

        col = layout.column()
        col.use_property_split = True

        row = col.row(align=True)
        row.prop(self, "manager_url")
        row.operator("flamenco.ping_manager", text="", icon="CHECKMARK")
        if context.window_manager.flamenco_status_ping:
            col.label(text=context.window_manager.flamenco_status_ping)

        col = layout.column(align=True)
        col.enabled = not self.is_shaman_enabled
        if self.is_shaman_enabled:
            col.label(
                text="This Manager supports the Shaman API, so this setting will be ignored:"
            )
        col.prop(self, "job_storage")


def get(context: bpy.types.Context) -> FlamencoPreferences:
    """Return the add-on preferences."""
    prefs = context.preferences.addons["flamenco"].preferences
    assert isinstance(
        prefs, FlamencoPreferences
    ), "Expected FlamencoPreferences, got %s instead" % (type(prefs))
    return prefs


def manager_url(context: bpy.types.Context) -> str:
    """Returns the configured Manager URL."""
    prefs = get(context)
    return str(prefs.manager_url)


def _register_rna_props(prefs: FlamencoPreferences) -> None:
    """RNA properties that have their defaults set in the preferences get registered here."""

    bpy.types.Scene.flamenco_job_storage = bpy.props.StringProperty(
        name="Flamenco Job Storage",
        subtype="DIR_PATH",
        default=prefs.job_storage,
        description="Directory where blend files are stored, when submitting them to Flamenco",
    )


def _unregister_rna_props() -> None:
    del bpy.types.Scene.flamenco_job_storage


classes = (FlamencoPreferences,)
_register, _unregister = bpy.utils.register_classes_factory(classes)


def register():
    _register()
    _register_rna_props(get(bpy.context))
    bpy.context.window_manager.flamenco_status_ping = ""


def unregister():
    _unregister_rna_props()
    _unregister()
