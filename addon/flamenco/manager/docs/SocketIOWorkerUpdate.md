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
**status_change** | [**WorkerStatusChangeRequest**](WorkerStatusChangeRequest.md) |  | [optional] 
**any string name** | **bool, date, datetime, dict, float, int, list, str, none_type** | any string name can be used but the value must be the correct type | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


