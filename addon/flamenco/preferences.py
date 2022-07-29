# SPDX-License-Identifier: GPL-3.0-or-later
# <pep8 compliant>

import bpy


def discard_flamenco_client(context):
    """Discard any cached Flamenco client after the Manager URL changes."""
    from . import comms

    comms.discard_flamenco_data()
    context.window_manager.flamenco_status_ping = ""


def _refresh_the_planet(
    prefs: "FlamencoPreferences", context: bpy.types.Context
) -> None:
    """Refresh all GUI areas."""
    for win in context.window_manager.windows:
        for area in win.screen.areas:
            for region in area.regions:
                region.tag_redraw()


def _manager_url_updated(prefs, context):
    discard_flamenco_client(context)

    from . import comms

    api_client = comms.flamenco_api_client(prefs.manager_url)

    # Warning, be careful what of the context to access here. Accessing /
    # changing too much can cause crashes, infinite loops, etc.
    comms.ping_manager(context, api_client, prefs)


class FlamencoPreferences(bpy.types.AddonPreferences):
    bl_idname = "flamenco"

    manager_url: bpy.props.StringProperty(  # type: ignore
        name="Manager URL",
        description="Location of the Manager",
        default="http://localhost:8080/",
        update=_manager_url_updated,
    )

    is_shaman_enabled: bpy.props.BoolProperty(  # type: ignore
        name="Shaman Enabled",
        description="Whether this Manager has the Shaman protocol enabled",
        default=False,
        update=_refresh_the_planet,
    )

    # Property that should be editable from Python. It's not exposed to the GUI.
    job_storage: bpy.props.StringProperty(  # type: ignore
        name="Job Storage Directory",
        subtype="DIR_PATH",
        default="",
        options={"HIDDEN"},
        description="Directory where blend files are stored when submitting them to Flamenco. This value is determined by Flamenco Manager",
    )

    # Property that gets its value from the above _job_storage, and cannot be
    # set. This makes it read-only in the GUI.
    job_storage_for_gui: bpy.props.StringProperty(  # type: ignore
        name="Job Storage Directory",
        subtype="DIR_PATH",
        default="",
        options={"SKIP_SAVE"},
        description="Directory where blend files are stored when submitting them to Flamenco. This value is determined by Flamenco Manager",
        get=lambda prefs: prefs.job_storage,
    )

    def draw(self, context: bpy.types.Context) -> None:
        layout = self.layout
        layout.use_property_decorate = False
        layout.use_property_split = True

        col = layout.column()

        row = col.row(align=True)
        row.prop(self, "manager_url")
        row.operator("flamenco.ping_manager", text="", icon="FILE_REFRESH")

        def text_row(parent, label):
            split = parent.split(factor=0.4)
            split.label(text="")
            split.label(text=label)

        if not self.job_storage:
            text_row(col, "Press the refresh button before using Flamenco")

        if context.window_manager.flamenco_status_ping:
            text_row(col, context.window_manager.flamenco_status_ping)
        else:
            aligned = col.column(align=True)
            text_row(aligned, "Press the refresh button to check the connection")
            text_row(aligned, "and update the job storage location")

        if self.is_shaman_enabled:
            text_row(col, "Shaman enabled")
        col.prop(self, "job_storage_for_gui", text="Job Storage")


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


classes = (FlamencoPreferences,)
_register, _unregister = bpy.utils.register_classes_factory(classes)


def register():
    _register()
    bpy.context.window_manager.flamenco_status_ping = ""


def unregister():
    _unregister()
