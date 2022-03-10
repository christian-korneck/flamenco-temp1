# flamenco.manager.MetaApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**get_version**](MetaApi.md#get_version) | **GET** /api/version | Get the Flamenco version of this Manager


# **get_version**
> FlamencoVersion get_version()

Get the Flamenco version of this Manager

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import meta_api
from flamenco.manager.model.flamenco_version import FlamencoVersion
from pprint import pprint
# Defining the host is optional and defaults to http://localhost
# See configuration.py for a list of all supported configuration parameters.
configuration = flamenco.manager.Configuration(
    host = "http://localhost"
)


# Enter a context with an instance of the API client
with flamenco.manager.ApiClient() as api_client:
    # Create an instance of the API class
    api_instance = meta_api.MetaApi(api_client)

    # example, this endpoint has no required or optional parameters
    try:
        # Get the Flamenco version of this Manager
        api_response = api_instance.get_version()
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling MetaApi->get_version: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

[**FlamencoVersion**](FlamencoVersion.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | normal response |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

