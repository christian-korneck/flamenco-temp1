# Job


## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **str** |  | 
**type** | **str** |  | 
**id** | **str** | UUID of the Job | 
**created** | **datetime** | Creation timestamp | 
**updated** | **datetime** | Creation timestamp | 
**status** | [**JobStatus**](JobStatus.md) |  | 
**activity** | **str** | Description of the last activity on this job. | 
**priority** | **int** |  | defaults to 50
**settings** | [**JobSettings**](JobSettings.md) |  | [optional] 
**metadata** | [**JobMetadata**](JobMetadata.md) |  | [optional] 
**any string name** | **bool, date, datetime, dict, float, int, list, str, none_type** | any string name can be used but the value must be the correct type | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


