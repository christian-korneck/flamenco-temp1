# SPDX-License-Identifier: GPL-3.0-or-later

# <pep8 compliant>

bl_info = {
    "name": "Flamenco 3",
    "author": "Sybren A. Stüvel",
    "version": (3, 0),
    "blender": (3, 1, 0),
    "description": "Flamenco client for Blender.",
    "location": "...",
    "wiki_url": "...",
    "category": "System",
    "support": "COMMUNITY",
}

__is_first_load = "operators" not in locals()
if __is_first_load:
    from . import operators, gui, job_types, comms, preferences
else:
    import importlib

    operators = importlib.reload(operators)
    gui = importlib.reload(gui)
    job_types = importlib.reload(job_types)
    comms = importlib.reload(comms)
    preferences = importlib.reload(preferences)

import bpy


@bpy.app.handlers.persistent
def discard_global_flamenco_data(_):
    job_types.discard_flamenco_data()
    comms.discard_flamenco_data()


def redraw(self, context):
    if context.area is None:
        return
    context.area.tag_redraw()


def register() -> None:
    from . import dependencies

    dependencies.preload_modules()

    bpy.app.handlers.load_pre.append(discard_global_flamenco_data)
    bpy.app.handlers.load_factory_preferences_post.append(discard_global_flamenco_data)

    bpy.types.WindowManager.flamenco_bat_status = bpy.props.EnumProperty(
        items=[
            ("IDLE", "IDLE", "Not doing anything."),
            ("SAVING", "SAVING", "Saving your file."),
            ("INVESTIGATING", "INVESTIGATING", "Finding all dependencies."),
            ("TRANSFERRING", "TRANSFERRING", "Transferring all dependencies."),
            ("COMMUNICATING", "COMMUNICATING", "Communicating with Flamenco Server."),
            ("DONE", "DONE", "Not doing anything, but doing something earlier."),
            ("ABORTING", "ABORTING", "User requested we stop doing something."),
            ("ABORTED", "ABORTED", "We stopped doing something."),
        ],
        name="flamenco_status",
        default="IDLE",
        description="Current status of the Flamenco add-on",
        update=redraw,
    )

    bpy.types.WindowManager.flamenco_bat_status_txt = bpy.props.StringProperty(
        name="Flamenco Status",
        default="",
        description="Textual description of what Flamenco is doing",
        update=redraw,
    )

    bpy.types.WindowManager.flamenco_bat_progress = bpy.props.IntProperty(
        name="Flamenco Progress",
        default=0,
        description="File transfer progress",
        subtype="PERCENTAGE",
        min=0,
        max=100,
        update=redraw,
    )

    bpy.types.Scene.flamenco_job_name = bpy.props.StringProperty(
        name="Flamenco Job Name",
        default="",
        description="Name of the Flamenco job; an empty name will use the blend file name as job name",
    )

    # Placeholder to contain the result of a 'ping' to Flamenco Manager,
    # so that it can be shown in the preferences panel.
    bpy.types.WindowManager.flamenco_status_ping = bpy.props.StringProperty()

    preferences.register()
    operators.register()
    gui.register()
    job_types.register()


def unregister() -> None:
    discard_global_flamenco_data(None)
    bpy.app.handlers.load_pre.remove(discard_global_flamenco_data)
    bpy.app.handlers.load_factory_preferences_post.remove(discard_global_flamenco_data)

    job_types.unregister()
    gui.unregister()
    operators.unregister()
    preferences.unregister()
