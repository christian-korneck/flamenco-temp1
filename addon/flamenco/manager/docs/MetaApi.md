# flamenco.manager.MetaApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**check_shared_storage_path**](MetaApi.md#check_shared_storage_path) | **POST** /api/v3/configuration/check/shared-storage | Validate a path for use as shared storage.
[**get_configuration**](MetaApi.md#get_configuration) | **GET** /api/v3/configuration | Get the configuration of this Manager.
[**get_configuration_file**](MetaApi.md#get_configuration_file) | **GET** /api/v3/configuration/file | Retrieve the configuration of Flamenco Manager.
[**get_variables**](MetaApi.md#get_variables) | **GET** /api/v3/configuration/variables/{audience}/{platform} | Get the variables of this Manager. Used by the Blender add-on to recognise two-way variables, and for the web interface to do variable replacement based on the browser&#39;s platform. 
[**get_version**](MetaApi.md#get_version) | **GET** /api/v3/version | Get the Flamenco version of this Manager


# **check_shared_storage_path**
> PathCheckResult check_shared_storage_path()

Validate a path for use as shared storage.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import meta_api
from flamenco.manager.model.error import Error
from flamenco.manager.model.path_check_result import PathCheckResult
from flamenco.manager.model.path_check_input import PathCheckInput
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
    path_check_input = PathCheckInput(
        path="path_example",
    ) # PathCheckInput | Path to check (optional)

    # example passing only required values which don't have defaults set
    # and optional values
    try:
        # Validate a path for use as shared storage.
        api_response = api_instance.check_shared_storage_path(path_check_input=path_check_input)
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling MetaApi->check_shared_storage_path: %s\n" % e)
```


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **path_check_input** | [**PathCheckInput**](PathCheckInput.md)| Path to check | [optional]

### Return type

[**PathCheckResult**](PathCheckResult.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Normal response, path check went fine. |  -  |
**0** | Something went wrong. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

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

# **get_configuration_file**
> {str: (bool, date, datetime, dict, float, int, list, str, none_type)} get_configuration_file()

Retrieve the configuration of Flamenco Manager.

### Example


```python
import time
import flamenco.manager
from flamenco.manager.api import meta_api
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
        # Retrieve the configuration of Flamenco Manager.
        api_response = api_instance.get_configuration_file()
        pprint(api_response)
    except flamenco.manager.ApiException as e:
        print("Exception when calling MetaApi->get_configuration_file: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

**{str: (bool, date, datetime, dict, float, int, list, str, none_type)}**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/yaml


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | Normal response. |  -  |

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

