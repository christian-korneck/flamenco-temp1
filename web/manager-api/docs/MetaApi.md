# flamencoManager.MetaApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getConfiguration**](MetaApi.md#getConfiguration) | **GET** /api/configuration | Get the configuration of this Manager.
[**getVersion**](MetaApi.md#getVersion) | **GET** /api/version | Get the Flamenco version of this Manager



## getConfiguration

> ManagerConfiguration getConfiguration()

Get the configuration of this Manager.

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.MetaApi();
apiInstance.getConfiguration().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

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


## getVersion

> FlamencoVersion getVersion()

Get the Flamenco version of this Manager

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.MetaApi();
apiInstance.getVersion().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

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

