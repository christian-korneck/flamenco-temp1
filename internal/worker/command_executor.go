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

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type CommandListener interface {
	// LogProduced sends any logging to whatever service for storing logging.
	LogProduced(taskID TaskID, logLines []string) error
	// OutputProduced tells the Manager there has been some output (most commonly a rendered frame or video).
	OutputProduced(taskID TaskID, outputLocation string) error
}

type CommandExecutor struct {
	listener CommandListener
}

var _ CommandRunner = (*CommandExecutor)(nil)

func NewCommandExecutor(listener CommandListener) *CommandExecutor {
	return &CommandExecutor{
		listener: listener,
	}
}

func (te *CommandExecutor) Run(ctx context.Context, taskID TaskID, cmd api.Command) error {
	logger := log.With().Str("task", string(taskID)).Str("command", cmd.Name).Logger()
	logger.Info().Interface("settings", cmd.Settings).Msg("running command")

	return errors.New("command running not implemented")
}
