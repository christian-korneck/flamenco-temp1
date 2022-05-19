package task_state_machine

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/pkg/api"
)

// taskFailJobPercentage is the percentage of a job's tasks that need to fail to
// trigger failure of the entire job.
const taskFailJobPercentage = 10 // Integer from 0 to 100.

// StateMachine handles task and job status changes.
type StateMachine struct {
	persist     PersistenceService
	broadcaster ChangeBroadcaster
}

// Generate mock implementations of these interfaces.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/interfaces_mock.gen.go -package mocks git.blender.org/flamenco/internal/manager/task_state_machine PersistenceService,ChangeBroadcaster

type PersistenceService interface {
	SaveTask(ctx context.Context, task *persistence.Task) error
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

func NewStateMachine(persist PersistenceService, broadcaster ChangeBroadcaster) *StateMachine {
	return &StateMachine{
		persist:     persist,
		broadcaster: broadcaster,
	}
}

// TaskStatusChange updates the task's status to the new one.
// `task` is expected to still have its original status, and have a filled `Job` pointer.
func (sm *StateMachine) TaskStatusChange(
	ctx context.Context,
	task *persistence.Task,
	newTaskStatus api.TaskStatus,
) error {
	oldTaskStatus := task.Status

	if err := sm.taskStatusChangeOnly(ctx, task, newTaskStatus); err != nil {
		return err
	}

	if err := sm.updateJobAfterTaskStatusChange(ctx, task, oldTaskStatus); err != nil {
		return fmt.Errorf("updating job after task status change: %w", err)
	}
	return nil
}

// taskStatusChangeOnly updates the task's status to the new one, but does not "ripple" the change to the job.
// `task` is expected to still have its original status, and have a filled `Job` pointer.
func (sm *StateMachine) taskStatusChangeOnly(
	ctx context.Context,
	task *persistence.Task,
	newTaskStatus api.TaskStatus,
) error {
	job := task.Job
	if job == nil {
		log.Panic().Str("task", task.UUID).Msg("task without job, cannot handle this")
		return nil // Will not run because of the panic.
	}

	oldTaskStatus := task.Status
	task.Status = newTaskStatus

	logger := log.With().
		Str("task", task.UUID).
		Str("job", job.UUID).
		Str("taskStatusOld", string(oldTaskStatus)).
		Str("taskStatusNew", string(newTaskStatus)).
		Logger()
	logger.Debug().Msg("task state changed")

	if err := sm.persist.SaveTask(ctx, task); err != nil {
		return fmt.Errorf("saving task to database: %w", err)
	}

	// Broadcast this change to the SocketIO clients.
	taskUpdate := webupdates.NewTaskUpdate(task)
	taskUpdate.PreviousStatus = &oldTaskStatus
	sm.broadcaster.BroadcastTaskUpdate(taskUpdate)

	return nil
}

// updateJobAfterTaskStatusChange updates the job status based on the status of
// this task and other tasks in the job.
func (sm *StateMachine) updateJobAfterTaskStatusChange(
	ctx context.Context, task *persistence.Task, oldTaskStatus api.TaskStatus,
) error {
	job := task.Job

	logger := log.With().
		Str("job", job.UUID).
		Str("task", task.UUID).
		Str("taskStatusOld", string(oldTaskStatus)).
		Str("taskStatusNew", string(task.Status)).
		Logger()

	// Every 'case' in this switch MUST return. Just for sanity's sake.
	switch task.Status {
	case api.TaskStatusQueued:
		// Re-queueing a task on a completed job should re-queue the job too.
		return sm.jobStatusIfAThenB(ctx, logger, job, api.JobStatusCompleted, api.JobStatusRequeued, "task was queued")

	case api.TaskStatusPaused:
		// Pausing a task has no impact on the job.
		return nil

	case api.TaskStatusCanceled:
		return sm.onTaskStatusCanceled(ctx, logger, job)

	case api.TaskStatusFailed:
		return sm.onTaskStatusFailed(ctx, logger, job)

	case api.TaskStatusActive, api.TaskStatusSoftFailed:
		switch job.Status {
		case api.JobStatusActive, api.JobStatusCancelRequested:
			// Do nothing, job is already in the desired status.
			return nil
		default:
			logger.Info().Msg("job became active because one of its task changed status")
			reason := fmt.Sprintf("task became %s", task.Status)
			return sm.JobStatusChange(ctx, job, api.JobStatusActive, reason)
		}

	case api.TaskStatusCompleted:
		return sm.onTaskStatusCompleted(ctx, logger, job)

	default:
		logger.Warn().Msg("task obtained status that Flamenco did not expect")
		return nil
	}
}

// If the job has status 'ifStatus', move it to status 'thenStatus'.
func (sm *StateMachine) jobStatusIfAThenB(
	ctx context.Context,
	logger zerolog.Logger,
	job *persistence.Job,
	ifStatus, thenStatus api.JobStatus,
	reason string,
) error {
	if job.Status != ifStatus {
		return nil
	}
	logger.Info().
		Str("jobStatusOld", string(ifStatus)).
		Str("jobStatusNew", string(thenStatus)).
		Msg("Job will change status because one of its task changed status")
	return sm.JobStatusChange(ctx, job, thenStatus, reason)
}

// onTaskStatusCanceled conditionally escalates the cancellation of a task to cancel the job.
func (sm *StateMachine) onTaskStatusCanceled(ctx context.Context, logger zerolog.Logger, job *persistence.Job) error {
	// If no more tasks can run, cancel the job.
	numRunnable, _, err := sm.persist.CountTasksOfJobInStatus(ctx, job,
		api.TaskStatusActive, api.TaskStatusQueued, api.TaskStatusSoftFailed)
	if err != nil {
		return err
	}
	if numRunnable == 0 {
		// NOTE: this does NOT cancel any non-runnable (paused/failed) tasks. If that's desired, just cancel the job as a whole.
		logger.Info().Msg("canceled task was last runnable task of job, canceling job")
		return sm.JobStatusChange(ctx, job, api.JobStatusCanceled, "canceled task was last runnable task of job, canceling job")
	}

	return nil
}

// onTaskStatusFailed conditionally escalates the failure of a task to fail the entire job.
func (sm *StateMachine) onTaskStatusFailed(ctx context.Context, logger zerolog.Logger, job *persistence.Job) error {
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
		return sm.JobStatusChange(ctx, job, api.JobStatusFailed, "too many tasks failed")
	}
	// If the job didn't fail, this failure indicates that at least the job is active.
	failLogger.Info().Msg("task failed, but not enough to fail the job")
	return sm.jobStatusIfAThenB(ctx, logger, job, api.JobStatusQueued, api.JobStatusActive,
		"task failed, but not enough to fail the job")
}

