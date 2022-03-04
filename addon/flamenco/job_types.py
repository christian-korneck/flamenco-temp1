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

from inspect import isclass
import logging
from typing import TYPE_CHECKING, Callable, Optional

import bpy

_log = logging.getLogger(__name__)


class JobTypePropertyGroup:
    @classmethod
    def register_property_group(cls):
        bpy.utils.register_class(cls)

    @classmethod
    def unregister_property_group(cls):
        bpy.utils.unregister_class(cls)


# Mapping from AvailableJobType.setting.type to a callable that converts a value
# to the appropriate type. This is necessary due to the ambiguity between floats
# and ints in JavaScript (and thus JSON).
_value_coerce = {
    "bool": bool,
    "string": str,
    "int32": int,
    "float": float,
}

_prop_types = {
    "bool": bpy.props.BoolProperty,
    "string": bpy.props.StringProperty,
    "int32": bpy.props.IntProperty,
    "float": bpy.props.FloatProperty,
}


if TYPE_CHECKING:
    from flamenco.manager.model.available_job_type import AvailableJobType

    _available_job_types: Optional[list[AvailableJobType]] = None
else:
    _available_job_types = None

# Items for a bpy.props.EnumProperty()
_job_type_enum_items = []

_selected_job_type_propgroup: Optional[JobTypePropertyGroup] = None


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

    # Convert from API response type to list suitable for an EnumProperty.
    _job_type_enum_items = [
        (job_type.name, job_type.label, "") for job_type in _available_job_types
    ]
    _job_type_enum_items.insert(0, ("", "Select a Job Type", "", 0, 0))


def are_job_types_available() -> bool:
    """Returns whether job types have been fetched and are available."""
    return bool(_job_type_enum_items)


def _update_job_type(
    window_manager: bpy.types.WindowManager, context: bpy.types.Context
) -> None:
    """Called whenever the selected job type changes."""
    global _selected_job_type_propgroup

    from flamenco.manager.model.available_job_type import AvailableJobType

    job_type = active_job_type(window_manager)
    assert isinstance(job_type, AvailableJobType), "did not expect type %r" % type(
        job_type
    )

    _clear_job_type_propgroup()

    pg = generate_property_group(job_type)
    pg.register_property_group()
    _selected_job_type_propgroup = pg

    bpy.types.WindowManager.flamenco_job_settings = bpy.props.PointerProperty(
        type=pg,
        name="Job Settings",
        description="Parameters for the Flamenco job",
    )


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


def active_job_type(window_manager: bpy.types.WindowManager):
    """Return the active job type.

    Returns a flamenco.manager.model.available_job_type.AvailableJobType,
    or None if there is none.
    """
    job_type_name = window_manager.flamenco_job_type
    for job_type in _available_job_types:
        if job_type.name == job_type_name:
            return job_type
    return None


def generate_property_group(job_type):
    """Create a PropertyGroup for the job type.

    Does not register the property group.
    """
    from flamenco.manager.model.available_job_type import AvailableJobType

    assert isinstance(job_type, AvailableJobType)

    classname = _job_type_to_class_name(job_type.name)

    pg_type = type(
        classname,
        (JobTypePropertyGroup, bpy.types.PropertyGroup),  # Base classes.
        {  # Class attributes.
            "job_type": job_type,
        },
    )
    pg_type.__annotations__ = {}

    print(f"\033[38;5;214m{job_type.label}\033[0m ({job_type.name})")
    for setting in job_type.settings:
        prop = _create_property(job_type, setting)
        pg_type.__annotations__[setting.key] = prop

    assert issubclass(pg_type, JobTypePropertyGroup), "did not expect type %r" % type(
        pg_type
    )

    from pprint import pprint

    print(pg_type)
    pprint(pg_type.__annotations__)

    return pg_type


