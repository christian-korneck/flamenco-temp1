package worker

// SPDX-License-Identifier: GPL-3.0-or-later

/* This file contains the commands in the "misc" type group. */

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

// cmdEcho executes the "echo" command.
func (ce *CommandExecutor) cmdEcho(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {
	message, ok := cmd.Parameters["message"]
	if !ok {
		return NewParameterMissingError("message", cmd)
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
		return NewParameterMissingError("duration_in_seconds", cmd)
	}

	var duration time.Duration
	switch v := sleepTime.(type) {
	case int:
		duration = time.Duration(v) * time.Second
	case float64:
		duration = time.Duration(v) * time.Second
	default:
		log.Warn().Interface("duration_in_seconds", v).Msg("bad type for setting 'duration_in_seconds', expected int")
		return NewParameterInvalidError("duration_in_seconds", cmd, "bad type %T, expecting int or float", sleepTime)
	}

	logger = log.With().Str("duration", duration.String()).Logger()
	if duration < 0 {
		logger.Error().Msg("cannot sleep negative durations")
		return NewParameterInvalidError("duration_in_seconds", cmd, "cannot be negative")
	}
	logger.Info().Msg("sleep")

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Warn().Msg("sleep command aborted due to context shutdown")
		return fmt.Errorf("sleep command aborted due to context shutdown: %w", err)
	case <-ce.timeService.After(duration):
		log.Debug().Msg("sleeping done")
	}

	if err := ce.listener.LogProduced(ctx, taskID, fmt.Sprintf("slept %v", duration)); err != nil {
		return err
	}

	return nil
}
