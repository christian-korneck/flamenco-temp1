# flamencoManager.JobsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**fetchJob**](JobsApi.md#fetchJob) | **GET** /api/jobs/{job_id} | Fetch info about the job.
[**getJobTypes**](JobsApi.md#getJobTypes) | **GET** /api/jobs/types | Get list of job types and their parameters.
[**queryJobs**](JobsApi.md#queryJobs) | **POST** /api/jobs/query | Fetch list of jobs.
[**submitJob**](JobsApi.md#submitJob) | **POST** /api/jobs | Submit a new job for Flamenco Manager to execute.



## fetchJob

> Job fetchJob(job_id)

Fetch info about the job.

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.JobsApi();
let job_id = "job_id_example"; // String | 
apiInstance.fetchJob(job_id).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **job_id** | **String**|  | 

### Return type

[**Job**](Job.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## getJobTypes

> AvailableJobTypes getJobTypes()

Get list of job types and their parameters.

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.JobsApi();
apiInstance.getJobTypes().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

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


## queryJobs

> JobsQueryResult queryJobs(JobsQuery)

Fetch list of jobs.

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.JobsApi();
let JobsQuery = new flamencoManager.JobsQuery(); // JobsQuery | Specification of which jobs to get.
apiInstance.queryJobs(JobsQuery).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **JobsQuery** | [**JobsQuery**](JobsQuery.md)| Specification of which jobs to get. | 

### Return type

[**JobsQueryResult**](JobsQueryResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json


## submitJob

> Job submitJob(SubmittedJob)

Submit a new job for Flamenco Manager to execute.

### Example

```javascript
import flamencoManager from 'flamenco-manager';

let apiInstance = new flamencoManager.JobsApi();
let SubmittedJob = new flamencoManager.SubmittedJob(); // SubmittedJob | Job to submit
apiInstance.submitJob(SubmittedJob).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **SubmittedJob** | [**SubmittedJob**](SubmittedJob.md)| Job to submit | 

### Return type

[**Job**](Job.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

