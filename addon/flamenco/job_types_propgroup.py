"""Flamenco Job Type to bpy.props.PropertyGroup conversion."""

# SPDX-License-Identifier: GPL-3.0-or-later

import logging
from typing import TYPE_CHECKING, Callable, Optional, Any

import bpy

_log = logging.getLogger(__name__)

if TYPE_CHECKING:
    from flamenco.manager.models import (
        AvailableJobType as _AvailableJobType,
        AvailableJobSetting as _AvailableJobSetting,
        JobSettings as _JobSettings,
    )
else:
    _AvailableJobType = object
    _AvailableJobSetting = object
    _JobSettings = object


class JobTypePropertyGroup:
    """Mix-in class for PropertyGroups for Flamenco Job Types.

    Use `generate(job_type: AvailableJobType)` to create such a subclass.
    """

    job_type: _AvailableJobType
    """The job type passed to `generate(job_type)`."""

    @classmethod
    def register_property_group(cls):
        bpy.utils.register_class(cls)

    @classmethod
    def unregister_property_group(cls):
        bpy.utils.unregister_class(cls)

    def as_jobsettings(self) -> _JobSettings:
        from flamenco.manager.models import JobSettings

        js: JobSettings = JobSettings()
        for setting in self.job_type.settings:
            js[setting.key] = getattr(self, setting.key)

        return js

    def eval_and_assign(
        self, context: bpy.types.Context, setting_key: str, setting_eval: str
    ) -> None:
        """Evaluate `setting_eval` and assign the result to the job setting."""
        value = self.eval_setting(context, setting_eval)
        setattr(self, setting_key, value)

    @staticmethod
    def eval_setting(context: bpy.types.Context, setting_eval: str) -> Any:
        """Evaluate `setting_eval` and return the result."""

        eval_globals = {
            "bpy": bpy,
            "C": context,
        }
        value = eval(setting_eval, eval_globals, {})
        return value


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


def generate(job_type: _AvailableJobType) -> type[JobTypePropertyGroup]:
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

    for setting in job_type.settings:
        prop = _create_property(job_type, setting)
        pg_type.__annotations__[setting.key] = prop

    assert issubclass(pg_type, JobTypePropertyGroup), "did not expect type %r" % type(
        pg_type
    )

    return pg_type


def _create_property(job_type: _AvailableJobType, setting: _AvailableJobSetting) -> Any:
    """Create a bpy.props property for the given job setting.

    Depending on the setting, will be a StringProperty, EnumProperty, FloatProperty, etc.
    """
    from flamenco.manager.model.available_job_setting import AvailableJobSetting
    from flamenco.manager.model_utils import ModelSimple

    assert isinstance(setting, AvailableJobSetting)

    # Special case: a string property with 'choices' setting. This should translate to an EnumProperty
    prop_type, prop_kwargs = _find_prop_type(job_type, setting)

    assert isinstance(setting.type, ModelSimple)
    value_coerce = _value_coerce[setting.type.to_str()]
    _set_if_available(prop_kwargs, setting, "description")
    _set_if_available(prop_kwargs, setting, "default", transform=value_coerce)
    _set_if_available(prop_kwargs, setting, "subtype", transform=_transform_subtype)

    prop_name = _job_setting_key_to_label(setting.key)
    prop = prop_type(name=prop_name, **prop_kwargs)
    return prop


def _find_prop_type(
    job_type: _AvailableJobType, setting: _AvailableJobSetting
) -> tuple[Any, dict[str, Any]]:
    """Return a tuple (bpy.props.XxxProperty, kwargs for construction)."""

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
    some_dict: dict[Any, Any],
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
