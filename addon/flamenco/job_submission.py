# SPDX-License-Identifier: GPL-3.0-or-later
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .manager import ApiClient as _ApiClient
    from .manager.models import (
        Job as _Job,
        SubmittedJob as _SubmittedJob,
    )
else:
    _ApiClient = object
    _Job = object
    _SubmittedJob = object


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
