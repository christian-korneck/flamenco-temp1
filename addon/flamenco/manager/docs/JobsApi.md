# flamenco.manager.JobsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**fetch_job**](JobsApi.md#fetch_job) | **GET** /api/jobs/{job_id} | Fetch info about the job.
[**fetch_job_tasks**](JobsApi.md#fetch_job_tasks) | **GET** /api/jobs/{job_id}/tasks | Fetch a summary of all tasks of the given job.
[**fetch_task**](JobsApi.md#fetch_task) | **GET** /api/tasks/{task_id} | Fetch a single task.
[**fetch_task_log_tail**](JobsApi.md#fetch_task_log_tail) | **GET** /api/tasks/{task_id}/logtail | Fetch the last few lines of the task&#39;s log.
[**get_job_type**](JobsApi.md#get_job_type) | **GET** /api/jobs/type/{typeName} | Get single job type and its parameters.
[**get_job_types**](JobsApi.md#get_job_types) | **GET** /api/jobs/types | Get list of job types and their parameters.
[**query_jobs**](JobsApi.md#query_jobs) | **POST** /api/jobs/query | Fetch list of jobs.
[**set_job_status**](JobsApi.md#set_job_status) | **POST** /api/jobs/{job_id}/setstatus | 
[**set_task_status**](JobsApi.md#set_task_status) | **POST** /api/tasks/{task_id}/setstatus | 
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

# **fetch_job_tasks**
> JobTasksSummary fetch_job_tasks(job_id)

Fetch a summary of all tasks of the given job.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.error import Error
from flamenco.manager.model.job_tasks_summary import JobTasksSummary
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
        # Fetch a summary of all tasks of the given job.
        api_response = api_instance.fetch_job_tasks(job_id)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->fetch_job_tasks: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **job_id** | **str**|  |

### Return type

[**JobTasksSummary**](JobTasksSummary.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Get summaries of the tasks of this job. |  -  |
**0** | Unexpected error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **fetch_task**
> Task fetch_task(task_id)

Fetch a single task.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.task import Task
from flamenco.manager.model.error import Error
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
    task_id = "task_id_example" # str | 

    # example passing only required values which don't have defaults set
    try:
        # Fetch a single task.
        api_response = api_instance.fetch_task(task_id)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->fetch_task: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**|  |

### Return type

[**Task**](Task.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The task info. |  -  |
**0** | Unexpected error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **fetch_task_log_tail**
> str fetch_task_log_tail(task_id)

Fetch the last few lines of the task's log.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.error import Error
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
    task_id = "task_id_example" # str | 

    # example passing only required values which don't have defaults set
    try:
        # Fetch the last few lines of the task's log.
        api_response = api_instance.fetch_task_log_tail(task_id)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->fetch_task_log_tail: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**|  |

### Return type

**str**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/plain, application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | The task log. |  -  |
**0** | Unexpected error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_job_type**
> AvailableJobType get_job_type(type_name)

Get single job type and its parameters.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.available_job_type import AvailableJobType
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
    type_name = "typeName_example" # str | 

    # example passing only required values which don't have defaults set
    try:
        # Get single job type and its parameters.
        api_response = api_instance.get_job_type(type_name)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->get_job_type: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **type_name** | **str**|  |

### Return type

[**AvailableJobType**](AvailableJobType.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Job type |  -  |

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

# **query_jobs**
> JobsQueryResult query_jobs(jobs_query)

Fetch list of jobs.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.error import Error
from flamenco.manager.model.jobs_query import JobsQuery
from flamenco.manager.model.jobs_query_result import JobsQueryResult
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
    jobs_query = JobsQuery(
        offset=0,
        limit=1,
        order_by=[
            "order_by_example",
        ],
        status_in=[
            JobStatus("active"),
        ],
        metadata={
            "key": "key_example",
        },
        settings={},
    ) # JobsQuery | Specification of which jobs to get.

    # example passing only required values which don't have defaults set
    try:
        # Fetch list of jobs.
        api_response = api_instance.query_jobs(jobs_query)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->query_jobs: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **jobs_query** | [**JobsQuery**](JobsQuery.md)| Specification of which jobs to get. |

### Return type

[**JobsQueryResult**](JobsQueryResult.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Normal query response, can be empty list if nothing matched the query. |  -  |
**0** | Error message |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **set_job_status**
> set_job_status(job_id, job_status_change)



### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.error import Error
from flamenco.manager.model.job_status_change import JobStatusChange
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
    job_status_change = JobStatusChange(
        status=JobStatus("active"),
        reason="reason_example",
    ) # JobStatusChange | The status change to request.

    # example passing only required values which don't have defaults set
    try:
        api_instance.set_job_status(job_id, job_status_change)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->set_job_status: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **job_id** | **str**|  |
 **job_status_change** | [**JobStatusChange**](JobStatusChange.md)| The status change to request. |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**204** | Status change was accepted. |  -  |
**422** | The requested status change is not valid for the current status of the job. |  -  |
**0** | Unexpected error. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **set_task_status**
> set_task_status(task_id, task_status_change)



### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import jobs_api
from flamenco.manager.model.task_status_change import TaskStatusChange
from flamenco.manager.model.error import Error
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
    task_id = "task_id_example" # str | 
    task_status_change = TaskStatusChange(
        status=TaskStatus("active"),
        reason="reason_example",
    ) # TaskStatusChange | The status change to request.

    # example passing only required values which don't have defaults set
    try:
        api_instance.set_task_status(task_id, task_status_change)
    except flamenco.manager.ApiException as e:
        print("Exception when calling JobsApi->set_task_status: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **str**|  |
 **task_status_change** | [**TaskStatusChange**](TaskStatusChange.md)| The status change to request. |

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**204** | Status change was accepted. |  -  |
**422** | The requested status change is not valid for the current status of the task. |  -  |
**0** | Unexpected error. |  -  |

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

