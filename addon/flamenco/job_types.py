# SPDX-License-Identifier: GPL-3.0-or-later

import logging
import uuid
from typing import TYPE_CHECKING, Callable, Optional, Union

import bpy

from . import job_types_propgroup

_log = logging.getLogger(__name__)

if TYPE_CHECKING:
    from flamenco.manager.models import AvailableJobType, SubmittedJob, JobSettings

    _available_job_types: Optional[list[AvailableJobType]] = None
else:
    SubmittedJob = object
    JobSettings = object
    _available_job_types = None

# Items for a bpy.props.EnumProperty()
_job_type_enum_items: list[
    Union[tuple[str, str, str], tuple[str, str, str, int, int]]
] = []

_selected_job_type_propgroup: Optional[job_types_propgroup.JobTypePropertyGroup] = None


def fetch_available_job_types(api_client):
    global _available_job_types
    global _job_type_enum_items

    from flamenco.manager import ApiClient
    from flamenco.manager.api import jobs_api
    from flamenco.manager.model.available_job_types import AvailableJobTypes

    assert isinstance(api_client, ApiClient)

    job_api_instance = jobs_api.JobsApi(api_client)
    response: AvailableJobTypes = job_api_instance.get_job_types()

    _clear_available_job_types()

    # Remember the available job types.
    _available_job_types = response.job_types
    if _available_job_types is None:
        _job_type_enum_items = []
    else:
        # Convert from API response type to list suitable for an EnumProperty.
        _job_type_enum_items = [
            (job_type.name, job_type.label, "") for job_type in _available_job_types
        ]
    _job_type_enum_items.insert(0, ("", "Select a Job Type", "", 0, 0))


def are_job_types_available() -> bool:
    """Returns whether job types have been fetched and are available."""
    return bool(_job_type_enum_items)


def _update_job_type(scene: bpy.types.Scene, context: bpy.types.Context) -> None:
    """Called whenever the selected job type changes."""
    update_job_type_properties(scene)


def update_job_type_properties(scene: bpy.types.Scene) -> None:
    """(Re)construct the PropertyGroup for the currently selected job type."""

    global _selected_job_type_propgroup

    from flamenco.manager.model.available_job_type import AvailableJobType

    job_type = active_job_type(scene)
    _clear_job_type_propgroup()

    if job_type is None:
        return

    assert isinstance(job_type, AvailableJobType), "did not expect type %r" % type(
        job_type
    )

    pg = job_types_propgroup.generate(job_type)
    pg.register_property_group()
    _selected_job_type_propgroup = pg

    bpy.types.Scene.flamenco_job_settings = bpy.props.PointerProperty(
        type=pg,
        name="Job Settings",
        description="Parameters for the Flamenco job",
    )


def get_job_settings(scene: bpy.types.Scene) -> Optional[JobSettings]:
    job_settings = getattr(scene, "flamenco_job_settings", None)
    if job_settings is None:
        return None
    assert isinstance(job_settings, JobSettings), "expected JobSettings, got %s" % (
        type(job_settings)
    )
    return job_settings


def job_for_scene(scene: bpy.types.Scene) -> Optional[SubmittedJob]:
    from flamenco.manager.models import SubmittedJob, JobSettings, JobMetadata

    settings_propgroup = get_job_settings(scene)
    if settings_propgroup is None:
        return None

    # TODO: convert settings_propgroup to JobSettings.
    # dict(settings_propgroup) only includes the user-modified items, which
    # isn't enough; the JobSettings() object should also have explicit values
    # for the still-default ones.
    settings = JobSettings()
    metadata = JobMetadata()

    job = SubmittedJob(
        name=scene.flamenco_job_name,
        type=settings_propgroup.job_type.name,
        priority=50,
        id=str(uuid.uuid4()),
        settings=settings,
        metadata=metadata,
    )
    return job


def _clear_available_job_types():
    global _available_job_types
    global _job_type_enum_items

    _clear_job_type_propgroup()

    _available_job_types = None
    _job_type_enum_items.clear()


def _clear_job_type_propgroup():
    global _selected_job_type_propgroup

    try:
        del bpy.types.WindowManager.flamenco_job_settings
    except AttributeError:
        pass

    # Make sure there is no old property group reference.
    if _selected_job_type_propgroup is not None:
        _selected_job_type_propgroup.unregister_property_group()
        _selected_job_type_propgroup = None


if TYPE_CHECKING:
    from flamenco.manager.model.available_job_type import (
        AvailableJobType as _AvailableJobType,
    )
else:
    _AvailableJobType = object


def active_job_type(scene: bpy.types.Scene) -> Optional[_AvailableJobType]:
    """Return the active job type.

    Returns a flamenco.manager.model.available_job_type.AvailableJobType,
    or None if there is none.
    """
    if _available_job_types is None:
        return None

    job_type_name = scene.flamenco_job_type
    for job_type in _available_job_types:
        if job_type.name == job_type_name:
            return job_type
    return None


def _get_job_types_enum_items(dummy1, dummy2):
    return _job_type_enum_items


def discard_flamenco_data():
    if _available_job_types:
        _available_job_types.clear()
    if _job_type_enum_items:
        _job_type_enum_items.clear()


def register() -> None:
    bpy.types.Scene.flamenco_job_type = bpy.props.EnumProperty(
        name="Job Type",
        items=_get_job_types_enum_items,
        update=_update_job_type,
    )

    bpy.types.Scene.flamenco_available_job_types_json = bpy.props.StringProperty(
        name="Available Job Types",
    )


def unregister() -> None:
    del bpy.types.Scene.flamenco_job_type

    try:
        # This property doesn't always exist.
        del bpy.types.Scene.flamenco_job_settings
    except AttributeError:
        pass


if __name__ == "__main__":
    import doctest

    print(doctest.testmod())
