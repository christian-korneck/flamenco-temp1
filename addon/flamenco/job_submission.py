# SPDX-License-Identifier: GPL-3.0-or-later
from pathlib import Path
from typing import TYPE_CHECKING, Optional, Union

import bpy

from .job_types_propgroup import JobTypePropertyGroup

if TYPE_CHECKING:
    from .manager import ApiClient as _ApiClient
    from .manager.models import (
        AvailableJobType as _AvailableJobType,
        Job as _Job,
        SubmittedJob as _SubmittedJob,
    )
else:
    _AvailableJobType = object
    _ApiClient = object
    _Job = object
    _SubmittedJob = object


# If a job has a setting with this key, type 'str', and subtype 'file_path',
# it'll be set to the path of the BAT-packed blend file.
BLENDFILE_SETTING_KEY = "blendfile"


def job_for_scene(scene: bpy.types.Scene) -> Optional[_SubmittedJob]:
    from flamenco.manager.models import SubmittedJob, JobMetadata

    propgroup = getattr(scene, "flamenco_job_settings", None)
    assert isinstance(propgroup, JobTypePropertyGroup), "did not expect %s" % (
        type(propgroup)
    )

    settings = propgroup.as_jobsettings()
    metadata = JobMetadata()

    job: SubmittedJob = SubmittedJob(
        name=scene.flamenco_job_name,
        type=propgroup.job_type.name,
        priority=50,
        settings=settings,
        metadata=metadata,
    )
    return job


def set_blend_file(
    job_type: _AvailableJobType, job: _SubmittedJob, blendfile: Union[str, Path]
) -> None:
    """Update the job's 'blendfile' setting, if available.

    If a job has a 'blendfile' setting, type 'str', it'll be set to the path/URL
    of the BAT-packed blend file.
    """
    from .manager.models import AvailableJobSetting, AvailableJobSettingType

    expected_type = AvailableJobSettingType("string")
    for setting in job_type.settings:
        if setting.key == BLENDFILE_SETTING_KEY and setting.type == expected_type:
            break
    else:
        # Not having this setting is fine.
        return

    assert isinstance(setting, AvailableJobSetting)
    job.settings[BLENDFILE_SETTING_KEY] = str(blendfile)


def submit_job(job: _SubmittedJob, api_client: _ApiClient) -> _Job:
    """Send the given job to Flamenco Manager."""
    from flamenco.manager import ApiClient
    from flamenco.manager.api import jobs_api
    from flamenco.manager.models import SubmittedJob, Job

    assert isinstance(job, SubmittedJob), "got %s" % type(job)
    assert isinstance(api_client, ApiClient), "got %s" % type(api_client)

    job_api_instance = jobs_api.JobsApi(api_client)
    response: Job = job_api_instance.submit_job(job)
    print("Job submitted: %s (%s)" % (response.name, response.id))

    return response
