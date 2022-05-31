// SPDX-License-Identifier: GPL-3.0-or-later
package webupdates

import (
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

// NewWorkerUpdate returns a partial SocketIOWorkerUpdate struct for the given worker.
// It only fills in the fields that represent the current state of the worker. For
// example, it omits `PreviousStatus`. The ommitted fields can be filled in by
// the caller.
func NewWorkerUpdate(worker *persistence.Worker) api.SocketIOWorkerUpdate {
	workerUpdate := api.SocketIOWorkerUpdate{
		Id:       worker.UUID,
		Nickname: worker.Name,
		Status:   worker.Status,
		Version:  worker.Software,
		Updated:  worker.UpdatedAt,
	}

	if worker.StatusRequested != "" {
		workerUpdate.StatusRequested = &worker.StatusRequested
		workerUpdate.LazyStatusRequest = &worker.LazyStatusRequest
	}

	return workerUpdate
}

// BroadcastWorkerUpdate sends the worker update to clients.
func (b *BiDirComms) BroadcastWorkerUpdate(workerUpdate api.SocketIOWorkerUpdate) {
	log.Debug().Interface("workerUpdate", workerUpdate).Msg("socketIO: broadcasting worker update")
	b.BroadcastTo(SocketIORoomWorkers, SIOEventWorkerUpdate, workerUpdate)
}

// BroadcastNewWorker sends a "new worker" notification to clients.
func (b *BiDirComms) BroadcastNewWorker(workerUpdate api.SocketIOWorkerUpdate) {
	if workerUpdate.PreviousStatus != nil {
		log.Warn().Interface("workerUpdate", workerUpdate).Msg("socketIO: new workers should not have a previous state")
		workerUpdate.PreviousStatus = nil
	}

	log.Debug().Interface("workerUpdate", workerUpdate).Msg("socketIO: broadcasting new worker")
	b.BroadcastTo(SocketIORoomWorkers, SIOEventWorkerUpdate, workerUpdate)
}
