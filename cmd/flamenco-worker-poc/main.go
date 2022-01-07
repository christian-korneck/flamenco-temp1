package main

import (
	"context"
	"os"
	"runtime"
	"time"

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

	flamenco, err := api.NewClientWithResponses("http://localhost:8080/")
	if err != nil {
		log.Fatal().Err(err).Msg("error creating client")
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting hostname")
	}

	ctx := context.Background()
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
			Int("code", resp.HTTPResponse.StatusCode).
			Interface("resp", resp.JSON200).
			Msg("registered at Manager")
	default:
		log.Warn().
			Int("code", resp.HTTPResponse.StatusCode).
			Interface("resp", resp.JSONDefault).
			Msg("unable to register at Manager")
	}
}
