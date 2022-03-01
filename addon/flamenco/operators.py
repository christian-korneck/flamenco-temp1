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

        try:
            job_types.fetch_available_job_types(api_client)
        except ApiException as ex:
            self.report({"ERROR"}, "Error getting job types: %s" % ex)
            return {"CANCELLED"}

        return {"FINISHED"}


classes = (FLAMENCO_OT_fetch_job_types,)
register, unregister = bpy.utils.register_classes_factory(classes)
