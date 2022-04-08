// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package api

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	Worker_authScopes = "worker_auth.Scopes"
)

// Defines values for AvailableJobSettingSubtype.
const (
	AvailableJobSettingSubtypeDirPath AvailableJobSettingSubtype = "dir_path"

	AvailableJobSettingSubtypeFileName AvailableJobSettingSubtype = "file_name"

	AvailableJobSettingSubtypeFilePath AvailableJobSettingSubtype = "file_path"

	AvailableJobSettingSubtypeHashedFilePath AvailableJobSettingSubtype = "hashed_file_path"
)

// Defines values for AvailableJobSettingType.
const (
	AvailableJobSettingTypeBool AvailableJobSettingType = "bool"

	AvailableJobSettingTypeFloat AvailableJobSettingType = "float"

	AvailableJobSettingTypeInt32 AvailableJobSettingType = "int32"

	AvailableJobSettingTypeString AvailableJobSettingType = "string"
)

// Defines values for JobStatus.
const (
	JobStatusActive JobStatus = "active"

	JobStatusArchived JobStatus = "archived"

	JobStatusArchiving JobStatus = "archiving"

	JobStatusCancelRequested JobStatus = "cancel-requested"

	JobStatusCanceled JobStatus = "canceled"

	JobStatusCompleted JobStatus = "completed"

	JobStatusConstructionFailed JobStatus = "construction-failed"

	JobStatusFailRequested JobStatus = "fail-requested"

	JobStatusFailed JobStatus = "failed"

	JobStatusPaused JobStatus = "paused"

	JobStatusQueued JobStatus = "queued"

	JobStatusRequeued JobStatus = "requeued"

	JobStatusUnderConstruction JobStatus = "under-construction"

	JobStatusWaitingForFiles JobStatus = "waiting-for-files"
)

// Defines values for ShamanFileStatus.
const (
	ShamanFileStatusStored ShamanFileStatus = "stored"

	ShamanFileStatusUnknown ShamanFileStatus = "unknown"

	ShamanFileStatusUploading ShamanFileStatus = "uploading"
)

// Defines values for TaskStatus.
const (
	TaskStatusActive TaskStatus = "active"

	TaskStatusCancelRequested TaskStatus = "cancel-requested"

	TaskStatusCanceled TaskStatus = "canceled"

	TaskStatusCompleted TaskStatus = "completed"

	TaskStatusFailed TaskStatus = "failed"

	TaskStatusPaused TaskStatus = "paused"

	TaskStatusQueued TaskStatus = "queued"

	TaskStatusSoftFailed TaskStatus = "soft-failed"
)

// Defines values for WorkerStatus.
const (
	WorkerStatusAsleep WorkerStatus = "asleep"

	WorkerStatusAwake WorkerStatus = "awake"

	WorkerStatusError WorkerStatus = "error"

	WorkerStatusOffline WorkerStatus = "offline"

	WorkerStatusShutdown WorkerStatus = "shutdown"

	WorkerStatusStarting WorkerStatus = "starting"

	WorkerStatusTesting WorkerStatus = "testing"
)

// AssignedTask is a task as it is received by the Worker.
type AssignedTask struct {
	Commands    []Command  `json:"commands"`
	Job         string     `json:"job"`
	JobPriority int        `json:"job_priority"`
	JobType     string     `json:"job_type"`
	Name        string     `json:"name"`
	Priority    int        `json:"priority"`
	Status      TaskStatus `json:"status"`
	TaskType    string     `json:"task_type"`
	Uuid        string     `json:"uuid"`
}

// Single setting of a Job types.
type AvailableJobSetting struct {
	// When given, limit the valid values to these choices. Only usable with string type.
	Choices *[]string `json:"choices,omitempty"`

	// The default value shown to the user when determining this setting.
	Default *interface{} `json:"default,omitempty"`

	// The description/tooltip shown in the user interface.
	Description *interface{} `json:"description,omitempty"`

	// Whether to allow editing this setting after the job has been submitted. Would imply deleting all existing tasks for this job, and recompiling it.
	Editable *bool `json:"editable,omitempty"`

	// Python expression to be evaluated in order to determine the default value for this setting.
	Eval *string `json:"eval,omitempty"`

	// Identifier for the setting, must be unique within the job type.
	Key string `json:"key"`

	// Any extra arguments to the bpy.props.SomeProperty() call used to create this property.
	Propargs *map[string]interface{} `json:"propargs,omitempty"`

	// Whether to immediately reject a job definition, of this type, without this particular setting.
	Required *bool `json:"required,omitempty"`

	// Sub-type of the job setting. Currently only available for string types. `HASHED_FILE_PATH` is a directory path + `"/######"` appended.
	Subtype *AvailableJobSettingSubtype `json:"subtype,omitempty"`

	// Type of job setting, must be usable as IDProperty type in Blender. No nested structures (arrays, dictionaries) are supported.
	Type AvailableJobSettingType `json:"type"`

	// Whether to show this setting in the UI of a job submitter (like a Blender add-on). Set to `false` when it is an internal setting that shouldn't be shown to end users.
	Visible *bool `json:"visible,omitempty"`
}