def _create_property(job_type, setting):
    from flamenco.manager.model.available_job_setting import AvailableJobSetting
    from flamenco.manager.model_utils import ModelSimple

    assert isinstance(setting, AvailableJobSetting)

    print(f"  - {setting.key:23}  type: {setting.type!r:10}", end="")

    # Special case: a string property with 'choices' setting. This should translate to an EnumProperty
    prop_type, prop_kwargs = _find_prop_type(job_type, setting)

    assert isinstance(setting.type, ModelSimple)
    value_coerce = _value_coerce[setting.type.to_str()]
    _set_if_available(prop_kwargs, setting, "default", transform=value_coerce)
    _set_if_available(prop_kwargs, setting, "subtype", transform=_transform_subtype)
    print()

    prop_name = _job_setting_key_to_label(setting.key)
    prop = prop_type(name=prop_name, **prop_kwargs)
    return prop


def _find_prop_type(job_type, setting):
    # The special case is a 'string' property with 'choices' setting, which
    # should translate to an EnumProperty. All others just map to a simple
    # bpy.props type.

    setting_type = setting.type.to_str()

    if "choices" not in setting:
        return _prop_types[setting_type], {}

    if setting_type != "string":
        # There was a 'choices' key, but not for a supported type. Ignore the
        # choices but complain about it.
        _log.warn(
            "job type %r, setting %r: only string choices are supported, but property is of type %s",
            job_type.name,
            setting.key,
            setting_type,
        )
        return _prop_types[setting_type], {}

    choices = setting.choices
    enum_items = [(choice, choice, "") for choice in choices]
    return bpy.props.EnumProperty, {"items": enum_items}


def _transform_subtype(subtype: object) -> str:
    uppercase = str(subtype).upper()
    if uppercase == "HASHED_FILE_PATH":
        # Flamenco has a concept of 'hashed file path' subtype, but Blender does not.
        return "FILE_PATH"
    return uppercase


def _job_type_to_class_name(job_type_name: str) -> str:
    """Change 'job-type-name' to 'JobTypeName'.

    >>> _job_type_to_class_name('job-type-name')
    'JobTypeName'
    """
    return job_type_name.title().replace("-", "")


def _job_setting_key_to_label(setting_key: str) -> str:
    """Change 'some_setting_key' to 'Some Setting Key'.

    >>> _job_setting_key_to_label('some_setting_key')
    'Some Setting Key'
    """
    return setting_key.title().replace("_", " ")


def _set_if_available(
    some_dict: dict[object, object],
    setting: object,
    key: str,
    transform: Optional[Callable[[object], object]] = None,
) -> None:
    """some_dict[key] = setting.key, if that key is available.

    >>> class Setting:
    ...     pass
    >>> setting = Setting()
    >>> setting.exists = 47
    >>> d = {}
    >>> _set_if_available(d, setting, "exists")
    >>> _set_if_available(d, setting, "other")
    >>> d
    {'exists': 47}
    >>> d = {}
    >>> _set_if_available(d, setting, "exists", transform=lambda v: str(v))
    >>> d
    {'exists': '47'}
    """
    try:
        value = getattr(setting, key)
    except AttributeError:
        return

    if transform is None:
        some_dict[key] = value
    else:
        some_dict[key] = transform(value)


def _get_job_types_enum_items(dummy1, dummy2):
    return _job_type_enum_items


def discard_flamenco_data():
    if _available_job_types:
        _available_job_types.clear()
    if _job_type_enum_items:
        _job_type_enum_items.clear()


def register() -> None:
    bpy.types.WindowManager.flamenco_job_type = bpy.props.EnumProperty(
        name="Job Type",
        items=_get_job_types_enum_items,
        update=_update_job_type,
    )


def unregister() -> None:
    del bpy.types.WindowManager.flamenco_job_type

    try:
        # This property doesn't always exist.
        del bpy.types.WindowManager.flamenco_job_settings
    except AttributeError:
        pass


if __name__ == "__main__":
    import doctest

    print(doctest.testmod())