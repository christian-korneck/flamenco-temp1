// Package api_impl implements the OpenAPI API from pkg/api/flamenco-manager.yaml.
package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/task_state_machine"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/shaman"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Flamenco struct {
	jobCompiler  JobCompiler
	persist      PersistenceService
	broadcaster  ChangeBroadcaster
	logStorage   LogStorage
	config       ConfigService
	stateMachine TaskStateMachine
	shaman       Shaman
}

var _ api.ServerInterface = (*Flamenco)(nil)

// Generate mock implementations of these interfaces.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/api_impl_mock.gen.go -package mocks git.blender.org/flamenco/internal/manager/api_impl PersistenceService,ChangeBroadcaster,JobCompiler,LogStorage,ConfigService,TaskStateMachine,Shaman

type PersistenceService interface {
	StoreAuthoredJob(ctx context.Context, authoredJob job_compilers.AuthoredJob) error
	FetchJob(ctx context.Context, jobID string) (*persistence.Job, error)
	// FetchTask fetches the given task and the accompanying job.
	FetchTask(ctx context.Context, taskID string) (*persistence.Task, error)
	SaveTask(ctx context.Context, task *persistence.Task) error
	SaveTaskActivity(ctx context.Context, t *persistence.Task) error
	FetchTasksOfWorkerInStatus(context.Context, *persistence.Worker, api.TaskStatus) ([]*persistence.Task, error)

	CreateWorker(ctx context.Context, w *persistence.Worker) error
	FetchWorker(ctx context.Context, uuid string) (*persistence.Worker, error)
	SaveWorker(ctx context.Context, w *persistence.Worker) error
	SaveWorkerStatus(ctx context.Context, w *persistence.Worker) error

	// ScheduleTask finds a task to execute by the given worker, and assigns it to that worker.
	// If no task is available, (nil, nil) is returned, as this is not an error situation.
	ScheduleTask(ctx context.Context, w *persistence.Worker) (*persistence.Task, error)

	// Database queries.
	QueryJobs(ctx context.Context, query api.JobsQuery) ([]*persistence.Job, error)
}

var _ PersistenceService = (*persistence.DB)(nil)

type TaskStateMachine interface {
	// TaskStatusChange gives a Task a new status, and handles the resulting status changes on the job.
	TaskStatusChange(ctx context.Context, task *persistence.Task, newStatus api.TaskStatus) error

	// JobStatusChange gives a Job a new status, and handles the resulting status changes on its tasks.
	JobStatusChange(ctx context.Context, job *persistence.Job, newJobStatus api.JobStatus) error
}

// TaskStateMachine should be a subset of task_state_machine.StateMachine.
var _ TaskStateMachine = (*task_state_machine.StateMachine)(nil)

type ChangeBroadcaster interface {
	// BroadcastNewJob sends a 'new job' notification to all SocketIO clients.
	BroadcastNewJob(jobUpdate api.JobUpdate)
}

// ChangeBroadcaster should be a subset of webupdates.BiDirComms.
var _ ChangeBroadcaster = (*webupdates.BiDirComms)(nil)

type JobCompiler interface {
	ListJobTypes() api.AvailableJobTypes
	GetJobType(typeName string) (api.AvailableJobType, error)
	Compile(ctx context.Context, job api.SubmittedJob) (*job_compilers.AuthoredJob, error)
}

// LogStorage handles incoming task logs.
type LogStorage interface {
	Write(logger zerolog.Logger, jobID, taskID string, logText string) error
	RotateFile(logger zerolog.Logger, jobID, taskID string)
}

type ConfigService interface {
	VariableReplacer

	// EffectiveStoragePath returns the job storage path used by Flamenco. It's
	// basically the configured storage path, but can be influenced by other
	// options (like Shaman).
	EffectiveStoragePath() string
}

type Shaman interface {
	// IsEnabled returns whether this Shaman service is enabled or not.
	IsEnabled() bool

	// Checkout creates a directory, and symlinks the required files into it. The
	// files must all have been uploaded to Shaman before calling this.
	// Returns the final checkout directory, as it may be modified to ensure uniqueness.
	Checkout(ctx context.Context, checkout api.ShamanCheckout) (string, error)

	// Requirements checks a Shaman Requirements file, and returns the subset
	// containing the unknown files.
	Requirements(ctx context.Context, requirements api.ShamanRequirementsRequest) (api.ShamanRequirementsResponse, error)

	// Check the status of a file on the Shaman server.
	FileStoreCheck(ctx context.Context, checksum string, filesize int64) api.ShamanFileStatus

	// Store a new file on the Shaman server. Note that the Shaman server can
	// return early when another client finishes uploading the exact same file, to
	// prevent double uploads.
	FileStore(ctx context.Context, file io.ReadCloser, checksum string, filesize int64, canDefer bool, originalFilename string) error
}

var _ Shaman = (*shaman.Server)(nil)

// NewFlamenco creates a new Flamenco service.
func NewFlamenco(
	jc JobCompiler,
	jps PersistenceService,
	b ChangeBroadcaster,
	ls LogStorage,
	cs ConfigService,
	sm TaskStateMachine,
	sha Shaman,
) *Flamenco {
	return &Flamenco{
		jobCompiler:  jc,
		persist:      jps,
		broadcaster:  b,
		logStorage:   ls,
		config:       cs,
		stateMachine: sm,
		shaman:       sha,
	}
}

// sendAPIError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendAPIError(e echo.Context, code int, message string, args ...interface{}) error {
	if len(args) > 0 {
		// Only interpret 'message' as format string if there are actually format parameters.
		message = fmt.Sprintf(message, args)
	}

	apiErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	return e.JSON(code, apiErr)
}

// sendAPIErrorDBBusy sends a HTTP 503 Service Unavailable, with a hopefully
// reasonable "retry after" header.
func sendAPIErrorDBBusy(e echo.Context, message string, args ...interface{}) error {
	if len(args) > 0 {
		// Only interpret 'message' as format string if there are actually format parameters.
		message = fmt.Sprintf(message, args)
	}

	code := http.StatusServiceUnavailable
	apiErr := api.Error{
		Code:    int32(code),
		Message: message,
	}

	retryAfter := 1 * time.Second
	seconds := int64(retryAfter.Seconds())
	e.Response().Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
	return e.JSON(code, apiErr)
}
