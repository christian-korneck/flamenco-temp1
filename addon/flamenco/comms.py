# SPDX-License-Identifier: GPL-3.0-or-later

# <pep8 compliant>

import logging
from typing import TYPE_CHECKING

_flamenco_client = None
_log = logging.getLogger(__name__)

if TYPE_CHECKING:
    from flamenco.manager import ApiClient as _ApiClient
else:
    _ApiClient = object


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
