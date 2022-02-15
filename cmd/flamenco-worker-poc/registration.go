package main

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
	"errors"
	"net/http"
	"os"
	"runtime"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/internal/appinfo"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

var errSignOnFailure = errors.New("unable to sign on at Manager")

func registerOrSignOn(ctx context.Context, configWrangler worker.FileConfigWrangler) (
	client api.ClientWithResponsesInterface, startupState api.WorkerStatus,
) {
	// Load configuration
	cfg, err := loadConfig(configWrangler)
	if err != nil {
		log.Fatal().Err(err).Msg("loading configuration")
	}

	log.Info().Interface("config", cfg).Msg("loaded configuration")

	if cfg.Manager == "" {
		log.Fatal().Msg("no manager configured")
	}

	// Load credentials
	creds, err := loadCredentials(configWrangler)
	if err == nil {
		// Credentials can be loaded just fine, try to sign on with them.
		client = authenticatedClient(cfg, creds)
		startupState, err = signOn(ctx, cfg, client)
		if err == nil {
			// Sign on is fine!
			return
		}
	}

	// Either there were no credentials, or existing ones weren't accepted, just register as new worker.
	client = authenticatedClient(cfg, worker.WorkerCredentials{})
	creds = register(ctx, cfg, client)

	// store ID and secretKey in config file when registration is complete.
	err = configWrangler.WriteConfig(credentialsFilename, "Credentials", creds)
	if err != nil {
		log.Fatal().Err(err).Str("file", credentialsFilename).
			Msg("unable to write credentials configuration file")
	}

	// Sign-on should work now.
	client = authenticatedClient(cfg, creds)
	startupState, err = signOn(ctx, cfg, client)
	if err != nil {
		log.Fatal().Err(err).Str("manager", cfg.Manager).Msg("unable to sign on after registering")
	}

	return
}

// (Re-)register ourselves at the Manager.
// Logs a fatal error if unsuccesful.
func register(ctx context.Context, cfg worker.WorkerConfig, client api.ClientWithResponsesInterface) worker.WorkerCredentials {
	// Construct our new password.
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		log.Fatal().Err(err).Msg("unable to generate secret key")
	}
	secretKey := hex.EncodeToString(secret)

	req := api.RegisterWorkerJSONRequestBody{
		Nickname:           mustHostname(),
		Platform:           runtime.GOOS,
		Secret:             secretKey,
		SupportedTaskTypes: cfg.TaskTypes,
	}
	resp, err := client.RegisterWorkerWithResponse(ctx, req)
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

	return worker.WorkerCredentials{
		WorkerID: resp.JSON200.Uuid,
		Secret:   secretKey,
	}
}

// signOn tells the Manager we're alive and returns the status the Manager tells us to go to.
func signOn(ctx context.Context, cfg worker.WorkerConfig, client api.ClientWithResponsesInterface) (api.WorkerStatus, error) {
	logger := log.With().Str("manager", cfg.Manager).Logger()
	logger.Info().Interface("taskTypes", cfg.TaskTypes).Msg("signing on at Manager")

	req := api.SignOnJSONRequestBody{
		Nickname:           mustHostname(),
		SupportedTaskTypes: cfg.TaskTypes,
		SoftwareVersion:    appinfo.ApplicationVersion,
	}
	resp, err := client.SignOnWithResponse(ctx, req)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to send sign-on request")
		return "", errSignOnFailure
	}
	switch {
	case resp.JSON200 != nil:
		log.Info().
			Int("code", resp.StatusCode()).
			Interface("resp", resp.JSON200).
			Msg("signed on at Manager")
	default:
		log.Warn().
			Int("code", resp.StatusCode()).
			Interface("resp", resp.JSONDefault).
			Msg("unable to sign on at Manager")
		return "", errSignOnFailure
	}

	startupState := resp.JSON200.StatusRequested
	log.Info().Str("startup_state", string(startupState)).Msg("manager accepted sign-on")
	return startupState, nil
}

// mustHostname either the hostname or logs a fatal error.
func mustHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting hostname")
	}
	return hostname
}

// authenticatedClient constructs a Flamenco client with the given credentials.
func authenticatedClient(cfg worker.WorkerConfig, creds worker.WorkerCredentials) api.ClientWithResponsesInterface {
	basicAuthProvider, err := securityprovider.NewSecurityProviderBasicAuth(creds.WorkerID, creds.Secret)
	if err != nil {
		log.Panic().Err(err).Msg("unable to create basic auth provider")
	}

	flamenco, err := api.NewClientWithResponses(
		cfg.Manager,
		api.WithRequestEditorFn(basicAuthProvider.Intercept),
		api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("User-Agent", appinfo.UserAgent())
			return nil
		}),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating client")
	}

	return flamenco
}
