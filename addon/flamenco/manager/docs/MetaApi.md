# flamenco.manager.MetaApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**get_configuration**](MetaApi.md#get_configuration) | **GET** /api/v3/configuration | Get the configuration of this Manager.
[**get_variables**](MetaApi.md#get_variables) | **GET** /api/v3/configuration/variables/{audience}/{platform} | Get the variables of this Manager. Used by the Blender add-on to recognise two-way variables, and for the web interface to do variable replacement based on the browser&#39;s platform. 
[**get_version**](MetaApi.md#get_version) | **GET** /api/v3/version | Get the Flamenco version of this Manager


# **get_configuration**
> ManagerConfiguration get_configuration()

Get the configuration of this Manager.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import meta_api
from flamenco.manager.model.manager_configuration import ManagerConfiguration
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
        # Get the configuration of this Manager.
        api_response = api_instance.get_configuration()
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling MetaApi->get_configuration: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

[**ManagerConfiguration**](ManagerConfiguration.md)

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

# **get_variables**
> ManagerVariables get_variables(audience, platform)

Get the variables of this Manager. Used by the Blender add-on to recognise two-way variables, and for the web interface to do variable replacement based on the browser's platform. 

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import meta_api
from flamenco.manager.model.manager_variable_audience import ManagerVariableAudience
from flamenco.manager.model.manager_variables import ManagerVariables
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
    audience = ManagerVariableAudience("workers") # ManagerVariableAudience | 
    platform = "platform_example" # str | 

    # example passing only required values which don't have defaults set
    try:
        # Get the variables of this Manager. Used by the Blender add-on to recognise two-way variables, and for the web interface to do variable replacement based on the browser's platform. 
        api_response = api_instance.get_variables(audience, platform)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling MetaApi->get_variables: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **audience** | **ManagerVariableAudience**|  |
 **platform** | **str**|  |

### Return type

[**ManagerVariables**](ManagerVariables.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Normal response. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

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

