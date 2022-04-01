# flamencoManager.Job

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **String** |  | 
**type** | **String** |  | 
**priority** | **Number** |  | [default to 50]
**settings** | **{String: Object}** |  | [optional] 
**metadata** | **{String: String}** | Arbitrary metadata strings. More complex structures can be modeled by using &#x60;a.b.c&#x60; notation for the key. | [optional] 
**id** | **String** | UUID of the Job | 
**created** | **Date** | Creation timestamp | 
**updated** | **Date** | Creation timestamp | 
**status** | [**JobStatus**](JobStatus.md) |  | 


