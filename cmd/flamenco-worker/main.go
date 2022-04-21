package main

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
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
	// Do-and-quit flags.
	version bool
	flush   bool

	// Logging level flags.
	quiet, debug, trace bool

	managerURL  *url.URL
	findManager bool

	manager  string
	register bool
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
	configLogLevel()

	if cliArgs.findManager {
		// TODO: move this to a more suitable place.
		discoverTimeout := 1 * time.Minute
		discoverCtx, discoverCancel := context.WithTimeout(context.Background(), discoverTimeout)
		defer discoverCancel()
		managerURL, err := worker.AutodiscoverManager(discoverCtx)
		if err != nil {
			logFatalManagerDiscoveryError(err, discoverTimeout)
		}
		log.Info().Str("manager", managerURL).Msg("found Manager")
		return
	}

	configWrangler := worker.NewConfigWrangler()

	// Give the auto-discovery some time to find a Manager.
	discoverTimeout := 10 * time.Minute
	discoverCtx, discoverCancel := context.WithTimeout(context.Background(), discoverTimeout)
	defer discoverCancel()
	if err := worker.MaybeAutodiscoverManager(discoverCtx, &configWrangler); err != nil {
		logFatalManagerDiscoveryError(err, discoverTimeout)
	}

	// Startup can take arbitrarily long, as it only ends when the Manager can be
	// reached and accepts our sign-on request. An offline Manager would cause the
	// Worker to wait for it indefinitely.
	startupCtx := context.Background()
	client, startupState := worker.RegisterOrSignOn(startupCtx, configWrangler)

	shutdownComplete = make(chan struct{})
	workerCtx, workerCtxCancel := context.WithCancel(context.Background())

	timeService := clock.New()
	buffer = upstreamBufferOrDie(client, timeService)
	if queueSize, err := buffer.QueueSize(); err != nil {
		log.Fatal().Err(err).Msg("error checking upstream buffer")
	} else if queueSize > 0 {
		// Flush any updates before actually starting the Worker.
		log.Info().Int("queueSize", queueSize).Msg("flushing upstream buffer")
		buffer.Flush(workerCtx)
	}

	if cliArgs.flush {
		log.Info().Msg("upstream buffer flushed, shutting down")
		workerCtxCancel()
		shutdown()
		return
	}

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
			log.Info().Str("signal", signum.String()).Msg("signal received, shutting down.")

			// Run the shutdown sequence in a goroutine, so that multiple Ctrl+C presses can be handled in parallel.
			workerCtxCancel()
			go shutdown()
		}
	}()

	go listener.Run(workerCtx)
	go w.Start(workerCtx, startupState)

	<-shutdownComplete

	log.Debug().Msg("process shutting down")
}

func shutdown() {
	done := make(chan struct{})
	go func() {
		if w != nil {
			shutdownCtx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancelFunc()
			w.SignOff(shutdownCtx)
			w.Close()
			listener.Wait()
			if err := buffer.Close(); err != nil {
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
	flag.BoolVar(&cliArgs.flush, "flush", false, "Flush any buffered task updates to the Manager, then exits.")

	flag.BoolVar(&cliArgs.quiet, "quiet", false, "Only log warning-level and worse.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")
	flag.BoolVar(&cliArgs.trace, "trace", false, "Enable trace-level logging.")

	// TODO: make this override whatever was stored in the configuration file.
	// flag.StringVar(&cliArgs.manager, "manager", "", "URL of the Flamenco Manager.")
	flag.BoolVar(&cliArgs.register, "register", false, "(Re-)register at the Manager.")
	flag.BoolVar(&cliArgs.findManager, "find-manager", false, "Autodiscover a Manager, then quit.")

	flag.Parse()

	if cliArgs.manager != "" {
		var err error
		cliArgs.managerURL, err = worker.ParseURL(cliArgs.manager)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid manager URL")
		}
	}
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

func upstreamBufferOrDie(client worker.FlamencoClient, timeService clock.Clock) *worker.UpstreamBufferDB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	buffer, err := worker.NewUpstreamBuffer(client, timeService)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create task update queue database")
	}

	// TODO: make filename configurable?
	if err := buffer.OpenDB(ctx, "flamenco-worker.sqlite"); err != nil {
		log.Fatal().Err(err).Msg("unable to open task update queue database")
	}

	return buffer
}

func logFatalManagerDiscoveryError(err error, discoverTimeout time.Duration) {
	if errors.Is(err, context.DeadlineExceeded) {
		log.Fatal().Str("timeout", discoverTimeout.String()).Msg("could not discover Manager in time")
	} else {
		log.Fatal().Err(err).Msg("auto-discovery error")
	}
}
