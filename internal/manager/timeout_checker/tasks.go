package timeout_checker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

func (ttc *TimeoutChecker) checkTasks(ctx context.Context) {
	timeoutThreshold := ttc.clock.Now().UTC().Add(-ttc.taskTimeout)
	logger := log.With().
		Time("threshold", timeoutThreshold.Local()).
		Logger()
	logger.Trace().Msg("TimeoutChecker: finding active tasks that have not been touched since threshold")

	tasks, err := ttc.persist.FetchTimedOutTasks(ctx, timeoutThreshold)
	if err != nil {
		log.Error().Err(err).Msg("TimeoutChecker: error fetching timed-out tasks from database")
		return
	}

	if len(tasks) == 0 {
		logger.Trace().Msg("TimeoutChecker: no timed-out tasks")
		return
	}
	logger.Debug().
		Int("numTasks", len(tasks)).
		Msg("TimeoutChecker: failing all active tasks that have not been touched since threshold")

	for _, task := range tasks {
		ttc.timeoutTask(ctx, task)
	}
}

// timeoutTask marks a task as 'failed' due to a timeout.
func (ttc *TimeoutChecker) timeoutTask(ctx context.Context, task *persistence.Task) {
	workerIdent, logger := ttc.assignedWorker(task)

	task.Activity = fmt.Sprintf("Task timed out on worker %s", workerIdent)
	err := ttc.taskStateMachine.TaskStatusChange(ctx, task, api.TaskStatusFailed)
	if err != nil {
		logger.Error().Err(err).Msg("TimeoutChecker: error saving timed-out task to database")
	}

	err = ttc.logStorage.WriteTimestamped(logger, task.Job.UUID, task.UUID,
		fmt.Sprintf("Task timed out. It was assigned to worker %s, but untouched since %s",
			workerIdent, task.LastTouchedAt.Format(time.RFC3339)))
	if err != nil {
		logger.Error().Err(err).Msg("TimeoutChecker: error writing timeout info to the task log")
	}
}

// assignedWorker returns a description of the worker assigned to this task,
// and a logger configured for it.
func (ttc *TimeoutChecker) assignedWorker(task *persistence.Task) (string, zerolog.Logger) {
	logCtx := log.With().Str("task", task.UUID)

	if task.WorkerID == nil {
		logger := logCtx.Logger()
		logger.Warn().Msg("TimeoutChecker: task timed out, but was not assigned to any worker")
		return "-unassigned-", logger
	}

	if task.Worker == nil {
		logger := logCtx.Logger()
		logger.Warn().Uint("workerDBID", *task.WorkerID).
			Msg("TimeoutChecker: task is assigned to worker that no longer exists")
		return "-unknown-", logger
	}

	logCtx = logCtx.
		Str("worker", task.Worker.UUID).
		Str("workerName", task.Worker.Name)
	logger := logCtx.Logger()
	logger.Warn().Msg("TimeoutChecker: task timed out")

	return task.Worker.Identifier(), logger
}