// onTaskStatusCompleted conditionally escalates the completion of a task to complete the entire job.
func (sm *StateMachine) onTaskStatusCompleted(ctx context.Context, logger zerolog.Logger, job *persistence.Job) error {
	numComplete, numTotal, err := sm.persist.CountTasksOfJobInStatus(ctx, job, api.TaskStatusCompleted)
	if err != nil {
		return err
	}
	if numComplete == numTotal {
		logger.Info().Msg("all tasks of job are completed, job is completed")
		return sm.JobStatusChange(ctx, job, api.JobStatusCompleted, "all tasks completed")
	}
	logger.Info().
		Int("taskNumTotal", numTotal).
		Int("taskNumComplete", numComplete).
		Msg("task completed; there are more tasks to do")
	return sm.jobStatusIfAThenB(ctx, logger, job, api.JobStatusQueued, api.JobStatusActive, "no more tasks to do")
}

// JobStatusChange gives a Job a new status, and handles the resulting status changes on its tasks.
func (sm *StateMachine) JobStatusChange(
	ctx context.Context,
	job *persistence.Job,
	newJobStatus api.JobStatus,
	reason string,
) error {
	// Job status changes can trigger task status changes, which can trigger the
	// next job status change. Keep looping over these job status changes until
	// there is no more change left to do.
	var err error
	for newJobStatus != "" && newJobStatus != job.Status {
		oldJobStatus := job.Status
		job.Activity = fmt.Sprintf("Changed to status %q: %s", newJobStatus, reason)

		logger := log.With().
			Str("job", job.UUID).
			Str("jobStatusOld", string(oldJobStatus)).
			Str("jobStatusNew", string(newJobStatus)).
			Str("reason", reason).
			Logger()
		logger.Info().Msg("job status changed")

		newJobStatus, err = sm.jobStatusSet(ctx, job, newJobStatus, reason, logger)
		if err != nil {
			return err
		}
	}

	return nil
}

