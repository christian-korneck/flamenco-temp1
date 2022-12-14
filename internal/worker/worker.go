package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
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
}
