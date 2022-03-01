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
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/appinfo"
	"git.blender.org/flamenco/internal/worker"
)

var (
	w                *worker.Worker
	listener         *worker.Listener
	buffer           *worker.UpstreamBufferDB
	shutdownComplete chan struct{}
)

var cliArgs struct {
	version    bool
	verbose    bool
	debug      bool
	managerURL *url.URL
	manager    string
	register   bool
}

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

	// Startup can take arbitrarily long, as it only ends when the Manager can be
	// reached and accepts our sign-on request. An offline Manager would cause the
	// Worker to wait for it indefinitely.
	startupCtx := context.Background()
	client, startupState := worker.RegisterOrSignOn(startupCtx, configWrangler)

	shutdownComplete = make(chan struct{})
	workerCtx, workerCtxCancel := context.WithCancel(context.Background())

	timeService := clock.New()
	buffer = upstreamBufferOrDie(client, timeService)
	go buffer.Flush(workerCtx) // Immediately try to flush any updates.

	cliRunner := worker.NewCLIRunner()
	listener = worker.NewListener(client, buffer)
	cmdRunner := worker.NewCommandExecutor(cliRunner, listener, timeService)
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
			if err := buffer.Close(shutdownCtx); err != nil {
				log.Error().Err(err).Msg("closing upstream task buffer")
			}
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

func parseCliArgs() {
	flag.BoolVar(&cliArgs.version, "version", false, "Shows the application version, then exits.")
	flag.BoolVar(&cliArgs.verbose, "verbose", false, "Enable info-level logging.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")

	flag.StringVar(&cliArgs.manager, "manager", "", "URL of the Flamenco Manager.")
	flag.BoolVar(&cliArgs.register, "register", false, "(Re-)register at the Manager.")

	flag.Parse()

	if cliArgs.manager != "" {
		var err error
		cliArgs.managerURL, err = worker.ParseURL(cliArgs.manager)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid manager URL")
		}
	}
}

func upstreamBufferOrDie(client worker.FlamencoClient, timeService clock.Clock) *worker.UpstreamBufferDB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	buffer, err := worker.NewUpstreamBuffer(client, timeService)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create task update queue database")
	}

	// TODO: make filename configurable?
	if err := buffer.OpenDB(ctx, "flamenco-worker-queue.db"); err != nil {
		log.Fatal().Err(err).Msg("unable to open task update queue database")
	}

	return buffer
}