// jobStatusReenforce acts as if the job just transitioned to its current
// status, and performs another round of task status updates. This is normally
// not necessary, but can be used when normal job/task status updates got
// interrupted somehow.
func (sm *StateMachine) jobStatusReenforce(
	ctx context.Context,
	job *persistence.Job,
	reason string,
) error {
	// Job status changes can trigger task status changes, which can trigger the
	// next job status change. Keep looping over these job status changes until
	// there is no more change left to do.
	var err error
	newJobStatus := job.Status

	for {
		oldJobStatus := job.Status
		job.Activity = fmt.Sprintf("Reenforcing status %q: %s", newJobStatus, reason)

		logger := log.With().
			Str("job", job.UUID).
			Str("reason", reason).
			Logger()
		if newJobStatus == job.Status {
			logger := logger.With().
				Str("jobStatus", string(job.Status)).
				Logger()
			logger.Info().Msg("job status reenforced")
		} else {
			logger := logger.With().
				Str("jobStatusOld", string(oldJobStatus)).
				Str("jobStatusNew", string(newJobStatus)).
				Logger()
			logger.Info().Msg("job status changed")
		}

		newJobStatus, err = sm.jobStatusSet(ctx, job, newJobStatus, reason, logger)
		if err != nil {
			return err
		}

		if newJobStatus == "" || newJobStatus == oldJobStatus {
			// Do this check at the end of the loop, and not the start, so that at
			// least one iteration is run.
			break
		}
	}

	return nil
}

// jobStatusSet saves the job with the new status and handles updates to tasks
// as well. If the task status change should trigger another job status change,
// the new job status is returned.
func (sm *StateMachine) jobStatusSet(ctx context.Context,
	job *persistence.Job,
	newJobStatus api.JobStatus,
	reason string,
	logger zerolog.Logger,
) (api.JobStatus, error) {
	oldJobStatus := job.Status
	job.Status = newJobStatus

	// Persist the new job status.
	err := sm.persist.SaveJobStatus(ctx, job)
	if err != nil {
		return "", fmt.Errorf("saving job status change %q to %q to database: %w",
			oldJobStatus, newJobStatus, err)
	}

	// Handle the status change.
	result, err := sm.updateTasksAfterJobStatusChange(ctx, logger, job, oldJobStatus)
	if err != nil {
		return "", fmt.Errorf("updating job's tasks after job status change: %w", err)
	}

	// Broadcast this change to the SocketIO clients.
	jobUpdate := webupdates.NewJobUpdate(job)
	jobUpdate.PreviousStatus = &oldJobStatus
	jobUpdate.RefreshTasks = result.massTaskUpdate
	sm.broadcaster.BroadcastJobUpdate(jobUpdate)

	return result.followingJobStatus, nil
}

// tasksUpdateResult is returned by `updateTasksAfterJobStatusChange`.
type tasksUpdateResult struct {
	// FollowingJobStatus is set when the task updates should trigger another job status update.
	followingJobStatus api.JobStatus
	// massTaskUpdate is true when multiple/all tasks were updated simultaneously.
	// This hasn't triggered individual task updates to SocketIO clients, and thus
	// the resulting SocketIO job update should indicate all tasks must be
	// reloaded.
	massTaskUpdate bool
}

// updateTasksAfterJobStatusChange updates the status of its tasks based on the
// new status of this job.
//
// NOTE: this function assumes that the job already has its new status.
//
// Returns the new state the job should go into after this change, or an empty
// string if there is no subsequent change necessary.
func (sm *StateMachine) updateTasksAfterJobStatusChange(
	ctx context.Context,
	logger zerolog.Logger,
	job *persistence.Job,
	oldJobStatus api.JobStatus,
) (tasksUpdateResult, error) {

	// Every case in this switch MUST return, for sanity sake.
	switch job.Status {
	case api.JobStatusCompleted, api.JobStatusCanceled:
		// Nothing to do; this will happen as a response to all tasks receiving this status.
		return tasksUpdateResult{}, nil

	case api.JobStatusActive:
		// Nothing to do; this happens when a task gets started, which has nothing to
		// do with other tasks in the job.
		return tasksUpdateResult{}, nil

	case api.JobStatusCancelRequested, api.JobStatusFailed:
		jobStatus, err := sm.cancelTasks(ctx, logger, job)
		return tasksUpdateResult{
			followingJobStatus: jobStatus,
			massTaskUpdate:     true,
		}, err

	case api.JobStatusRequeued:
		jobStatus, err := sm.requeueTasks(ctx, logger, job, oldJobStatus)
		return tasksUpdateResult{
			followingJobStatus: jobStatus,
			massTaskUpdate:     true,
		}, err

	case api.JobStatusQueued:
		jobStatus, err := sm.checkTaskCompletion(ctx, logger, job)
		return tasksUpdateResult{
			followingJobStatus: jobStatus,
			massTaskUpdate:     true,
		}, err

	default:
		logger.Warn().Msg("unknown job status change, ignoring")
		return tasksUpdateResult{}, nil
	}
}

