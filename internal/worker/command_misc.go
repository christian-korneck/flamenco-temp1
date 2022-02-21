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

/* This file contains the commands in the "misc" type group. */

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

// cmdEcho executes the "echo" command.
func (ce *CommandExecutor) cmdEcho(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {
	message, ok := cmd.Parameters["message"]
	if !ok {
		return fmt.Errorf("missing 'message' setting")
	}
	messageStr := fmt.Sprintf("%v", message)

	logger.Info().Str("message", messageStr).Msg("echo")
	if err := ce.listener.LogProduced(ctx, taskID, fmt.Sprintf("echo: %q", messageStr)); err != nil {
		return err
	}
	return nil
}

// cmdSleep executes the "sleep" command.
func (ce *CommandExecutor) cmdSleep(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {

	sleepTime, ok := cmd.Parameters["duration_in_seconds"]
	if !ok {
		return errors.New("missing setting 'duration_in_seconds'")
	}

	var duration time.Duration
	switch v := sleepTime.(type) {
	case int:
		duration = time.Duration(v) * time.Second
	case float64:
		duration = time.Duration(v) * time.Second
	default:
		log.Warn().Interface("duration_in_seconds", v).Msg("bad type for setting 'duration_in_seconds', expected int")
		return fmt.Errorf("bad type for setting 'duration_in_seconds', expected int, not %T", v)
	}

	log.Info().Str("duration", duration.String()).Msg("sleep")

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Warn().Err(err).Msg("sleep aborted because context closed")
		return fmt.Errorf("sleep aborted because context closed: %w", err)
	case <-ce.timeService.After(duration):
		log.Debug().Msg("sleeping done")
	}

	if err := ce.listener.LogProduced(ctx, taskID, fmt.Sprintf("slept %v", duration)); err != nil {
		return err
	}

	return nil
}
