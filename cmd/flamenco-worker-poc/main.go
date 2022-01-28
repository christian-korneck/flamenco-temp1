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
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/blender/flamenco-ng-poc/internal/appinfo"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker/ssdp"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

func main() {
	parseCliArgs()
	if cliArgs.version {
		fmt.Println(appinfo.ApplicationVersion)
		return
	}

	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting %v Worker", appinfo.ApplicationName)

	// configWrangler := worker.NewConfigWrangler()
	managerFinder := ssdp.NewManagerFinder(cliArgs.managerURL)
	// taskRunner := struct{}{}
	findManager(managerFinder)

	// basicAuthProvider, err := securityprovider.NewSecurityProviderBasicAuth("MY_USER", "MY_PASS")
	// if err != nil {
	// 	log.Panic().Err(err).Msg("unable to create basic authr")
	// }

	// flamenco, err := api.NewClientWithResponses(
	// 	"http://localhost:8080/",
	// 	api.WithRequestEditorFn(basicAuthProvider.Intercept),
	// 	api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
	// 		req.Header.Set("User-Agent", appinfo.UserAgent())
	// 		return nil
	// 	}),
	// )
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("error creating client")
	// }

	// w := worker.NewWorker(flamenco, configWrangler, managerFinder, taskRunner)
	// ctx := context.Background()
	// registerWorker(ctx, flamenco)
	// obtainTask(ctx, flamenco)
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

func findManager(managerFinder worker.ManagerFinder) *url.URL {
	finder := managerFinder.FindFlamencoManager()
	select {
	case manager := <-finder:
		log.Info().Str("manager", manager.String()).Msg("found Manager")
		return manager
	case <-time.After(10 * time.Second):
		log.Fatal().Msg("unable to autodetect Flamenco Manager via UPnP/SSDP; configure the URL explicitly")
	}

	return nil
}
