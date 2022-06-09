package timeout_checker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

// Interval for checking all active tasks for timeouts.
const timeoutCheckInterval = 1 * time.Minute

// Delay for the intial check. This gives workers a chance to reconnect to the Manager
// and send updates after the Manager has started.
const timeoutInitialSleep = 5 * time.Minute

// TimeoutChecker periodically times out tasks and workers if the worker hasn't sent any update recently.
type TimeoutChecker struct {
	taskTimeout time.Duration

	clock            clock.Clock
	persist          PersistenceService
	taskStateMachine TaskStateMachine
	logStorage       LogStorage
}

// New creates a new TimeoutChecker.
func New(
	taskTimeout time.Duration,
	clock clock.Clock,
	persist PersistenceService,
	taskStateMachine TaskStateMachine,
	logStorage LogStorage,
) *TimeoutChecker {
	return &TimeoutChecker{
		taskTimeout: taskTimeout,

		clock:            clock,
		persist:          persist,
		taskStateMachine: taskStateMachine,
		logStorage:       logStorage,
	}
}

// Run runs the timeout checker until the context closes.
func (ttc *TimeoutChecker) Run(ctx context.Context) {
	defer log.Info().Msg("TimeoutChecker: shutting down")

	if ttc.taskTimeout == 0 {
		log.Warn().Msg("TimeoutChecker: no timeout duration configured, will not check for task timeouts")
		return
	}

	log.Info().
		Str("taskTimeout", ttc.taskTimeout.String()).
		Str("initialSleep", timeoutInitialSleep.String()).
		Str("checkInterval", timeoutCheckInterval.String()).
		Msg("TimeoutChecker: starting up")

	// Start with a delay, so that workers get a chance to push their updates
	// after the manager has started up.
	waitDur := timeoutInitialSleep

	for {
		select {
		case <-ctx.Done():
			return
		case <-ttc.clock.After(waitDur):
			waitDur = timeoutCheckInterval
		}
		ttc.checkTasks(ctx)
		// ttc.checkWorkers(ctx)
	}
}

func (ttc *TimeoutChecker) checkTasks(ctx context.Context) {
	timeoutThreshold := ttc.clock.Now().UTC().Add(-ttc.taskTimeout)
	logger := log.With().
		Time("threshold", timeoutThreshold.Local()).
		Logger()
	logger.Debug().Msg("TimeoutChecker: finding active tasks that have not been touched since threshold")

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

// func (ttc *TimeoutChecker) checkWorkers(db *mgo.Database) {
// 	timeoutThreshold := UtcNow().Add(-ttc.config.ActiveWorkerTimeoutInterval)
// 	log.Debugf("Failing all awake workers that have not been seen since %s", timeoutThreshold)

// 	var timedoutWorkers []Worker
// 	// find all awake workers that either have never been seen, or were seen long ago.
// 	query := M{
// 		"status": workerStatusAwake,
// 		"$or": []M{
// 			M{"last_activity": M{"$lte": timeoutThreshold}},
// 			M{"last_activity": M{"$exists": false}},
// 		},
// 	}
// 	projection := M{
// 		"_id":      1,
// 		"nickname": 1,
// 		"address":  1,
// 		"status":   1,
// 	}
// 	if err := db.C("flamenco_workers").Find(query).Select(projection).All(&timedoutWorkers); err != nil {
// 		log.Warningf("Error finding timed-out workers: %s", err)
// 	}

// 	for _, worker := range timedoutWorkers {
// 		worker.Timeout(db, ttc.scheduler)
// 	}
// }
