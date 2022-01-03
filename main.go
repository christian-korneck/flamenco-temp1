package main

import (
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"stuvel.eu/flamenco-test/goja/appinfo"
	"stuvel.eu/flamenco-test/goja/job_compilers"
)

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting %v", appinfo.ApplicationName)

	compiler, err := job_compilers.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading job compilers")
	}

	if err := compiler.Run("simple-blender-render"); err != nil {
		log.Fatal().Err(err).Msg("error running job compiler")
	}
}
