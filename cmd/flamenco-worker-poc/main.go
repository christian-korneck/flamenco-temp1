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
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/blender/flamenco-ng-poc/internal/appinfo"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker"
)

var (
	w                *worker.Worker
	listener         *worker.Listener
	shutdownComplete chan struct{}
)

func main() {
	parseCliArgs()
	if cliArgs.version {
		fmt.Println(appinfo.ApplicationVersion)
		return
	}

	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	log.Info().
		Str("version", appinfo.ApplicationVersion).
		Str("OS", runtime.GOOS).
		Str("ARCH", runtime.GOARCH).
		Int("pid", os.Getpid()).
		Msgf("starting %v Worker", appinfo.ApplicationName)

	configWrangler := worker.NewConfigWrangler()

	startupCtx, sctxCancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	client, startupState := registerOrSignOn(startupCtx, configWrangler)
	sctxCancelFunc()

	shutdownComplete = make(chan struct{})
	workerCtx, workerCtxCancel := context.WithCancel(context.Background())

	listener = worker.NewListener(client)
	timeService := clock.New()
	cmdRunner := worker.NewCommandExecutor(listener, timeService)
	taskRunner := worker.NewTaskExecutor(cmdRunner, listener)
	w = worker.NewWorker(client, taskRunner)

	// Handle Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for signum := range c {
			workerCtxCancel()
			// Run the shutdown sequence in a goroutine, so that multiple Ctrl+C presses can be handled in parallel.
			go shutdown(signum)
		}
	}()

	go listener.Run(workerCtx)
	go w.Start(workerCtx, startupState)

	<-shutdownComplete

	log.Debug().Msg("process shutting down")
}

func shutdown(signum os.Signal) {
	done := make(chan struct{})
	go func() {
		log.Info().Str("signal", signum.String()).Msg("signal received, shutting down.")

		if w != nil {
			shutdownCtx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancelFunc()
			w.SignOff(shutdownCtx)
			w.Close()
			listener.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
		log.Debug().Msg("shutdown OK")
	case <-time.After(20 * time.Second):
		log.Error().Msg("shutdown forced, stopping process.")
		os.Exit(-2)
	}

	log.Warn().Msg("shutdown complete, stopping process.")
	close(shutdownComplete)
}
