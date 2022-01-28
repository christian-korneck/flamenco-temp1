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
	"crypto/rand"
	"encoding/hex"
	"os"
	"runtime"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

// (Re-)register ourselves at the Manager.
func (w *Worker) register(ctx context.Context) {
	// Construct our new password.
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		log.Fatal().Err(err).Msg("unable to generate secret key")
	}
	secretKey := hex.EncodeToString(secret)

	// TODO: load taskTypes from config file.
	taskTypes := []string{"unknown", "sleep", "blender-render", "debug", "ffmpeg"}

	req := api.RegisterWorkerJSONRequestBody{
		Nickname:           mustHostname(),
		Platform:           runtime.GOOS,
		Secret:             secretKey,
		SupportedTaskTypes: taskTypes,
	}
	resp, err := w.client.RegisterWorkerWithResponse(ctx, req)
	if err != nil {
		log.Fatal().Err(err).Msg("error registering at Manager")
	}
	switch {
	case resp.JSON200 != nil:
		log.Info().
			Int("code", resp.StatusCode()).
			Interface("resp", resp.JSON200).
			Msg("registered at Manager")
	default:
		log.Fatal().
			Int("code", resp.StatusCode()).
			Interface("resp", resp.JSONDefault).
			Msg("unable to register at Manager")
	}

	// store ID and secretKey in config file when registration is complete.
	err = w.configWrangler.WriteConfig(credentialsFilename, "Credentials", workerCredentials{
		WorkerID: resp.JSON200.Uuid,
		Secret:   secretKey,
	})
	if err != nil {
		log.Fatal().Err(err).Str("file", credentialsFilename).
			Msg("unable to write credentials configuration file")
	}
}

func (w *Worker) reregister(ctx context.Context) {
	w.register(ctx)
	w.loadConfig()
}

// signOn tells the Manager we're alive and returns the status the Manager tells us to go to.
// Failure to sign on is fatal.
func (w *Worker) signOn(ctx context.Context) api.WorkerStatus {
	logger := log.With().Str("manager", w.manager.String()).Logger()
	logger.Info().Msg("signing on at Manager")

	if w.creds == nil {
		logger.Fatal().Msg("no credentials, unable to sign on")
	}

	// TODO: load taskTypes from config file.
	taskTypes := []string{"unknown", "sleep", "blender-render", "debug", "ffmpeg"}

	req := api.SignOnJSONRequestBody{
		Nickname:           mustHostname(),
		SupportedTaskTypes: taskTypes,
	}
	resp, err := w.client.SignOnWithResponse(ctx, req)
	if err != nil {
		log.Fatal().Err(err).Msg("error registering at Manager")
	}
	switch {
	case resp.JSON200 != nil:
		log.Info().
			Int("code", resp.StatusCode()).
			Interface("resp", resp.JSON200).
			Msg("signed on at Manager")
	default:
		log.Fatal().
			Int("code", resp.StatusCode()).
			Interface("resp", resp.JSONDefault).
			Msg("unable to sign on at Manager")
	}

	startupState := resp.JSON200.StatusRequested
	log.Info().Str("startup_state", string(startupState)).Msg("manager accepted sign-on")
	return startupState
}

// mustHostname either the hostname or logs a fatal error.
func mustHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting hostname")
	}
	return hostname
}
