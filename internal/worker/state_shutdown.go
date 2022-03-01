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
	"os"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

func (w *Worker) gotoStateShutdown(context.Context) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	w.state = api.WorkerStatusShutdown

	logger := log.With().Int("pid", os.Getpid()).Logger()
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to find our own process for clean shutdown")
	}

	logger.Warn().Msg("sending our own process an interrupt signal")
	err = proc.Signal(os.Interrupt)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to find send interrupt signal to our own process")
	}
}

// SignOff forces the worker in shutdown state and acknlowedges this to the Manager.
// Does NOT actually peform a shutdown; is intended to be called while shutdown is in progress.
func (w *Worker) SignOff(ctx context.Context) {
	w.stateMutex.Lock()
	w.state = api.WorkerStatusShutdown
	logger := log.With().Str("state", string(w.state)).Logger()
	w.stateMutex.Unlock()

	logger.Info().Msg("signing off at Manager")

	resp, err := w.client.SignOffWithResponse(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("unable to sign off at Manager")
		return
	}
	if resp.JSONDefault != nil {
		logger.Error().Interface("error", resp.JSONDefault).Msg("error received when signing off at Manager")
		return
	}
}
