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
	"fmt"
	"time"

	"github.com/rs/zerolog"
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
	// registry maps a command name to a function that runs that command.
	registry map[string]commandCallable
}

var _ CommandRunner = (*CommandExecutor)(nil)

type commandCallable func(ctx context.Context, logger zerolog.Logger, taskID TaskID, cmd api.Command) error

func NewCommandExecutor(listener CommandListener) *CommandExecutor {
	ce := &CommandExecutor{
		listener: listener,
	}
	ce.registry = map[string]commandCallable{
		"echo":  ce.cmdEcho,
		"sleep": ce.cmdSleep,
	}
	return ce
}

func (ce *CommandExecutor) Run(ctx context.Context, taskID TaskID, cmd api.Command) error {
	logger := log.With().Str("task", string(taskID)).Str("command", cmd.Name).Logger()
	logger.Info().Interface("settings", cmd.Settings).Msg("running command")

	runner, ok := ce.registry[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %q", cmd.Name)
	}

	return runner(ctx, logger, taskID, cmd)
}

func (ce *CommandExecutor) cmdEcho(ctx context.Context, logger zerolog.Logger, taskID TaskID, cmd api.Command) error {
	message, ok := cmd.Settings["message"]
	if !ok {
		return fmt.Errorf("missing 'message' setting")
	}
	messageStr := fmt.Sprintf("%v", message)

	logger.Info().Str("message", messageStr).Msg("echo")
	logLines := []string{
		fmt.Sprintf("echo: %q", messageStr),
	}

	if err := ce.listener.LogProduced(taskID, logLines); err != nil {
		return err
	}
	return nil
}

func (ce *CommandExecutor) cmdSleep(ctx context.Context, logger zerolog.Logger, taskID TaskID, cmd api.Command) error {

	sleepTime, ok := cmd.Settings["time_in_seconds"]
	if !ok {
		return errors.New("missing setting 'time_in_seconds'")
	}

	var duration time.Duration
	switch v := sleepTime.(type) {
	case int:
		duration = time.Duration(v) * time.Second
	default:
		log.Warn().Interface("time_in_seconds", v).Msg("bad type for setting 'time_in_seconds', expected int")
		return fmt.Errorf("bad type for setting 'time_in_seconds', expected int, not %v", v)
	}

	log.Info().Str("duration", duration.String()).Msg("sleep")
	time.Sleep(duration)

	return nil
}
