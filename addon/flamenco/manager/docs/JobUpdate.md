# JobUpdate

Subset of a Job, sent over SocketIO when a job changes. For new jobs, `previous_status` will be excluded. 

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **str** | UUID of the Job | 
**updated** | **datetime** | Timestamp of last update | 
**status** | [**JobStatus**](JobStatus.md) |  | 
**type** | **str** |  | 
**priority** | **int** |  | defaults to 50
**name** | **str** | Name of the job | [optional] 
**previous_status** | [**JobStatus**](JobStatus.md) |  | [optional] 
**any string name** | **bool, date, datetime, dict, float, int, list, str, none_type** | any string name can be used but the value must be the correct type | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


