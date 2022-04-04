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

> RegisteredWorker registerWorker(WorkerRegistration)

Register a new worker

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.WorkerApi();
let WorkerRegistration = new flamencoManager.WorkerRegistration(); // WorkerRegistration | Worker to register
apiInstance.registerWorker(WorkerRegistration).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **WorkerRegistration** | [**WorkerRegistration**](WorkerRegistration.md)| Worker to register | 

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

> WorkerStateChange signOn(WorkerSignOn)

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
let WorkerSignOn = new flamencoManager.WorkerSignOn(); // WorkerSignOn | Worker metadata
apiInstance.signOn(WorkerSignOn).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **WorkerSignOn** | [**WorkerSignOn**](WorkerSignOn.md)| Worker metadata | 

### Return type

[**WorkerStateChange**](WorkerStateChange.md)

### Authorization

[worker_auth](../README.md#worker_auth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## taskUpdate

> taskUpdate(task_id, TaskUpdate)

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
let task_id = "task_id_example"; // String | 
let TaskUpdate = new flamencoManager.TaskUpdate(); // TaskUpdate | Task update information
apiInstance.taskUpdate(task_id, TaskUpdate).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **task_id** | **String**|  | 
 **TaskUpdate** | [**TaskUpdate**](TaskUpdate.md)| Task update information | 

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

> workerStateChanged(WorkerStateChanged)

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
let WorkerStateChanged = new flamencoManager.WorkerStateChanged(); // WorkerStateChanged | New worker state
apiInstance.workerStateChanged(WorkerStateChanged).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **WorkerStateChanged** | [**WorkerStateChanged**](WorkerStateChanged.md)| New worker state | 

### Return type

null (empty response body)

### Authorization

[worker_auth](../README.md#worker_auth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

