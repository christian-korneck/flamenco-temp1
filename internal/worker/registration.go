package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/appinfo"
	"git.blender.org/flamenco/pkg/api"
)

var (
	errSignOnCanceled          = errors.New("sign-on cancelled")                             // For example by closing the context.
	errSignOnRepeatableFailure = errors.New("unable to sign on at Manager, try again later") // For example failed connections
	errSignOnRejected          = errors.New("manager rejected our sign-on credentials")      // Reached Manager, but it rejected our creds.
)

// registerOrSignOn tries to sign on, and if that fails (or there are no credentials) tries to register.
// Returns an authenticated Flamenco OpenAPI client.
func RegisterOrSignOn(ctx context.Context, configWrangler FileConfigWrangler) (
	client FlamencoClient, startupState api.WorkerStatus,
) {
	// Load configuration
	cfg, err := configWrangler.WorkerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("loading configuration")
	}

	log.Info().Interface("config", cfg).Msg("loaded configuration")
	if cfg.ManagerURL == "" {
		log.Fatal().Msg("no Manager configured")
	}

	// Load credentials
	creds, err := configWrangler.WorkerCredentials()
	if err == nil {
		// Credentials can be loaded just fine, try to sign on with them.
		client = authenticatedClient(cfg, creds)
		startupState, err = repeatSignOnUntilAnswer(ctx, cfg, client)
		if err == nil {
			// Sign on is fine!
			return
		}
	}

	// Either there were no credentials, or existing ones weren't accepted, just register as new worker.
	client = authenticatedClient(cfg, WorkerCredentials{})
	creds = register(ctx, cfg, client)

	// store ID and secretKey in config file when registration is complete.
	err = configWrangler.SaveCredentials(creds)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to write credentials file")
	}

	// Sign-on should work now.
	client = authenticatedClient(cfg, creds)
	startupState, err = signOn(ctx, cfg, client)
	if err != nil {
		log.Fatal().Err(err).Str("manager", cfg.ManagerURL).Msg("unable to sign on after registering")
	}

	return
}

// (Re-)register ourselves at the Manager.
// Logs a fatal error if unsuccesful.
func register(ctx context.Context, cfg WorkerConfig, client FlamencoClient) WorkerCredentials {
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

	return WorkerCredentials{
		WorkerID: resp.JSON200.Uuid,
		Secret:   secretKey,
	}
}

// repeatSignOnUntilAnswer tries to sign on, and only returns when it has been able to reach the Manager.
// Return still doesn't mean that the sign-on was succesful; inspect the returned error.
func repeatSignOnUntilAnswer(ctx context.Context, cfg WorkerConfig, client FlamencoClient) (api.WorkerStatus, error) {
	waitTime := 0 * time.Second
	for {
		select {
		case <-ctx.Done():
			return api.WorkerStatus(""), errSignOnCanceled
		case <-time.After(waitTime):
		}

		status, err := signOn(ctx, cfg, client)
		if err == nil {
			// Sign-on was succesful, we're done!
			return status, nil
		}
		if err != errSignOnRepeatableFailure {
			// We shouldn't repeat the sign-on; communication was succesful but somehow our credentials were rejected.
			return status, err
		}

		// Try again after a while.
		waitTime = 5 * time.Second
	}
}

// signOn tells the Manager we're alive and returns the status the Manager tells us to go to.
func signOn(ctx context.Context, cfg WorkerConfig, client FlamencoClient) (api.WorkerStatus, error) {
	logger := log.With().Str("manager", cfg.ManagerURL).Logger()

	req := api.SignOnJSONRequestBody{
		Nickname:           mustHostname(),
		SupportedTaskTypes: cfg.TaskTypes,
		SoftwareVersion:    appinfo.ApplicationVersion,
	}

	logger.Info().
		Str("nickname", req.Nickname).
		Str("softwareVersion", req.SoftwareVersion).
		Interface("taskTypes", req.SupportedTaskTypes).
		Msg("signing on at Manager")

	resp, err := client.SignOnWithResponse(ctx, req)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to send sign-on request")
		return "", errSignOnRepeatableFailure
	}
	switch {
	case resp.JSON200 != nil:
		log.Debug().
			Int("code", resp.StatusCode()).
			Interface("resp", resp.JSON200).
			Msg("signed on at Manager")
	default:
		log.Warn().
			Int("code", resp.StatusCode()).
			Interface("resp", resp.JSONDefault).
			Msg("unable to sign on at Manager")
		return "", errSignOnRejected
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
func authenticatedClient(cfg WorkerConfig, creds WorkerCredentials) FlamencoClient {
	flamenco, err := api.NewClientWithResponses(
		cfg.ManagerURL,

		// Add a Basic HTTP authentication header to every request to Flamenco Manager.
		api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.SetBasicAuth(creds.WorkerID, creds.Secret)
			return nil
		}),

		// Add a User-Agent header to identify this software + its version.
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
