package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

// This file contains the interfaces used by the package. They are intended to
// allow swapping actual services with mocked versions for unit tests.

import (
	"context"
	"io"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/rs/zerolog"

	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/manager/last_rendered"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/task_state_machine"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/shaman"
)

// Generate mock implementations of these interfaces.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/api_impl_mock.gen.go -package mocks git.blender.org/flamenco/internal/manager/api_impl PersistenceService,ChangeBroadcaster,JobCompiler,LogStorage,ConfigService,TaskStateMachine,Shaman,LastRendered,LocalStorage

type PersistenceService interface {
	StoreAuthoredJob(ctx context.Context, authoredJob job_compilers.AuthoredJob) error
	// FetchJob fetches a single job, without fetching its tasks.
	FetchJob(ctx context.Context, jobID string) (*persistence.Job, error)
	// FetchTask fetches the given task and the accompanying job.
	FetchTask(ctx context.Context, taskID string) (*persistence.Task, error)
	FetchTaskFailureList(context.Context, *persistence.Task) ([]*persistence.Worker, error)
	SaveTask(ctx context.Context, task *persistence.Task) error
	SaveTaskActivity(ctx context.Context, t *persistence.Task) error
	// TaskTouchedByWorker marks the task as 'touched' by a worker. This is used for timeout detection.
	TaskTouchedByWorker(context.Context, *persistence.Task) error

	CreateWorker(ctx context.Context, w *persistence.Worker) error
	FetchWorker(ctx context.Context, uuid string) (*persistence.Worker, error)
	FetchWorkers(ctx context.Context) ([]*persistence.Worker, error)
	SaveWorker(ctx context.Context, w *persistence.Worker) error
	SaveWorkerStatus(ctx context.Context, w *persistence.Worker) error
	WorkerSeen(ctx context.Context, w *persistence.Worker) error

	// ScheduleTask finds a task to execute by the given worker, and assigns it to that worker.
	// If no task is available, (nil, nil) is returned, as this is not an error situation.
	ScheduleTask(ctx context.Context, w *persistence.Worker) (*persistence.Task, error)
	AddWorkerToTaskFailedList(context.Context, *persistence.Task, *persistence.Worker) (numFailed int, err error)
	// ClearFailureListOfTask clears the list of workers that failed this task.
	ClearFailureListOfTask(context.Context, *persistence.Task) error
	// ClearFailureListOfJob en-mass, for all tasks of this job, clears the list of workers that failed those tasks.
	ClearFailureListOfJob(context.Context, *persistence.Job) error

	// AddWorkerToJobBlocklist prevents this Worker of getting any task, of this type, on this job, from the task scheduler.
	AddWorkerToJobBlocklist(ctx context.Context, job *persistence.Job, worker *persistence.Worker, taskType string) error
	FetchJobBlocklist(ctx context.Context, jobUUID string) ([]persistence.JobBlock, error)
	RemoveFromJobBlocklist(ctx context.Context, jobUUID, workerUUID, taskType string) error
	ClearJobBlocklist(ctx context.Context, job *persistence.Job) error

	// WorkersLeftToRun returns a set of worker UUIDs that can run tasks of the given type on the given job.
	WorkersLeftToRun(ctx context.Context, job *persistence.Job, taskType string) (map[string]bool, error)
	// CountTaskFailuresOfWorker returns the number of task failures of this worker, on this particular job and task type.
	CountTaskFailuresOfWorker(ctx context.Context, job *persistence.Job, worker *persistence.Worker, taskType string) (int, error)

	FetchWorkerSleepSchedule(ctx context.Context, workerUUID string) (*persistence.SleepSchedule, error)
	SetWorkerSleepSchedule(ctx context.Context, workerUUID string, schedule persistence.SleepSchedule) error

	// Database queries.
	QueryJobs(ctx context.Context, query api.JobsQuery) ([]*persistence.Job, error)
	QueryJobTaskSummaries(ctx context.Context, jobUUID string) ([]*persistence.Task, error)

	// SetLastRendered sets this job as the one with the most recent rendered image.
	SetLastRendered(ctx context.Context, j *persistence.Job) error
	// GetLastRendered returns the UUID of the job with the most recent rendered image.
	GetLastRenderedJobUUID(ctx context.Context) (string, error)
}

