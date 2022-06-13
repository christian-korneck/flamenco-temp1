package task_state_machine

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
	"github.com/rs/zerolog/log"
)

// RequeueTasksOfWorker re-queues all active tasks (should be max one) of this worker.
//
// `reason`: a string that can be appended to text like "Task requeued because "
func (sm *StateMachine) RequeueTasksOfWorker(
	ctx context.Context,
	worker *persistence.Worker,
	reason string,
) error {
	logger := log.With().
		Str("worker", worker.UUID).
		Logger()

	// Fetch the tasks to update.
	tasks, err := sm.persist.FetchTasksOfWorkerInStatus(ctx, worker, api.TaskStatusActive)
	if err != nil {
		return fmt.Errorf("fetching tasks of worker %s in status %q: %w", worker.UUID, api.TaskStatusActive, err)
	}

	// Run each task change through the task state machine.
	var lastErr error
	for _, task := range tasks {
		logger.Info().
			Str("task", task.UUID).
			Msg("re-queueing task")

			// Write to task activity that it got requeued because of worker sign-off.
		task.Activity = "Task was requeued by Manager because " + reason
		if err := sm.persist.SaveTaskActivity(ctx, task); err != nil {
			logger.Warn().Err(err).
				Str("task", task.UUID).
				Str("reason", reason).
				Str("activity", task.Activity).
				Msg("error saving task activity to database")
			lastErr = err
		}

		if err := sm.TaskStatusChange(ctx, task, api.TaskStatusQueued); err != nil {
			logger.Warn().Err(err).
				Str("task", task.UUID).
				Str("reason", reason).
				Msg("error queueing task")
			lastErr = err
		}

		_ = sm.logStorage.WriteTimestamped(logger, task.Job.UUID, task.UUID, task.Activity)
	}

	return lastErr
}
