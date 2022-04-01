# flamencoManager.WorkerApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**registerWorker**](WorkerApi.md#registerWorker) | **POST** /api/worker/register-worker | Register a new worker
[**scheduleTask**](WorkerApi.md#scheduleTask) | **POST** /api/worker/task | Obtain a new task to execute
[**signOff**](WorkerApi.md#signOff) | **POST** /api/worker/sign-off | Mark the worker as offline
[**signOn**](WorkerApi.md#signOn) | **POST** /api/worker/sign-on | Authenticate &amp; sign in the worker.
[**taskUpdate**](WorkerApi.md#taskUpdate) | **POST** /api/worker/task/{task_id} | Update the task, typically to indicate progress, completion, or failure.
[**workerState**](WorkerApi.md#workerState) | **GET** /api/worker/state | 
[**workerStateChanged**](WorkerApi.md#workerStateChanged) | **POST** /api/worker/state-changed | Worker changed state. This could be as acknowledgement of a Manager-requested state change, or in response to worker-local signals.



## registerWorker

> RegisteredWorker registerWorker(workerRegistration)

Register a new worker

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.WorkerApi();
let workerRegistration = new flamencoManager.WorkerRegistration(); // WorkerRegistration | Worker to register
apiInstance.registerWorker(workerRegistration).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workerRegistration** | [**WorkerRegistration**](WorkerRegistration.md)| Worker to register | 

### Return type

[**RegisteredWorker**](RegisteredWorker.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## scheduleTask

> AssignedTask scheduleTask()

Obtain a new task to execute

### Example

```javascript
import flamencoManager from 'flamenco-manager';
let defaultClient = flamencoManager.ApiClient.instance;
// Configure HTTP basic authorization: worker_auth
let worker_auth = defaultClient.authentications['worker_auth'];
worker_auth.username = 'YOUR USERNAME';
worker_auth.password = 'YOUR PASSWORD';

let apiInstance = new flamencoManager.WorkerApi();
apiInstance.scheduleTask().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**AssignedTask**](AssignedTask.md)

### Authorization

[worker_auth](../README.md#worker_auth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## signOff

> signOff()

Mark the worker as offline

### Example

```javascript
import flamencoManager from 'flamenco-manager';
let defaultClient = flamencoManager.ApiClient.instance;
// Configure HTTP basic authorization: worker_auth
let worker_auth = defaultClient.authentications['worker_auth'];
worker_auth.username = 'YOUR USERNAME';
worker_auth.password = 'YOUR PASSWORD';

let apiInstance = new flamencoManager.WorkerApi();
apiInstance.signOff().then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

null (empty response body)

### Authorization

[worker_auth](../README.md#worker_auth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## signOn

> WorkerStateChange signOn(workerSignOn)

Authenticate &amp; sign in the worker.

### Example

```javascript
import flamencoManager from 'flamenco-manager';
let defaultClient = flamencoManager.ApiClient.instance;
// Configure HTTP basic authorization: worker_auth
let worker_auth = defaultClient.authentications['worker_auth'];
worker_auth.username = 'YOUR USERNAME';
worker_auth.password = 'YOUR PASSWORD';

let apiInstance = new flamencoManager.WorkerApi();
let workerSignOn = new flamencoManager.WorkerSignOn(); // WorkerSignOn | Worker metadata
apiInstance.signOn(workerSignOn).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workerSignOn** | [**WorkerSignOn**](WorkerSignOn.md)| Worker metadata | 

### Return type

[**WorkerStateChange**](WorkerStateChange.md)

### Authorization

[worker_auth](../README.md#worker_auth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## taskUpdate

> taskUpdate(taskId, taskUpdate)

Update the task, typically to indicate progress, completion, or failure.

### Example

```javascript
import flamencoManager from 'flamenco-manager';
let defaultClient = flamencoManager.ApiClient.instance;
// Configure HTTP basic authorization: worker_auth
let worker_auth = defaultClient.authentications['worker_auth'];
worker_auth.username = 'YOUR USERNAME';
worker_auth.password = 'YOUR PASSWORD';

let apiInstance = new flamencoManager.WorkerApi();
let taskId = "taskId_example"; // String | 
let taskUpdate = new flamencoManager.TaskUpdate(); // TaskUpdate | Task update information
apiInstance.taskUpdate(taskId, taskUpdate).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **taskId** | **String**|  | 
 **taskUpdate** | [**TaskUpdate**](TaskUpdate.md)| Task update information | 

### Return type

null (empty response body)

### Authorization

[worker_auth](../README.md#worker_auth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## workerState

> WorkerStateChange workerState()



### Example

```javascript
import flamencoManager from 'flamenco-manager';
let defaultClient = flamencoManager.ApiClient.instance;
// Configure HTTP basic authorization: worker_auth
let worker_auth = defaultClient.authentications['worker_auth'];
worker_auth.username = 'YOUR USERNAME';
worker_auth.password = 'YOUR PASSWORD';

let apiInstance = new flamencoManager.WorkerApi();
apiInstance.workerState().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**WorkerStateChange**](WorkerStateChange.md)

### Authorization

[worker_auth](../README.md#worker_auth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## workerStateChanged

> workerStateChanged(workerStateChanged)

Worker changed state. This could be as acknowledgement of a Manager-requested state change, or in response to worker-local signals.

### Example

```javascript
import flamencoManager from 'flamenco-manager';
let defaultClient = flamencoManager.ApiClient.instance;
// Configure HTTP basic authorization: worker_auth
let worker_auth = defaultClient.authentications['worker_auth'];
worker_auth.username = 'YOUR USERNAME';
worker_auth.password = 'YOUR PASSWORD';

let apiInstance = new flamencoManager.WorkerApi();
let workerStateChanged = new flamencoManager.WorkerStateChanged(); // WorkerStateChanged | New worker state
apiInstance.workerStateChanged(workerStateChanged).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workerStateChanged** | [**WorkerStateChanged**](WorkerStateChanged.md)| New worker state | 

### Return type

null (empty response body)

### Authorization

[worker_auth](../README.md#worker_auth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

