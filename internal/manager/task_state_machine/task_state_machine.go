package task_state_machine

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

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

// taskFailJobPercentage is the percentage of a job's tasks that need to fail to
// trigger failure of the entire job.
const taskFailJobPercentage = 10 // Integer from 0 to 100.

// StateMachine handles task and job status changes.
type StateMachine struct {
	persist PersistenceService
}

// Generate mock implementations of these interfaces.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/interfaces_mock.gen.go -package mocks gitlab.com/blender/flamenco-ng-poc/internal/manager/task_state_machine PersistenceService

type PersistenceService interface { // Subset of persistence.DB
	SaveTask(ctx context.Context, task *persistence.Task) error
	SaveJobStatus(ctx context.Context, j *persistence.Job) error

	JobHasTasksInStatus(ctx context.Context, job *persistence.Job, taskStatus api.TaskStatus) (bool, error)
	CountTasksOfJobInStatus(ctx context.Context, job *persistence.Job, taskStatus api.TaskStatus) (numInStatus, numTotal int, err error)
}

func NewStateMachine(persist PersistenceService) *StateMachine {
	return &StateMachine{
		persist: persist,
	}
}

// TaskStatusChange updates the task's status to the new one.
// `task` is expected to still have its original status, and have a filled `Job` pointer.
func (sm *StateMachine) TaskStatusChange(ctx context.Context, task *persistence.Task, newTaskStatus api.TaskStatus) error {
	job := task.Job
	if job == nil {
		log.Panic().Str("task", task.UUID).Msg("task without job, cannot handle this")
		return nil // Will not run because of the panic.
	}

	logger := log.With().
		Str("task", task.UUID).
		Str("job", job.UUID).
		Str("taskStatusOld", string(task.Status)).
		Str("taskStatusNew", string(newTaskStatus)).
		Logger()
	logger.Debug().Msg("task state changed")

	task.Status = newTaskStatus
	if err := sm.persist.SaveTask(ctx, task); err != nil {
		return fmt.Errorf("error saving task to database: %w", err)
	}
	if err := sm.updateJobAfterTaskStatusChange(ctx, task, newTaskStatus); err != nil {
		return fmt.Errorf("error updating job after task status change: %w", err)
	}
	return nil
}

// updateJobAfterTaskStatusChange updates the job status based on the status of
// this task and other tasks in the job.
func (sm *StateMachine) updateJobAfterTaskStatusChange(
	ctx context.Context, task *persistence.Task, newTaskStatus api.TaskStatus,
) error {

	job := task.Job

	logger := log.With().
		Str("job", job.UUID).
		Str("task", task.UUID).
		Str("taskStatusOld", string(task.Status)).
		Str("taskStatusNew", string(newTaskStatus)).
		Logger()

	// If the job has status 'ifStatus', move it to status 'thenStatus'.
	jobStatusIfAThenB := func(ifStatus, thenStatus api.JobStatus) error {
		if job.Status != ifStatus {
			return nil
		}
		logger.Info().
			Str("jobStatusOld", string(ifStatus)).
			Str("jobStatusNew", string(thenStatus)).
			Msg("Job changed status because one of its task changed status")
		return sm.JobStatusChange(ctx, job, thenStatus)
	}

	// Every 'case' in this switch MUST return. Just for sanity's sake.
	switch newTaskStatus {
	case api.TaskStatusQueued:
		// Re-queueing a task on a completed job should re-queue the job too.
		return jobStatusIfAThenB(api.JobStatusCompleted, api.JobStatusQueued)

	case api.TaskStatusCancelRequested:
		// Requesting cancellation of a single task has no influence on the job itself.
		return nil

	case api.TaskStatusPaused:
		// Pausing a task has no impact on the job.
		return nil

	case api.TaskStatusCanceled:
		// Only trigger cancellation/failure of the job if that was actually requested.
		// A user can also cancel a single task from the web UI or API, in which
		// case the job should just keep running.
		if job.Status != api.JobStatusCancelRequested {
			return nil
		}
		// This could be the last 'cancel-requested' task to go to 'canceled'.
		hasCancelReq, err := sm.persist.JobHasTasksInStatus(ctx, job, api.TaskStatusCancelRequested)
		if err != nil {
			return err
		}
		if !hasCancelReq {
			logger.Info().Msg("last task of job went from cancel-requested to canceled")
			return sm.JobStatusChange(ctx, job, api.JobStatusCanceled)
		}
		return nil

	case api.TaskStatusFailed:
		// Count the number of failed tasks. If it is over the threshold, fail the job.
		numFailed, numTotal, err := sm.persist.CountTasksOfJobInStatus(ctx, job, api.TaskStatusFailed)
		if err != nil {
			return err
		}
		failedPercentage := int(float64(numFailed) / float64(numTotal) * 100)
		failLogger := logger.With().
			Int("taskNumTotal", numTotal).
			Int("taskNumFailed", numFailed).
			Int("failedPercentage", failedPercentage).
			Int("threshold", taskFailJobPercentage).
			Logger()

		if failedPercentage >= taskFailJobPercentage {
			failLogger.Info().Msg("failing job because too many of its tasks failed")
			return sm.JobStatusChange(ctx, job, api.JobStatusFailed)
		}
		// If the job didn't fail, this failure indicates that at least the job is active.
		failLogger.Info().Msg("task failed, but not enough to fail the job")
		return jobStatusIfAThenB(api.JobStatusQueued, api.JobStatusActive)

	case api.TaskStatusActive, api.TaskStatusSoftFailed:
		switch job.Status {
		case api.JobStatusActive, api.JobStatusCancelRequested:
			// Do nothing, job is already in the desired status.
			return nil
		default:
			logger.Info().Msg("job became active because one of its task changed status")
			return sm.JobStatusChange(ctx, job, api.JobStatusActive)
		}

	case api.TaskStatusCompleted:
		numComplete, numTotal, err := sm.persist.CountTasksOfJobInStatus(ctx, job, api.TaskStatusCompleted)
		if err != nil {
			return err
		}
		if numComplete == numTotal {
			logger.Info().Msg("all tasks of job are completed, job is completed")
			return sm.JobStatusChange(ctx, job, api.JobStatusCompleted)
		}
		logger.Info().
			Int("taskNumTotal", numTotal).
			Int("taskNumComplete", numComplete).
			Msg("task completed; there are more tasks to do")
		return jobStatusIfAThenB(api.JobStatusQueued, api.JobStatusActive)

	default:
		logger.Warn().Msg("task obtained status that Flamenco did not expect")
		return nil
	}
}

func (sm *StateMachine) JobStatusChange(ctx context.Context, job *persistence.Job, newJobStatus api.JobStatus) error {
	logger := log.With().
		Str("job", job.UUID).
		Str("jobStatusOld", string(job.Status)).
		Str("jobStatusNew", string(newJobStatus)).
		Logger()

	logger.Info().Msg("job status changed")

	// TODO: actually respond to status change, instead of just saving the new job state.

	job.Status = newJobStatus
	return sm.persist.SaveJobStatus(ctx, job)
}
