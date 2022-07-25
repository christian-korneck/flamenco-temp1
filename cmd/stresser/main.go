package main

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/appinfo"
	"git.blender.org/flamenco/internal/stresser"
)

var cliArgs struct {
	quiet, debug, trace bool

	workerID string
	secret   string
}

func main() {
	parseCliArgs()

	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	log.Info().
		Str("version", appinfo.ApplicationVersion).
		Str("OS", runtime.GOOS).
		Str("ARCH", runtime.GOARCH).
		Int("pid", os.Getpid()).
		Msgf("starting %v Worker", appinfo.ApplicationName)
	configLogLevel()

	mainCtx, mainCtxCancel := context.WithCancel(context.Background())

	// Handle Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for signum := range c {
			log.Info().Str("signal", signum.String()).Msg("signal received, shutting down.")
			mainCtxCancel()
		}
	}()

	config := stresser.NewFakeConfig(cliArgs.workerID, cliArgs.secret)
	client := stresser.GetFlamencoClient(mainCtx, config)
	stresser.Run(mainCtx, client)

	log.Info().Msg("signing off at Manager")
	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCtxCancel()
	if _, err := client.SignOffWithResponse(shutdownCtx); err != nil {
		log.Warn().Err(err).Msg("error signing off at Manager")
	}

	log.Info().Msg("stresser shutting down")
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.quiet, "quiet", false, "Only log warning-level and worse.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")
	flag.BoolVar(&cliArgs.trace, "trace", false, "Enable trace-level logging.")

	flag.StringVar(&cliArgs.workerID, "worker", "", "UUID of the Worker")
	flag.StringVar(&cliArgs.secret, "secret", "", "Secret of the Worker")

	flag.Parse()
}

func configLogLevel() {
	var logLevel zerolog.Level
	switch {
	case cliArgs.trace:
		logLevel = zerolog.TraceLevel
	case cliArgs.debug:
		logLevel = zerolog.DebugLevel
	case cliArgs.quiet:
		logLevel = zerolog.WarnLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
}
