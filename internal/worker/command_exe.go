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
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

// Generate mock implementation of this interface.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/command_listener.gen.go -package mocks gitlab.com/blender/flamenco-ng-poc/internal/worker CommandListener

// CommandListener sends the result of commands (log, output files) to the Manager.
type CommandListener interface {
	// LogProduced sends any logging to whatever service for storing logging.
	LogProduced(ctx context.Context, taskID string, logLines ...string) error
	// OutputProduced tells the Manager there has been some output (most commonly a rendered frame or video).
	OutputProduced(ctx context.Context, taskID string, outputLocation string) error
}

type CommandExecutor struct {
	listener CommandListener
	// registry maps a command name to a function that runs that command.
	registry map[string]commandCallable

	timeService TimeService
}

var _ CommandRunner = (*CommandExecutor)(nil)

type commandCallable func(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error

// TimeService is a service that operates on time.
type TimeService interface {
	After(duration time.Duration) <-chan time.Time
}

func NewCommandExecutor(listener CommandListener, timeService TimeService) *CommandExecutor {
	ce := &CommandExecutor{
		listener:    listener,
		timeService: timeService,
	}

	// Registry of supported commands. Having this as a map (instead of a big
	// switch statement) makes it possible to do things like reporting the list of
	// supported commands.
	ce.registry = map[string]commandCallable{
		"echo":           ce.cmdEcho,
		"sleep":          ce.cmdSleep,
		"blender-render": ce.cmdBlenderRender,
	}

	return ce
}

func (ce *CommandExecutor) Run(ctx context.Context, taskID string, cmd api.Command) error {
	logger := log.With().Str("task", string(taskID)).Str("command", cmd.Name).Logger()
	logger.Info().Interface("settings", cmd.Settings).Msg("running command")

	runner, ok := ce.registry[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %q", cmd.Name)
	}

	return runner(ctx, logger, taskID, cmd)
}
