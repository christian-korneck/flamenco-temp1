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
	"fmt"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type CommandRunner interface {
	Run(ctx context.Context, taskID string, cmd api.Command) error
}

// Generate mock implementation of this interface.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/task_exe_listener.gen.go -package mocks gitlab.com/blender/flamenco-ng-poc/internal/worker TaskExecutionListener

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

func (te *TaskExecutor) Run(ctx context.Context, task api.AssignedTask) error {
	logger := log.With().Str("task", task.Uuid).Logger()
	logger.Info().Str("taskType", task.TaskType).Msg("starting task")

	if err := te.listener.TaskStarted(ctx, task.Uuid); err != nil {
		return fmt.Errorf("error sending 'task started' notification to manager: %w", err)
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

		err := te.cmdRunner.Run(ctx, task.Uuid, cmd)

		if err != nil {
			if err := te.listener.TaskFailed(ctx, task.Uuid, err.Error()); err != nil {
				return fmt.Errorf("error sending 'task failed' notification to manager: %w", err)
			}
			return err
		}
	}

	if err := te.listener.TaskCompleted(ctx, task.Uuid); err != nil {
		return fmt.Errorf("error sending 'task completed' notification to manager: %w", err)
	}

	return nil
}
