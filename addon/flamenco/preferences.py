# SPDX-License-Identifier: GPL-3.0-or-later
# <pep8 compliant>

import bpy


def discard_flamenco_client(prefs, context):
    """Discard any cached Flamenco client after the Manager URL changes."""
    from flamenco import comms

    comms.discard_flamenco_data()


class FlamencoPreferences(bpy.types.AddonPreferences):
    bl_idname = "flamenco"

    manager_url: bpy.props.StringProperty(  # type: ignore
        name="Manager URL",
        description="Location of the Manager",
        default="http://localhost:8080/",
        update=discard_flamenco_client,
    )

    def draw(self, context: bpy.types.Context) -> None:
        layout = self.layout

        col = layout.column()
        row = col.row(align=True)
        row.prop(self, "manager_url")
        row.operator("flamenco.ping_manager", text="Test Connection")
        col.label(text=context.window_manager.flamenco_status_ping)


def manager_url(context: bpy.types.Context) -> str:
    """Returns the configured Manager URL."""
    prefs = context.preferences.addons["flamenco"].preferences
    assert isinstance(
        prefs, FlamencoPreferences
    ), "Expected FlamencoPreferences, got %s instead" % (type(prefs))
    return prefs.manager_url


classes = (FlamencoPreferences,)
register, unregister = bpy.utils.register_classes_factory(classes)
