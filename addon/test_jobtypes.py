#!/usr/bin/env python

import sys
from pathlib import Path

my_dir = Path(__file__).parent
sys.path.append(str(my_dir))


import atexit
from flamenco import dependencies, job_types_propgroup as jt_propgroup

dependencies.preload_modules()

import flamenco.manager

from flamenco.manager.api import jobs_api
from flamenco.manager.model.available_job_types import AvailableJobTypes

# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = flamenco.manager.Configuration(host="http://localhost:8080")


api_client = flamenco.manager.ApiClient(configuration)
atexit.register(api_client.close)

job_api_instance = jobs_api.JobsApi(api_client)

try:
    response: AvailableJobTypes = job_api_instance.get_job_types()
except flamenco.manager.ApiException as ex:
    raise SystemExit("Exception when calling JobsApi->fetch_job: %s" % ex)

job_type = next(jt for jt in response.job_types if jt.name == "simple-blender-render")
pg = jt_propgroup.generate(job_type)
