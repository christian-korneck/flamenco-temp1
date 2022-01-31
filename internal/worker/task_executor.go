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
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type TaskExecutor struct{}

var _ TaskRunner = (*TaskExecutor)(nil)

func (te *TaskExecutor) Run(ctx context.Context, task api.AssignedTask) error {
	logger := log.With().Str("task", task.Uuid).Logger()
	logger.Info().Str("taskType", task.TaskType).Msg("starting task")

	for _, cmd := range task.Commands {
		cmdLogger := logger.With().Str("command", cmd.Name).Interface("settings", cmd.Settings).Logger()
		cmdLogger.Info().Msg("running command")

		select {
		case <-ctx.Done():
			cmdLogger.Warn().Msg("command execution aborted due to context shutdown")
		case <-time.After(1 * time.Second):
			cmdLogger.Debug().Msg("mocked duration of command")
		}
	}
	return errors.New("task running not implemented")
}
