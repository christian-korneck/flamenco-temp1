# SocketIOWorkerUpdate

Subset of a Worker, sent over SocketIO when a worker changes. 

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **str** | UUID of the Worker | 
**nickname** | **str** | Name of the worker | 
**updated** | **datetime** | Timestamp of last update | 
**status** | [**WorkerStatus**](WorkerStatus.md) |  | 
**version** | **str** |  | 
**previous_status** | [**WorkerStatus**](WorkerStatus.md) |  | [optional] 
**status_requested** | [**WorkerStatus**](WorkerStatus.md) |  | [optional] 
**lazy_status_request** | **bool** | Whether the worker is allowed to finish its current task before the status change is enforced. Mandatory when &#x60;status_requested&#x60; is set.  | [optional] 
**any string name** | **bool, date, datetime, dict, float, int, list, str, none_type** | any string name can be used but the value must be the correct type | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


