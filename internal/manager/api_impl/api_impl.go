// Package api_impl implements the OpenAPI API from pkg/api/flamenco-manager.yaml.
package api_impl

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/task_state_machine"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type Flamenco struct {
	jobCompiler  JobCompiler
	persist      PersistenceService
	logStorage   LogStorage
	config       ConfigService
	stateMachine TaskStateMachine
}

var _ api.ServerInterface = (*Flamenco)(nil)

// Generate mock implementations of these interfaces.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/api_impl_mock.gen.go -package mocks gitlab.com/blender/flamenco-ng-poc/internal/manager/api_impl PersistenceService,JobCompiler,LogStorage,ConfigService,TaskStateMachine

type PersistenceService interface {
	StoreAuthoredJob(ctx context.Context, authoredJob job_compilers.AuthoredJob) error
	FetchJob(ctx context.Context, jobID string) (*persistence.Job, error)
	// FetchTask fetches the given task and the accompanying job.
	FetchTask(ctx context.Context, taskID string) (*persistence.Task, error)
	SaveTask(ctx context.Context, task *persistence.Task) error
	SaveTaskActivity(ctx context.Context, t *persistence.Task) error

	CreateWorker(ctx context.Context, w *persistence.Worker) error
	FetchWorker(ctx context.Context, uuid string) (*persistence.Worker, error)
	SaveWorker(ctx context.Context, w *persistence.Worker) error
	SaveWorkerStatus(ctx context.Context, w *persistence.Worker) error

	// ScheduleTask finds a task to execute by the given worker, and assigns it to that worker.
	// If no task is available, (nil, nil) is returned, as this is not an error situation.
	ScheduleTask(w *persistence.Worker) (*persistence.Task, error)
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

type JobCompiler interface {
	ListJobTypes() api.AvailableJobTypes
	Compile(ctx context.Context, job api.SubmittedJob) (*job_compilers.AuthoredJob, error)
}

// LogStorage handles incoming task logs.
type LogStorage interface {
	Write(logger zerolog.Logger, jobID, taskID string, logText string) error
	RotateFile(logger zerolog.Logger, jobID, taskID string)
}

type ConfigService interface {
	VariableReplacer
}

// NewFlamenco creates a new Flamenco service.
func NewFlamenco(
	jc JobCompiler,
	jps PersistenceService,
	ls LogStorage,
	cs ConfigService,
	sm TaskStateMachine,
) *Flamenco {
	return &Flamenco{
		jobCompiler:  jc,
		persist:      jps,
		logStorage:   ls,
		config:       cs,
		stateMachine: sm,
	}
}

// sendPetstoreError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendAPIError(e echo.Context, code int, message string, args ...interface{}) error {
	if len(args) > 0 {
		// Only interpret 'message' as format string if there are actually format parameters.
		message = fmt.Sprintf(message, args)
	}

	petErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	return e.JSON(code, petErr)
}
