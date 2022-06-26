package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

var _ CommandListener = (*Listener)(nil)
var _ TaskExecutionListener = (*Listener)(nil)

var (
	ErrTaskReassigned = errors.New("task was not assigned to this worker")
)

// Listener listens to the result of task and command execution, and sends it to the Manager.
type Listener struct {
	doneWg         *sync.WaitGroup
	client         FlamencoClient
	buffer         UpstreamBuffer
	outputUploader *OutputUploader
}

// UpstreamBuffer can buffer up-stream task updates, in case the Manager cannot be reached.
type UpstreamBuffer interface {
	SendTaskUpdate(ctx context.Context, taskID string, update api.TaskUpdateJSONRequestBody) error
}

// NewListener creates a new Listener that will send updates to the API client.
func NewListener(client FlamencoClient, buffer UpstreamBuffer) *Listener {
	l := &Listener{
		doneWg:         new(sync.WaitGroup),
		client:         client,
		buffer:         buffer,
		outputUploader: NewOutputUploader(client),
	}
	l.doneWg.Add(1)
	return l
}

func (l *Listener) Run(ctx context.Context) {
	defer l.doneWg.Done()
	defer log.Debug().Msg("listener shutting down")

	log.Debug().Msg("listener starting up")
	l.outputUploader.Run(ctx)
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
	msg := fmt.Sprintf("Failed: %v", reason)
	return l.sendTaskUpdate(ctx, taskID, api.TaskUpdateJSONRequestBody{
		Activity:   &msg,
		Log:        &msg, // Make sure that this failure also ends up in the task log.
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
	l.outputUploader.OutputProduced(taskID, outputLocation)
	return nil
}

func (l *Listener) sendTaskUpdate(ctx context.Context, taskID string, update api.TaskUpdateJSONRequestBody) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return l.buffer.SendTaskUpdate(ctx, taskID, update)
}