var _ PersistenceService = (*persistence.DB)(nil)

type TaskStateMachine interface {
	// TaskStatusChange gives a Task a new status, and handles the resulting status changes on the job.
	TaskStatusChange(ctx context.Context, task *persistence.Task, newStatus api.TaskStatus) error

	// JobStatusChange gives a Job a new status, and handles the resulting status changes on its tasks.
	JobStatusChange(ctx context.Context, job *persistence.Job, newJobStatus api.JobStatus, reason string) error

	RequeueActiveTasksOfWorker(ctx context.Context, worker *persistence.Worker, reason string) error
	RequeueFailedTasksOfWorkerOfJob(ctx context.Context, worker *persistence.Worker, job *persistence.Job, reason string) error
}

// TaskStateMachine should be a subset of task_state_machine.StateMachine.
var _ TaskStateMachine = (*task_state_machine.StateMachine)(nil)

type ChangeBroadcaster interface {
	// BroadcastNewJob sends a 'new job' notification to all SocketIO clients.
	BroadcastNewJob(jobUpdate api.SocketIOJobUpdate)
	BroadcastLastRenderedImage(update api.SocketIOLastRenderedUpdate)

	// Note that there is no BroadcastNewTask. The 'new job' broadcast is sent
	// after the job's tasks have been created, and thus there is no need for a
	// separate broadcast per task.

	// Note that there is no call to BoardcastTaskLogUpdate. It's the
	// responsibility of `LogStorage.Write` to broadcast the changes to SocketIO
	// clients.

	BroadcastWorkerUpdate(workerUpdate api.SocketIOWorkerUpdate)
	BroadcastNewWorker(workerUpdate api.SocketIOWorkerUpdate)
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
	WriteTimestamped(logger zerolog.Logger, jobID, taskID string, logText string) error
	RotateFile(logger zerolog.Logger, jobID, taskID string)
	Tail(jobID, taskID string) (string, error)
	TaskLogSize(jobID, taskID string) (int64, error)
	Filepath(jobID, taskID string) string
}

// LastRendered processes the "last rendered" images.
type LastRendered interface {
	// QueueImage queues an image for processing. Returns
	// `last_rendered.ErrQueueFull` if there is no more space in the queue for
	// new images.
	QueueImage(payload last_rendered.Payload) error

	// PathForJob returns the base path for this job's last-rendered images.
	PathForJob(jobUUID string) string

	// ThumbSpecs returns the thumbnail specifications.
	ThumbSpecs() []last_rendered.Thumbspec

	// JobHasImage returns true only if the job actually has a last-rendered image.
	JobHasImage(jobUUID string) bool
}

// LocalStorage handles the storage organisation of local files like the last-rendered images.
type LocalStorage interface {
	// RelPath tries to make the given path relative to the local storage root.
	// Assumes `path` is already an absolute path.
	RelPath(path string) (string, error)
}

type ConfigService interface {
	VariableReplacer

	Get() *config.Conf

	// EffectiveStoragePath returns the job storage path used by Flamenco. It's
	// basically the configured storage path, but can be influenced by other
	// options (like Shaman).
	EffectiveStoragePath() string

	// IsFirstRun returns true if this is likely to be the first run of Flamenco.
	IsFirstRun() (bool, error)

	// ForceFirstRun forces IsFirstRun() to return true. This is used to force the
	// first-time wizard on a configured system.
	ForceFirstRun()

	// Save writes the in-memory configuration to the config file.
	Save() error
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

// TimeService provides functionality from the stdlib `time` module, but in a
// way that allows mocking.
type TimeService interface {
	Now() time.Time
}

var _ TimeService = (clock.Clock)(nil)
