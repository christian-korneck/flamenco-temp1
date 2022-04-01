# flamencoManager.AvailableJobSetting

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**key** | **String** | Identifier for the setting, must be unique within the job type. | 
**type** | [**AvailableJobSettingType**](AvailableJobSettingType.md) |  | 
**subtype** | [**AvailableJobSettingSubtype**](AvailableJobSettingSubtype.md) |  | [optional] 
**choices** | **[String]** | When given, limit the valid values to these choices. Only usable with string type. | [optional] 
**propargs** | **Object** | Any extra arguments to the bpy.props.SomeProperty() call used to create this property. | [optional] 
**description** | **Object** | The description/tooltip shown in the user interface. | [optional] 
**_default** | **Object** | The default value shown to the user when determining this setting. | [optional] 
**_eval** | **String** | Python expression to be evaluated in order to determine the default value for this setting. | [optional] 
**visible** | **Boolean** | Whether to show this setting in the UI of a job submitter (like a Blender add-on). Set to &#x60;false&#x60; when it is an internal setting that shouldn&#39;t be shown to end users.  | [optional] [default to true]
**required** | **Boolean** | Whether to immediately reject a job definition, of this type, without this particular setting.  | [optional] [default to false]
**editable** | **Boolean** | Whether to allow editing this setting after the job has been submitted. Would imply deleting all existing tasks for this job, and recompiling it.  | [optional] [default to false]


