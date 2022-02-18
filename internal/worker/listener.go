package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
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

var (
	ErrTaskReassigned = errors.New("task was not assigned to this worker")
)

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

func ptr[T any](value T) *T {
	return &value
}

// TaskStarted tells the Manager that task execution has started.
func (l *Listener) TaskStarted(ctx context.Context, taskID string) error {
	return l.sendTaskUpdate(ctx, taskID, api.TaskUpdateJSONRequestBody{
		Activity:   ptr("Started"),
		TaskStatus: ptr(api.TaskStatusActive),
	})
}

// TaskFailed tells the Manager the task failed for some reason.
func (l *Listener) TaskFailed(ctx context.Context, taskID string, reason string) error {
	return l.sendTaskUpdate(ctx, taskID, api.TaskUpdateJSONRequestBody{
		Activity:   ptr(fmt.Sprintf("Failed: %v", reason)),
		TaskStatus: ptr(api.TaskStatusFailed),
	})
}

// TaskCompleted tells the Manager the task has been completed.
func (l *Listener) TaskCompleted(ctx context.Context, taskID string) error {
	return l.sendTaskUpdate(ctx, taskID, api.TaskUpdateJSONRequestBody{
		Activity:   ptr("Completed"),
		TaskStatus: ptr(api.TaskStatusCompleted),
	})
}

// LogProduced sends any logging to whatever service for storing logging.
func (l *Listener) LogProduced(ctx context.Context, taskID string, logLines ...string) error {
	return l.sendTaskUpdate(ctx, taskID, api.TaskUpdateJSONRequestBody{
		Log: ptr(strings.Join(logLines, "\n")),
	})
}

// OutputProduced tells the Manager there has been some output (most commonly a rendered frame or video).
func (l *Listener) OutputProduced(ctx context.Context, taskID string, outputLocation string) error {
	// TODO: implement
	return nil
}

func (l *Listener) sendTaskUpdate(ctx context.Context, taskID string, update api.TaskUpdateJSONRequestBody) error {
	resp, err := l.client.TaskUpdateWithResponse(ctx, string(taskID), update)
	if err != nil {
		return fmt.Errorf("error sending task update: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusConflict:
		return ErrTaskReassigned
	default:
		return fmt.Errorf("unknown error from Manager, code %d: %v", resp.StatusCode(), resp.JSONDefault)
	}
}
