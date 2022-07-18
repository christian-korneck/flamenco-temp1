package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

func (w *Worker) gotoStateOffline(context.Context) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	w.state = api.WorkerStatusOffline

	logger := log.With().Int("pid", os.Getpid()).Logger()
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to find our own process for clean shutdown")
	}

	logger.Warn().Msg("sending our own process an interrupt signal")
	err = proc.Signal(os.Interrupt)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to send interrupt signal to our own process")
	}
}

// SignOff forces the worker in shutdown state and acknlowedges this to the Manager.
// Does NOT actually peform a shutdown; is intended to be called while shutdown is in progress.
func (w *Worker) SignOff(ctx context.Context) {
	w.stateMutex.Lock()
	w.state = api.WorkerStatusOffline
	logger := log.With().Str("state", string(w.state)).Logger()
	w.stateMutex.Unlock()

	logger.Info().Msg("signing off at Manager")

	SignOff(ctx, logger, w.client)
}

// SignOff sends a signoff request to the Manager.
// Any error is logged but not returned.
func SignOff(ctx context.Context, logger zerolog.Logger, client FlamencoClient) {
	resp, err := client.SignOffWithResponse(ctx)
	switch {
	case err != nil:
		logger.Error().Err(err).Msg("unable to sign off at Manager")
	case resp.JSONDefault != nil:
		logger.Error().Interface("error", resp.JSONDefault).Msg("error received when signing off at Manager")
	}
}
