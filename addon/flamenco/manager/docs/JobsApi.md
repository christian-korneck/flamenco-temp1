# flamenco.manager.JobsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**fetch_job**](JobsApi.md#fetch_job) | **GET** /api/jobs/{job_id} | Fetch info about the job.
[**get_job_types**](JobsApi.md#get_job_types) | **GET** /api/jobs/types | Get list of job types and their parameters.
[**submit_job**](JobsApi.md#submit_job) | **POST** /api/jobs | Submit a new job for Flamenco Manager to execute.


# **fetch_job**
> Job fetch_job(job_id)

Fetch info about the job.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.job import Job
from pprint import pprint
# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = flamenco.manager.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with flamenco.manager.ApiClient() as api_client:
    # Create an instance of the API class
    api_instance = jobs_api.JobsApi(api_client)
    job_id = "job_id_example" # str | 

    # example passing only required values which don't have defaults set
    try:
        # Fetch info about the job.
        api_response = api_instance.fetch_job(job_id)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->fetch_job: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **job_id** | **str**|  |

### Return type

[**Job**](Job.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Job info |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_job_types**
> AvailableJobTypes get_job_types()

Get list of job types and their parameters.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.available_job_types import AvailableJobTypes
from pprint import pprint
# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = flamenco.manager.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with flamenco.manager.ApiClient() as api_client:
    # Create an instance of the API class
    api_instance = jobs_api.JobsApi(api_client)

    # example, this endpoint has no required or optional parameters
    try:
        # Get list of job types and their parameters.
        api_response = api_instance.get_job_types()
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->get_job_types: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

[**AvailableJobTypes**](AvailableJobTypes.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Available job types |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **submit_job**
> Job submit_job(submitted_job)

Submit a new job for Flamenco Manager to execute.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.submitted_job import SubmittedJob
from flamenco.manager.model.error import Error
from flamenco.manager.model.job import Job
from pprint import pprint
# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = flamenco.manager.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with flamenco.manager.ApiClient() as api_client:
    # Create an instance of the API class
    api_instance = jobs_api.JobsApi(api_client)
    submitted_job = SubmittedJob(
        name="name_example",
        type="type_example",
        priority=50,
        settings=JobSettings(),
        metadata=JobMetadata(
            key="key_example",
        ),
    ) # SubmittedJob | Job to submit

    # example passing only required values which don't have defaults set
    try:
        # Submit a new job for Flamenco Manager to execute.
        api_response = api_instance.submit_job(submitted_job)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->submit_job: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **submitted_job** | [**SubmittedJob**](SubmittedJob.md)| Job to submit |

### Return type

[**Job**](Job.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Job was succesfully compiled into individual tasks. |  -  |
**0** | Error message |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

