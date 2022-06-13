package task_state_machine

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/task_logs"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/pkg/api"
	"github.com/rs/zerolog"
)

// Generate mock implementations of these interfaces.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/interfaces_mock.gen.go -package mocks git.blender.org/flamenco/internal/manager/task_state_machine PersistenceService,ChangeBroadcaster,LogStorage

type PersistenceService interface {
	SaveTask(ctx context.Context, task *persistence.Task) error
	SaveTaskActivity(ctx context.Context, t *persistence.Task) error
	SaveJobStatus(ctx context.Context, j *persistence.Job) error

	JobHasTasksInStatus(ctx context.Context, job *persistence.Job, taskStatus api.TaskStatus) (bool, error)
	CountTasksOfJobInStatus(ctx context.Context, job *persistence.Job, taskStatuses ...api.TaskStatus) (numInStatus, numTotal int, err error)

	// UpdateJobsTaskStatuses updates the status & activity of the tasks of `job`.
	UpdateJobsTaskStatuses(ctx context.Context, job *persistence.Job,
		taskStatus api.TaskStatus, activity string) error

	// UpdateJobsTaskStatusesConditional updates the status & activity of the tasks of `job`,
	// limited to those tasks with status in `statusesToUpdate`.
	UpdateJobsTaskStatusesConditional(ctx context.Context, job *persistence.Job,
		statusesToUpdate []api.TaskStatus, taskStatus api.TaskStatus, activity string) error

	FetchJobsInStatus(ctx context.Context, jobStatuses ...api.JobStatus) ([]*persistence.Job, error)
	FetchTasksOfWorkerInStatus(context.Context, *persistence.Worker, api.TaskStatus) ([]*persistence.Task, error)
}

// PersistenceService should be a subset of persistence.DB
var _ PersistenceService = (*persistence.DB)(nil)

type ChangeBroadcaster interface {
	// BroadcastJobUpdate sends the job update to SocketIO clients.
	BroadcastJobUpdate(jobUpdate api.SocketIOJobUpdate)

	// BroadcastTaskUpdate sends the task update to SocketIO clients.
	BroadcastTaskUpdate(jobUpdate api.SocketIOTaskUpdate)
}

// ChangeBroadcaster should be a subset of webupdates.BiDirComms
var _ ChangeBroadcaster = (*webupdates.BiDirComms)(nil)

// LogStorage writes to task logs.
type LogStorage interface {
	WriteTimestamped(logger zerolog.Logger, jobID, taskID string, logText string) error
}

var _ LogStorage = (*task_logs.Storage)(nil)
