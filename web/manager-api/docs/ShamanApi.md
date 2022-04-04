# flamencoManager.ShamanApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**shamanCheckout**](ShamanApi.md#shamanCheckout) | **POST** /shaman/checkout/create | Create a directory, and symlink the required files into it. The files must all have been uploaded to Shaman before calling this endpoint.
[**shamanCheckoutRequirements**](ShamanApi.md#shamanCheckoutRequirements) | **POST** /shaman/checkout/requirements | Checks a Shaman Requirements file, and reports which files are unknown.
[**shamanFileStore**](ShamanApi.md#shamanFileStore) | **POST** /shaman/files/{checksum}/{filesize} | Store a new file on the Shaman server. Note that the Shaman server can forcibly close the HTTP connection when another client finishes uploading the exact same file, to prevent double uploads. The file&#39;s contents should be sent in the request body. 
[**shamanFileStoreCheck**](ShamanApi.md#shamanFileStoreCheck) | **GET** /shaman/files/{checksum}/{filesize} | Check the status of a file on the Shaman server. 



## shamanCheckout

> ShamanCheckoutResult shamanCheckout(ShamanCheckout)

Create a directory, and symlink the required files into it. The files must all have been uploaded to Shaman before calling this endpoint.

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.ShamanApi();
let ShamanCheckout = new flamencoManager.ShamanCheckout(); // ShamanCheckout | Set of files to check out.
apiInstance.shamanCheckout(ShamanCheckout).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ShamanCheckout** | [**ShamanCheckout**](ShamanCheckout.md)| Set of files to check out. | 

### Return type

[**ShamanCheckoutResult**](ShamanCheckoutResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## shamanCheckoutRequirements

> ShamanRequirementsResponse shamanCheckoutRequirements(ShamanRequirementsRequest)

Checks a Shaman Requirements file, and reports which files are unknown.

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.ShamanApi();
let ShamanRequirementsRequest = new flamencoManager.ShamanRequirementsRequest(); // ShamanRequirementsRequest | Set of files to check
apiInstance.shamanCheckoutRequirements(ShamanRequirementsRequest).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ShamanRequirementsRequest** | [**ShamanRequirementsRequest**](ShamanRequirementsRequest.md)| Set of files to check | 

### Return type

[**ShamanRequirementsResponse**](ShamanRequirementsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## shamanFileStore

> shamanFileStore(checksum, filesize, body, opts)

Store a new file on the Shaman server. Note that the Shaman server can forcibly close the HTTP connection when another client finishes uploading the exact same file, to prevent double uploads. The file&#39;s contents should be sent in the request body. 

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.ShamanApi();
let checksum = "checksum_example"; // String | SHA256 checksum of the file.
let filesize = 56; // Number | Size of the file in bytes.
let body = "/path/to/file"; // File | Contents of the file
let opts = {
  'X_Shaman_Can_Defer_Upload': true, // Boolean | The client indicates that it can defer uploading this file. The \"208\" response will not only be returned when the file is already fully known to the Shaman server, but also when someone else is currently uploading this file. 
  'X_Shaman_Original_Filename': "X_Shaman_Original_Filename_example" // String | The original filename. If sent along with the request, it will be included in the server logs, which can aid in debugging. 
};
apiInstance.shamanFileStore(checksum, filesize, body, opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **checksum** | **String**| SHA256 checksum of the file. | 
 **filesize** | **Number**| Size of the file in bytes. | 
 **body** | **File**| Contents of the file | 
 **X_Shaman_Can_Defer_Upload** | **Boolean**| The client indicates that it can defer uploading this file. The \&quot;208\&quot; response will not only be returned when the file is already fully known to the Shaman server, but also when someone else is currently uploading this file.  | [optional] 
 **X_Shaman_Original_Filename** | **String**| The original filename. If sent along with the request, it will be included in the server logs, which can aid in debugging.  | [optional] 

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/octet-stream
- **Accept**: application/json


## shamanFileStoreCheck

> ShamanSingleFileStatus shamanFileStoreCheck(checksum, filesize)

Check the status of a file on the Shaman server. 

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.ShamanApi();
let checksum = "checksum_example"; // String | SHA256 checksum of the file.
let filesize = 56; // Number | Size of the file in bytes.
apiInstance.shamanFileStoreCheck(checksum, filesize).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **checksum** | **String**| SHA256 checksum of the file. | 
 **filesize** | **Number**| Size of the file in bytes. | 

### Return type

[**ShamanSingleFileStatus**](ShamanSingleFileStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

