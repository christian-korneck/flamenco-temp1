# Job


## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **str** |  | 
**type** | **str** |  | 
**submitter_platform** | **str** | Operating system of the submitter. This is used to recognise two-way variables. This should be a lower-case version of the platform, like \&quot;linux\&quot;, \&quot;windows\&quot;, \&quot;darwin\&quot;, \&quot;openbsd\&quot;, etc. Should be ompatible with Go&#39;s &#x60;runtime.GOOS&#x60;; run &#x60;go tool dist list&#x60; to get a list of possible platforms. As a special case, the platform \&quot;manager\&quot; can be given, which will be interpreted as \&quot;the Manager&#39;s platform\&quot;. This is mostly to make test/debug scripts easier, as they can use a static document on all platforms.  | 
**id** | **str** | UUID of the Job | 
**created** | **datetime** | Creation timestamp | 
**updated** | **datetime** | Timestamp of last update. | 
**status** | [**JobStatus**](JobStatus.md) |  | 
**activity** | **str** | Description of the last activity on this job. | 
**priority** | **int** |  | defaults to 50
**settings** | [**JobSettings**](JobSettings.md) |  | [optional] 
**metadata** | [**JobMetadata**](JobMetadata.md) |  | [optional] 
**any string name** | **bool, date, datetime, dict, float, int, list, str, none_type** | any string name can be used but the value must be the correct type | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


