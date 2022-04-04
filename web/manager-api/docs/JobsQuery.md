# flamencoManager.JobsQuery

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**offset** | **Number** |  | [optional] 
**limit** | **Number** |  | [optional] 
**order_by** | **[String]** |  | [optional] 
**status_in** | [**[JobStatus]**](JobStatus.md) | Return only jobs with a status in this array. | [optional] 
**metadata** | **{String: String}** | Filter by metadata, using &#x60;LIKE&#x60; notation. | [optional] 
**settings** | **{String: Object}** | Filter by job settings, using &#x60;LIKE&#x60; notation. | [optional] 