// Sub-type of the job setting. Currently only available for string types. `HASHED_FILE_PATH` is a directory path + `"/######"` appended.
type AvailableJobSettingSubtype string

// Type of job setting, must be usable as IDProperty type in Blender. No nested structures (arrays, dictionaries) are supported.
type AvailableJobSettingType string

// Job type supported by this Manager, and its parameters.
type AvailableJobType struct {
	Label    string                `json:"label"`
	Name     string                `json:"name"`
	Settings []AvailableJobSetting `json:"settings"`
}

// List of job types supported by this Manager.
type AvailableJobTypes struct {
	JobTypes []AvailableJobType `json:"job_types"`
}

// Command represents a single command to execute by the Worker.
type Command struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// Generic error response.
type Error struct {
	// HTTP status code of this response. Is included in the payload so that a single object represents all error information.
	// Code 503 is used when the database is busy. The HTTP response will contain a 'Retry-After' HTTP header that indicates after which time the request can be retried. Following the header is not mandatory, and it's up to the client to do something reasonable like exponential backoff.
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// FlamencoVersion defines model for FlamencoVersion.
type FlamencoVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Job defines model for Job.
type Job struct {
	// Embedded struct due to allOf(#/components/schemas/SubmittedJob)
	SubmittedJob `yaml:",inline"`
	// Embedded fields due to inline allOf schema
	// Creation timestamp
	Created time.Time `json:"created"`

	// UUID of the Job
	Id     string    `json:"id"`
	Status JobStatus `json:"status"`

	// Creation timestamp
	Updated time.Time `json:"updated"`
}

// Arbitrary metadata strings. More complex structures can be modeled by using `a.b.c` notation for the key.
type JobMetadata struct {
	AdditionalProperties map[string]string `json:"-"`
}

// JobSettings defines model for JobSettings.
type JobSettings struct {
	AdditionalProperties map[string]interface{} `json:"-"`
}

// JobStatus defines model for JobStatus.
type JobStatus string

// JobUpdate defines model for JobUpdate.
type JobUpdate struct {
	// UUID of the Job
	Id string `json:"id"`

	// Name of the job
	Name           *string    `json:"name,omitempty"`
	PreviousStatus *JobStatus `json:"previous_status,omitempty"`
	Status         JobStatus  `json:"status"`

	// Timestamp of last update
	Updated time.Time `json:"updated"`
}

// JobsQuery defines model for JobsQuery.
type JobsQuery struct {
	Limit *int `json:"limit,omitempty"`

	// Filter by metadata, using `LIKE` notation.
	Metadata *JobsQuery_Metadata `json:"metadata,omitempty"`
	Offset   *int                `json:"offset,omitempty"`
	OrderBy  *[]string           `json:"order_by,omitempty"`

	// Filter by job settings, using `LIKE` notation.
	Settings *JobsQuery_Settings `json:"settings,omitempty"`

	// Return only jobs with a status in this array.
	StatusIn *[]JobStatus `json:"status_in,omitempty"`
}

// Filter by metadata, using `LIKE` notation.
type JobsQuery_Metadata struct {
	AdditionalProperties map[string]string `json:"-"`
}

// Filter by job settings, using `LIKE` notation.
type JobsQuery_Settings struct {
	AdditionalProperties map[string]interface{} `json:"-"`
}

// JobsQueryResult defines model for JobsQueryResult.
type JobsQueryResult struct {
	Jobs []Job `json:"jobs"`
}

// ManagerConfiguration defines model for ManagerConfiguration.
type ManagerConfiguration struct {
	// Whether the Shaman file transfer API is available.
	ShamanEnabled bool `json:"shamanEnabled"`

	// Directory used for job file storage.
	StorageLocation string `json:"storageLocation"`
}

// RegisteredWorker defines model for RegisteredWorker.
type RegisteredWorker struct {
	Address            string       `json:"address"`
	LastActivity       string       `json:"last_activity"`
	Nickname           string       `json:"nickname"`
	Platform           string       `json:"platform"`
	Software           string       `json:"software"`
	Status             WorkerStatus `json:"status"`
	SupportedTaskTypes []string     `json:"supported_task_types"`
	Uuid               string       `json:"uuid"`
}

// SecurityError defines model for SecurityError.
type SecurityError struct {
	Message string `json:"message"`
}

// Set of files with their SHA256 checksum, size in bytes, and desired location in the checkout directory.
type ShamanCheckout struct {
	// Path where the Manager should create this checkout. It is relative to the Shaman checkout path as configured on the Manager. In older versions of the Shaman this was just the "checkout ID", but in this version it can be a path like `project-slug/scene-name/unique-ID`.
	CheckoutPath string           `json:"checkoutPath"`
	Files        []ShamanFileSpec `json:"files"`
}

// The result of a Shaman checkout.
type ShamanCheckoutResult struct {
	// Path where the Manager created this checkout. This can be different than what was requested, as the Manager will ensure a unique directory. The path is relative to the Shaman checkout path as configured on the Manager.
	CheckoutPath string `json:"checkoutPath"`
}

// Specification of a file in the Shaman storage.
type ShamanFileSpec struct {
	// Location of the file in the checkout
	Path string `json:"path"`

	// SHA256 checksum of the file
	Sha string `json:"sha"`

	// File size in bytes
	Size int `json:"size"`
}

// Specification of a file, which could be in the Shaman storage, or not, depending on its status.
type ShamanFileSpecWithStatus struct {
	// Location of the file in the checkout
	Path string `json:"path"`

	// SHA256 checksum of the file
	Sha string `json:"sha"`

	// File size in bytes
	Size   int              `json:"size"`
	Status ShamanFileStatus `json:"status"`
}

// ShamanFileStatus defines model for ShamanFileStatus.
type ShamanFileStatus string

// Set of files with their SHA256 checksum and size in bytes.
type ShamanRequirementsRequest struct {
	Files []ShamanFileSpec `json:"files"`
}

// The files from a requirements request, with their status on the Shaman server. Files that are known to Shaman are excluded from the response.
type ShamanRequirementsResponse struct {
	Files []ShamanFileSpecWithStatus `json:"files"`
}

// Status of a file in the Shaman storage.
type ShamanSingleFileStatus struct {
	Status ShamanFileStatus `json:"status"`
}

// Job definition submitted to Flamenco.
type SubmittedJob struct {
	// Arbitrary metadata strings. More complex structures can be modeled by using `a.b.c` notation for the key.
	Metadata *JobMetadata `json:"metadata,omitempty"`
	Name     string       `json:"name"`
	Priority int          `json:"priority"`
	Settings *JobSettings `json:"settings,omitempty"`
	Type     string       `json:"type"`
}

// TaskStatus defines model for TaskStatus.
type TaskStatus string

// TaskUpdate is sent by a Worker to update the status & logs of a task it's executing.
type TaskUpdate struct {
	// One-liner to indicate what's currently happening with the task. Overwrites previously sent activity strings.
	Activity *string `json:"activity,omitempty"`

	// Log lines for this task, will be appended to logs sent earlier.
	Log        *string     `json:"log,omitempty"`
	TaskStatus *TaskStatus `json:"taskStatus,omitempty"`
}

// WorkerRegistration defines model for WorkerRegistration.
type WorkerRegistration struct {
	Nickname           string   `json:"nickname"`
	Platform           string   `json:"platform"`
	Secret             string   `json:"secret"`
	SupportedTaskTypes []string `json:"supported_task_types"`
}

// WorkerSignOn defines model for WorkerSignOn.
type WorkerSignOn struct {
	Nickname           string   `json:"nickname"`
	SoftwareVersion    string   `json:"software_version"`
	SupportedTaskTypes []string `json:"supported_task_types"`
}

// WorkerStateChange defines model for WorkerStateChange.
type WorkerStateChange struct {
	StatusRequested WorkerStatus `json:"status_requested"`
}

// WorkerStateChanged defines model for WorkerStateChanged.
type WorkerStateChanged struct {
	Status WorkerStatus `json:"status"`
}

// WorkerStatus defines model for WorkerStatus.
type WorkerStatus string

// SubmitJobJSONBody defines parameters for SubmitJob.
type SubmitJobJSONBody SubmittedJob

// QueryJobsJSONBody defines parameters for QueryJobs.
type QueryJobsJSONBody JobsQuery

// RegisterWorkerJSONBody defines parameters for RegisterWorker.
type RegisterWorkerJSONBody WorkerRegistration

// SignOnJSONBody defines parameters for SignOn.
type SignOnJSONBody WorkerSignOn

// WorkerStateChangedJSONBody defines parameters for WorkerStateChanged.
type WorkerStateChangedJSONBody WorkerStateChanged

// TaskUpdateJSONBody defines parameters for TaskUpdate.
type TaskUpdateJSONBody TaskUpdate

// ShamanCheckoutJSONBody defines parameters for ShamanCheckout.
type ShamanCheckoutJSONBody ShamanCheckout

// ShamanCheckoutRequirementsJSONBody defines parameters for ShamanCheckoutRequirements.
type ShamanCheckoutRequirementsJSONBody ShamanRequirementsRequest

// ShamanFileStoreParams defines parameters for ShamanFileStore.
type ShamanFileStoreParams struct {
	// The client indicates that it can defer uploading this file. The "208" response will not only be returned when the file is already fully known to the Shaman server, but also when someone else is currently uploading this file.
	XShamanCanDeferUpload *bool `json:"X-Shaman-Can-Defer-Upload,omitempty"`

	// The original filename. If sent along with the request, it will be included in the server logs, which can aid in debugging.
	XShamanOriginalFilename *string `json:"X-Shaman-Original-Filename,omitempty"`
}

// SubmitJobJSONRequestBody defines body for SubmitJob for application/json ContentType.
type SubmitJobJSONRequestBody SubmitJobJSONBody

// QueryJobsJSONRequestBody defines body for QueryJobs for application/json ContentType.
type QueryJobsJSONRequestBody QueryJobsJSONBody

// RegisterWorkerJSONRequestBody defines body for RegisterWorker for application/json ContentType.
type RegisterWorkerJSONRequestBody RegisterWorkerJSONBody

// SignOnJSONRequestBody defines body for SignOn for application/json ContentType.
type SignOnJSONRequestBody SignOnJSONBody

// WorkerStateChangedJSONRequestBody defines body for WorkerStateChanged for application/json ContentType.
type WorkerStateChangedJSONRequestBody WorkerStateChangedJSONBody

// TaskUpdateJSONRequestBody defines body for TaskUpdate for application/json ContentType.
type TaskUpdateJSONRequestBody TaskUpdateJSONBody

// ShamanCheckoutJSONRequestBody defines body for ShamanCheckout for application/json ContentType.
type ShamanCheckoutJSONRequestBody ShamanCheckoutJSONBody

// ShamanCheckoutRequirementsJSONRequestBody defines body for ShamanCheckoutRequirements for application/json ContentType.
type ShamanCheckoutRequirementsJSONRequestBody ShamanCheckoutRequirementsJSONBody

// Getter for additional properties for JobMetadata. Returns the specified
// element and whether it was found
func (a JobMetadata) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for JobMetadata
func (a *JobMetadata) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for JobMetadata to handle AdditionalProperties
func (a *JobMetadata) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for JobMetadata to handle AdditionalProperties
func (a JobMetadata) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for JobSettings. Returns the specified
// element and whether it was found
func (a JobSettings) Get(fieldName string) (value interface{}, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for JobSettings
func (a *JobSettings) Set(fieldName string, value interface{}) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]interface{})
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for JobSettings to handle AdditionalProperties
func (a *JobSettings) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]interface{})
		for fieldName, fieldBuf := range object {
			var fieldVal interface{}
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for JobSettings to handle AdditionalProperties
func (a JobSettings) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for JobsQuery_Metadata. Returns the specified
// element and whether it was found
func (a JobsQuery_Metadata) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for JobsQuery_Metadata
func (a *JobsQuery_Metadata) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for JobsQuery_Metadata to handle AdditionalProperties
func (a *JobsQuery_Metadata) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for JobsQuery_Metadata to handle AdditionalProperties
func (a JobsQuery_Metadata) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for JobsQuery_Settings. Returns the specified
// element and whether it was found
func (a JobsQuery_Settings) Get(fieldName string) (value interface{}, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for JobsQuery_Settings
func (a *JobsQuery_Settings) Set(fieldName string, value interface{}) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]interface{})
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for JobsQuery_Settings to handle AdditionalProperties
func (a *JobsQuery_Settings) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]interface{})
		for fieldName, fieldBuf := range object {
			var fieldVal interface{}
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for JobsQuery_Settings to handle AdditionalProperties
func (a JobsQuery_Settings) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}
