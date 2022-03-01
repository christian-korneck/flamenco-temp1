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

import logging

import bpy

_flamenco_client = None
_log = logging.getLogger(__name__)


def flamenco_api_client(manager_url="http://localhost:8080"):
    """Returns an API client for communicating with a Manager."""
    global _flamenco_client

    if _flamenco_client is not None:
        return _flamenco_client

    from . import dependencies

    dependencies.preload_modules()

    from flamenco import manager

    configuration = manager.Configuration(host=manager_url.rstrip("/"))
    _flamenco_client = manager.ApiClient(configuration)
    _log.info("created API client for Manager at %s", manager_url)

    return _flamenco_client


def discard_flamenco_data():
    global _flamenco_client

    if _flamenco_client is None:
        return

    _log.info("closing Flamenco client")
    _flamenco_client.close()
    _flamenco_client = None
