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
	Run(ctx context.Context, taskID TaskID, cmd api.Command) error
}

type TaskExecutionListener interface {
	// TaskStarted tells the Manager that task execution has started.
	TaskStarted(taskID TaskID) error

	// TaskFailed tells the Manager the task failed for some reason.
	TaskFailed(taskID TaskID, reason string) error

	// TaskCompleted tells the Manager the task has been completed.
	TaskCompleted(taskID TaskID) error
}

// TODO: move me to a more appropriate place.
type TaskID string

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

	taskID := TaskID(task.Uuid)

	if err := te.listener.TaskStarted(taskID); err != nil {
		return fmt.Errorf("error sending notification to manager: %w", err)
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

		err := te.cmdRunner.Run(ctx, taskID, cmd)

		if err != nil {
			if err := te.listener.TaskFailed(taskID, err.Error()); err != nil {
				return fmt.Errorf("error sending notification to manager: %w", err)
			}
			return err
		}
	}

	if err := te.listener.TaskCompleted(taskID); err != nil {
		return fmt.Errorf("error sending notification to manager: %w", err)
	}

	return nil
}
