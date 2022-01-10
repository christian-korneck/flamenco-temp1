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
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/blender/flamenco-goja-test/internal/appinfo"
	"gitlab.com/blender/flamenco-goja-test/pkg/api"
)

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting %v Worker", appinfo.ApplicationName)

	basicAuthProvider, err := securityprovider.NewSecurityProviderBasicAuth("MY_USER", "MY_PASS")
	if err != nil {
		log.Panic().Err(err).Msg("unable to create basic authr")
	}

	flamenco, err := api.NewClientWithResponses(
		"http://localhost:8080/",
		api.WithRequestEditorFn(basicAuthProvider.Intercept),
		api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("User-Agent", appinfo.UserAgent())
			return nil
		}),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating client")
	}

	ctx := context.Background()
	registerWorker(ctx, flamenco)
	obtainTask(ctx, flamenco)
}

func registerWorker(ctx context.Context, flamenco *api.ClientWithResponses) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting hostname")
	}

	req := api.RegisterWorkerJSONRequestBody{
		Nickname:           hostname,
		Platform:           runtime.GOOS,
		Secret:             "secret",
		SupportedTaskTypes: []string{"sleep", "blender-render", "ffmpeg", "file-management"},
	}
	resp, err := flamenco.RegisterWorkerWithResponse(ctx, req)
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
}

func obtainTask(ctx context.Context, flamenco *api.ClientWithResponses) {
	resp, err := flamenco.ScheduleTaskWithResponse(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("error obtaining task")
	}
	switch {
	case resp.JSON200 != nil:
		log.Info().
			Interface("task", resp.JSON200).
			Msg("obtained task")
	case resp.JSON403 != nil:
		log.Fatal().
			Int("code", resp.StatusCode()).
			Str("error", string(resp.JSON403.Message)).
			Msg("access denied")
	case resp.StatusCode() == http.StatusNoContent:
		log.Info().Msg("no task available")
	default:
		log.Fatal().
			Int("code", resp.StatusCode()).
			Str("error", string(resp.Body)).
			Msg("unable to obtain task")
	}
}
