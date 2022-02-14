package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

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

var _ CommandListener = (*Listener)(nil)
var _ TaskExecutionListener = (*Listener)(nil)

// Listener listens to the result of task and command execution, and sends it to the Manager.
type Listener struct {
	doneWg *sync.WaitGroup
	client api.ClientWithResponsesInterface
}

// NewListener creates a new Listener that will send updates to the API client.
func NewListener(client api.ClientWithResponsesInterface) *Listener {
	l := &Listener{
		doneWg: new(sync.WaitGroup),
		client: client,
	}
	l.doneWg.Add(1)
	return l
}

func (l *Listener) Run(ctx context.Context) {
	keepRunning := true
	for keepRunning {
		select {
		case <-ctx.Done():
			keepRunning = false
			continue
		case <-time.After(10 * time.Second):
			// This is just a dummy thing.
		}
		log.Debug().Msg("listener is still running")
	}

	log.Debug().Msg("listener shutting down")
	l.doneWg.Done()
}

func (l *Listener) Wait() {
	log.Debug().Msg("waiting for listener to shut down")
	l.doneWg.Wait()
}

// TaskStarted tells the Manager that task execution has started.
func (l *Listener) TaskStarted(taskID TaskID) error {
	return errors.New("not implemented")
}

// TaskFailed tells the Manager the task failed for some reason.
func (l *Listener) TaskFailed(taskID TaskID, reason string) error {
	return errors.New("not implemented")
}

// TaskCompleted tells the Manager the task has been completed.
func (l *Listener) TaskCompleted(taskID TaskID) error {
	return errors.New("not implemented")
}

// LogProduced sends any logging to whatever service for storing logging.
func (l *Listener) LogProduced(taskID TaskID, logLines ...string) error {
	return errors.New("not implemented")
}

// OutputProduced tells the Manager there has been some output (most commonly a rendered frame or video).
func (l *Listener) OutputProduced(taskID TaskID, outputLocation string) error {
	return errors.New("not implemented")
}
