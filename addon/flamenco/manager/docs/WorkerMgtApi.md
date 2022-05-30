# flamenco.manager.WorkerMgtApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**fetch_workers**](WorkerMgtApi.md#fetch_workers) | **GET** /api/worker-mgt/workers | Get list of workers.


# **fetch_workers**
> WorkerList fetch_workers()

Get list of workers.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import worker_mgt_api
from flamenco.manager.model.worker_list import WorkerList
from pprint import pprint
# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = flamenco.manager.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with flamenco.manager.ApiClient() as api_client:
    # Create an instance of the API class
    api_instance = worker_mgt_api.WorkerMgtApi(api_client)

    # example, this endpoint has no required or optional parameters
    try:
        # Get list of workers.
        api_response = api_instance.fetch_workers()
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling WorkerMgtApi->fetch_workers: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

[**WorkerList**](WorkerList.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Known workers |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

