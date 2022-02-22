package worker

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
	"sync"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

// Worker performs regular Flamenco Worker operations.
type Worker struct {
	doneChan chan struct{}
	doneWg   *sync.WaitGroup

	client FlamencoClient

	state         api.WorkerStatus
	stateStarters map[api.WorkerStatus]StateStarter // gotoStateXXX functions
	stateMutex    *sync.Mutex

	taskRunner TaskRunner
}

type StateStarter func(context.Context)

type TaskRunner interface {
	Run(ctx context.Context, task api.AssignedTask) error
}

// NewWorker constructs and returns a new Worker.
func NewWorker(
	flamenco FlamencoClient,
	taskRunner TaskRunner,
) *Worker {

	worker := &Worker{
		doneChan: make(chan struct{}),
		doneWg:   new(sync.WaitGroup),

		client: flamenco,

		state:         api.WorkerStatusStarting,
		stateStarters: make(map[api.WorkerStatus]StateStarter),
		stateMutex:    new(sync.Mutex),

		taskRunner: taskRunner,
	}
	worker.setupStateMachine()
	return worker
}

// Start starts the worker by sending it to the given state.
func (w *Worker) Start(ctx context.Context, state api.WorkerStatus) {
	w.changeState(ctx, state)
}

// Close gracefully shuts down the Worker.
func (w *Worker) Close() {
	log.Debug().Msg("worker gracefully shutting down")
	close(w.doneChan)
	w.doneWg.Wait()
	log.Debug().Msg("worker shut down")
}
