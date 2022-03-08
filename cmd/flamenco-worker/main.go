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
	"git.blender.org/flamenco/internal/upnp_ssdp"
	"git.blender.org/flamenco/internal/worker"
	"git.blender.org/flamenco/pkg/api"
)

var (
	w                *worker.Worker
	listener         *worker.Listener
	buffer           *worker.UpstreamBufferDB
	shutdownComplete chan struct{}
)

var cliArgs struct {
	version bool

	quiet, debug, trace bool

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
	configLogLevel()

	configWrangler := worker.NewConfigWrangler()
	maybeAutodiscoverManager(&configWrangler)

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
	flag.BoolVar(&cliArgs.quiet, "quiet", false, "Only log warning-level and worse.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")
	flag.BoolVar(&cliArgs.trace, "trace", false, "Enable trace-level logging.")

	// TODO: make this override whatever was stored in the configuration file.
	// flag.StringVar(&cliArgs.manager, "manager", "", "URL of the Flamenco Manager.")
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
	if err := buffer.OpenDB(ctx, "flamenco-worker-queue.db"); err != nil {
		log.Fatal().Err(err).Msg("unable to open task update queue database")
	}

	return buffer
}

// maybeAutodiscoverManager starts Manager auto-discovery if there is no Manager URL configured yet.
func maybeAutodiscoverManager(configWrangler *worker.FileConfigWrangler) {
	cfg, err := configWrangler.WorkerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to load configuration")
	}

	if cfg.ManagerURL != "" {
		// Manager URL is already known, don't bother with auto-discovery.
		return
	}

	discoverCtx, discoverCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer discoverCancel()

	foundManager, err := autodiscoverManager(discoverCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to discover manager")
	}

	configWrangler.SetManagerURL(foundManager)
}

func autodiscoverManager(ctx context.Context) (string, error) {
	c, err := upnp_ssdp.NewClient(log.Logger)
	if err != nil {
		return "", fmt.Errorf("unable to create UPnP/SSDP client: %w", err)
	}

	log.Info().Msg("auto-discovering Manager via UPnP/SSDP")

	urls, err := c.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to find Manager: %w", err)
	}

	if len(urls) == 0 {
		return "", errors.New("no Manager could be found")
	}

	// Try out the URLs to see which one responds.
	// TODO: parallelise this.
	usableURLs := make([]string, 0)
	for _, url := range urls {
		if pingManager(ctx, url) {
			usableURLs = append(usableURLs, url)
		}
	}

	switch len(usableURLs) {
	case 0:
		return "", fmt.Errorf("autodetected %d URLs, but none were usable", len(urls))

	case 1:
		log.Info().Str("url", usableURLs[0]).Msg("found Manager")
		return usableURLs[0], nil

	default:
		log.Info().
			Strs("urls", usableURLs).
			Str("url", usableURLs[0]).
			Msg("found multiple usable URLs, using the first one")
		return usableURLs[0], nil
	}
}

// pingManager connects to a Manager and returns true if it responds.
func pingManager(ctx context.Context, url string) bool {
	logger := log.With().Str("manager", url).Logger()

	client, err := api.NewClientWithResponses(url)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to create API client with this URL")
		return false
	}

	resp, err := client.GetVersionWithResponse(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to get Flamenco version from Manager")
		return false
	}

	if resp.JSON200 == nil {
		logger.Warn().
			Int("httpStatus", resp.StatusCode()).
			Msg("unable to get Flamenco version, unexpected reply")
		return false
	}

	logger.Info().
		Str("version", resp.JSON200.Version).
		Str("name", resp.JSON200.Name).
		Msg("found Flamenco Manager")
	return true
}
