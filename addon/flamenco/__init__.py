# ##### BEGIN GPL LICENSE BLOCK #####
#
#  This program is free software; you can redistribute it and/or
#  modify it under the terms of the GNU General Public License
#  as published by the Free Software Foundation; either version 2
#  of the License, or (at your option) any later version.
#
#  This program is distributed in the hope that it will be useful,
#  but WITHOUT ANY WARRANTY; without even the implied warranty of
#  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#  GNU General Public License for more details.
#
#  You should have received a copy of the GNU General Public License
#  along with this program; if not, write to the Free Software Foundation,
#  Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.
#
# ##### END GPL LICENSE BLOCK #####

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
    from . import operators, gui, job_types, comms
else:
    import importlib

    operators = importlib.reload(operators)
    gui = importlib.reload(gui)
    job_types = importlib.reload(job_types)
    comms = importlib.reload(comms)

import bpy


@bpy.app.handlers.persistent
def discard_global_flamenco_data(_) -> None:
    job_types.discard_flamenco_data()
    comms.discard_flamenco_data()


def register() -> None:
    from . import dependencies

    dependencies.preload_modules()

    bpy.app.handlers.load_pre.append(discard_global_flamenco_data)
    bpy.app.handlers.load_factory_preferences_post.append(discard_global_flamenco_data)

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
