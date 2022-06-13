package timeout_checker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
	"github.com/rs/zerolog/log"
)

func (ttc *TimeoutChecker) checkWorkers(ctx context.Context) {
	timeoutThreshold := ttc.clock.Now().UTC().Add(-ttc.workerTimeout)
	logger := log.With().
		Time("threshold", timeoutThreshold.Local()).
		Logger()
	logger.Trace().Msg("TimeoutChecker: finding all awake workers that have not been seen since threshold")

	workers, err := ttc.persist.FetchTimedOutWorkers(ctx, timeoutThreshold)
	if err != nil {
		log.Error().Err(err).Msg("TimeoutChecker: error fetching timed-out workers from database")
		return
	}

	if len(workers) == 0 {
		logger.Trace().Msg("TimeoutChecker: no timed-out workers")
		return
	}
	logger.Debug().
		Int("numWorkers", len(workers)).
		Msg("TimeoutChecker: failing all awake workers that have not been seen since threshold")

	for _, worker := range workers {
		ttc.timeoutWorker(ctx, worker)
	}
}

// timeoutTask marks a task as 'failed' due to a timeout.
func (ttc *TimeoutChecker) timeoutWorker(ctx context.Context, worker *persistence.Worker) {
	logger := log.With().
		Str("worker", worker.UUID).
		Str("name", worker.Name).
		Str("lastSeenAt", worker.LastSeenAt.String()).
		Logger()
	logger.Warn().Msg("TimeoutChecker: worker timed out")

	prevStatus := worker.Status
	worker.Status = api.WorkerStatusError
	worker.StatusChangeClear()

	err := ttc.persist.SaveWorker(ctx, worker)
	if err != nil {
		logger.Error().Err(err).Msg("TimeoutChecker: error saving timed-out worker to database")
	}

	err = ttc.taskStateMachine.RequeueTasksOfWorker(ctx, worker, "worker timed out")
	if err != nil {
		logger.Error().Err(err).Msg("TimeoutChecker: error re-queueing tasks of timed-out worker")
	}

	// Broadcast worker change via SocketIO
	ttc.broadcaster.BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:             worker.UUID,
		Nickname:       worker.Name,
		PreviousStatus: &prevStatus,
		Status:         api.WorkerStatusError,
		Updated:        worker.UpdatedAt,
		Version:        worker.Software,
	})
}
