package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

type CommandRunner interface {
	Run(ctx context.Context, taskID string, cmd api.Command) error
}

// Generate mock implementation of this interface.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/task_exe_listener.gen.go -package mocks git.blender.org/flamenco/internal/worker TaskExecutionListener

// TaskExecutionListener sends task lifecycle events (start/fail/complete) to the Manager.
type TaskExecutionListener interface {
	// TaskStarted tells the Manager that task execution has started.
	TaskStarted(ctx context.Context, taskID string) error

	// TaskFailed tells the Manager the task failed for some reason.
	TaskFailed(ctx context.Context, taskID string, reason string) error

	// TaskCompleted tells the Manager the task has been completed.
	TaskCompleted(ctx context.Context, taskID string) error
}

type TaskExecutor struct {
	cmdRunner CommandRunner
	listener  TaskExecutionListener
}

var _ TaskRunner = (*TaskExecutor)(nil)

func NewTaskExecutor(cmdRunner CommandRunner, listener TaskExecutionListener) *TaskExecutor {
	return &TaskExecutor{
		cmdRunner: cmdRunner,
		listener:  listener,
	}
}

// Run runs a task.
// Returns ErrTaskReassigned when the task was reassigned to another worker.
func (te *TaskExecutor) Run(ctx context.Context, task api.AssignedTask) error {
	logger := log.With().Str("task", task.Uuid).Logger()
	logger.Info().Str("taskType", task.TaskType).Msg("starting task")

	if err := te.listener.TaskStarted(ctx, task.Uuid); err != nil {
		if err == ErrTaskReassigned {
			return ErrTaskReassigned
		}
		return fmt.Errorf("sending 'task started' notification to manager: %w", err)
	}

	for _, cmd := range task.Commands {
		select {
		case <-ctx.Done():
			// Shutdown does not mean task failure; cleanly shutting down will hand
			// back the task for requeueing on the Manager.
			logger.Warn().Msg("task execution aborted due to context shutdown")
			return nil
		default:
		}

		runErr := te.cmdRunner.Run(ctx, task.Uuid, cmd)
		if runErr == nil {
			// All was fine, go run the next command.
			continue
		}
		if errors.Is(runErr, context.Canceled) {
			logger.Warn().Msg("task execution aborted due to context shutdown")
			return nil
		}

		// Notify Manager that this task failed.
		if err := te.listener.TaskFailed(ctx, task.Uuid, runErr.Error()); err != nil {
			if err == ErrTaskReassigned {
				return ErrTaskReassigned
			}
			return fmt.Errorf("sending 'task failed' notification to manager: %w", err)
		}
		return runErr
	}

	if err := te.listener.TaskCompleted(ctx, task.Uuid); err != nil {
		if err == ErrTaskReassigned {
			return ErrTaskReassigned
		}
		return fmt.Errorf("sending 'task completed' notification to manager: %w", err)
	}

	return nil
}
