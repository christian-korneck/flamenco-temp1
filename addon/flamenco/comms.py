# SPDX-License-Identifier: GPL-3.0-or-later

# <pep8 compliant>

import logging
from typing import TYPE_CHECKING

from urllib3.exceptions import HTTPError, MaxRetryError
import bpy

_flamenco_client = None
_log = logging.getLogger(__name__)

if TYPE_CHECKING:
    from flamenco.manager import ApiClient as _ApiClient
    from .preferences import FlamencoPreferences as _FlamencoPreferences
else:
    _ApiClient = object
    _FlamencoPreferences = object


def flamenco_api_client(manager_url: str) -> _ApiClient:
    """Returns an API client for communicating with a Manager."""
    global _flamenco_client

    if _flamenco_client is not None:
        return _flamenco_client

    from . import dependencies

    dependencies.preload_modules()

    from . import manager

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


def ping_manager(context: bpy.types.Context, api_client: _ApiClient, prefs: _FlamencoPreferences) -> tuple[str, str]:
    """Ping the Manager, update preferences, and return a report as string.

    :returns: tuple (report, level). The report will be something like "<name>
        version <version> found", or an error message. The level will be
        'ERROR', 'WARNING', or 'INFO', suitable for reporting via
        `Operator.report()`.
    """

    # Do a late import, so that the API is only imported when actually used.
    from flamenco.manager import ApiException
    from flamenco.manager.apis import MetaApi
    from flamenco.manager.models import FlamencoVersion, ManagerConfiguration

    context.window_manager.flamenco_status_ping = "..."

    meta_api = MetaApi(api_client)
    try:
        version: FlamencoVersion = meta_api.get_version()
        config: ManagerConfiguration = meta_api.get_configuration()
    except ApiException as ex:
        report = "Manager cannot be reached: %s" % ex
        level = "ERROR"
    except MaxRetryError as ex:
        # This is the common error, when for example the port number is
        # incorrect and nothing is listening.
        report = "Manager cannot be reached"
        level = "WARNING"
    except HTTPError as ex:
        report = "Manager cannot be reached: %s" % ex
        level = "ERROR"
    else:
        report = "%s version %s found" % (version.name, version.version)
        level = "INFO"

        # Store whether this Manager supports the Shaman API.
        prefs.is_shaman_enabled = config.shaman_enabled
        prefs.job_storage = config.storage_location

    context.window_manager.flamenco_status_ping = report

    return report, level