// Directly cancel any task that might run in the future.
//
// Returns the next job status, if a status change is required.
func (sm *StateMachine) cancelTasks(
	ctx context.Context, logger zerolog.Logger, job *persistence.Job,
) (api.JobStatus, error) {
	logger.Info().Msg("cancelling tasks of job")

	// Any task that is running or might run in the future should get cancelled.
	taskStatusesToCancel := []api.TaskStatus{
		api.TaskStatusActive,
		api.TaskStatusQueued,
		api.TaskStatusSoftFailed,
	}
	err := sm.persist.UpdateJobsTaskStatusesConditional(
		ctx, job, taskStatusesToCancel, api.TaskStatusCanceled,
		fmt.Sprintf("Manager cancelled this task because the job got status %q.", job.Status),
	)
	if err != nil {
		return "", fmt.Errorf("cancelling tasks of job %s: %w", job.UUID, err)
	}

	// If cancellation was requested, it has now happened, so the job can transition.
	if job.Status == api.JobStatusCancelRequested {
		logger.Info().Msg("all tasks of job cancelled, job can go to 'cancelled' status")
		return api.JobStatusCanceled, nil
	}

	// This could mean cancellation was triggered by failure of the job, in which
	// case the job is already in the correct status.
	return "", nil
}

// requeueTasks re-queues all tasks of the job.
//
// This function assumes that the current job status is "requeued".
//
// Returns the new job status, if this status transition should be followed by
// another one.
func (sm *StateMachine) requeueTasks(
	ctx context.Context, logger zerolog.Logger, job *persistence.Job, oldJobStatus api.JobStatus,
) (api.JobStatus, error) {
	if job.Status != api.JobStatusRequeued {
		logger.Warn().Msg("unexpected job status in StateMachine::requeueTasks()")
	}

	var err error

	switch oldJobStatus {
	case api.JobStatusUnderConstruction:
		// Nothing to do, the job compiler has just finished its work; the tasks have
		// already been set to 'queued' status.
		logger.Debug().Msg("ignoring job status change")
		return "", nil
	case api.JobStatusCompleted:
		// Re-queue all tasks.
		err = sm.persist.UpdateJobsTaskStatuses(ctx, job, api.TaskStatusQueued,
			fmt.Sprintf("Queued because job transitioned status from %q to %q", oldJobStatus, job.Status))
	default:
		statusesToUpdate := []api.TaskStatus{
			api.TaskStatusCanceled,
			api.TaskStatusFailed,
			api.TaskStatusPaused,
			api.TaskStatusSoftFailed,
		}
		// Re-queue only the non-completed tasks.
		err = sm.persist.UpdateJobsTaskStatusesConditional(ctx, job,
			statusesToUpdate, api.TaskStatusQueued,
			fmt.Sprintf("Queued because job transitioned status from %q to %q", oldJobStatus, job.Status))
	}
	if err != nil {
		return "", fmt.Errorf("queueing tasks of job %s: %w", job.UUID, err)
	}

	// TODO: also reset the 'failed by workers' blacklist.

	// The appropriate tasks have been requeued, so now the job can go from "requeued" to "queued".
	return api.JobStatusQueued, nil
}

// checkTaskCompletion returns "completed" as next job status when all tasks of
// the job are completed.
//
// Returns the new job status, if this status transition should be followed by
// another one.
func (sm *StateMachine) checkTaskCompletion(
	ctx context.Context, logger zerolog.Logger, job *persistence.Job,
) (api.JobStatus, error) {

	numCompleted, numTotal, err := sm.persist.CountTasksOfJobInStatus(ctx, job, api.TaskStatusCompleted)
	if err != nil {
		return "", fmt.Errorf("checking task completion of job %s: %w", job.UUID, err)
	}

	if numCompleted < numTotal {
		logger.Debug().
			Int("numTasksCompleted", numCompleted).
			Int("numTasksTotal", numTotal).
			Msg("not all tasks of job are completed")
		return "", nil
	}

	logger.Info().Msg("job has all tasks completed, transition job to 'completed'")
	return api.JobStatusCompleted, nil
}

// CheckStuck finds jobs that are 'stuck' in their current status. This is meant
// to run at startup of Flamenco Manager, and checks to see if there are any
// jobs in a status that a human will not be able to fix otherwise.
func (sm *StateMachine) CheckStuck(ctx context.Context) {
	stuckJobs, err := sm.persist.FetchJobsInStatus(ctx, api.JobStatusCancelRequested, api.JobStatusRequeued)
	if err != nil {
		log.Error().Err(err).Msg("unable to fetch stuck jobs")
		return
	}

	for _, job := range stuckJobs {
		err := sm.jobStatusReenforce(ctx, job, "checking stuck jobs")
		if err != nil {
			log.Error().Str("job", job.UUID).Err(err).Msg("error getting job un-stuck")
		}
	}
}
