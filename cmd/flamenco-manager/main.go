package main

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v3"

	"git.blender.org/flamenco/internal/appinfo"
	"git.blender.org/flamenco/internal/manager/api_impl"
	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/manager/last_rendered"
	"git.blender.org/flamenco/internal/manager/local_storage"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/swagger_ui"
	"git.blender.org/flamenco/internal/manager/task_logs"
	"git.blender.org/flamenco/internal/manager/task_state_machine"
	"git.blender.org/flamenco/internal/manager/timeout_checker"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/internal/own_url"
	"git.blender.org/flamenco/internal/upnp_ssdp"
	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/shaman"
)

var cliArgs struct {
	version        bool
	writeConfig    bool
	delayResponses bool
}

const developmentWebInterfacePort = 8081

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)
	log.Info().
		Str("version", appinfo.ApplicationVersion).
		Str("os", runtime.GOOS).
		Str("arch", runtime.GOARCH).
		Msgf("starting %v", appinfo.ApplicationName)

	parseCliArgs()
	if cliArgs.version {
		return
	}

	// The main context determines the lifetime of the application. All
	// long-running goroutines need to keep an eye on this, and stop their work
	// once it closes.
	mainCtx, mainCtxCancel := context.WithCancel(context.Background())

	// Load configuration.
	configService := config.NewService()
	err := configService.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Error().Err(err).Msg("loading configuration")
	}

	if cliArgs.writeConfig {
		err := configService.Save()
		if err != nil {
			log.Error().Err(err).Msg("could not write configuration file")
			os.Exit(1)
		}
		return
	}

	// TODO: enable TLS via Let's Encrypt.
	listen := configService.Get().Listen
	_, port, _ := net.SplitHostPort(listen)
	log.Info().Str("port", port).Msg("listening")

	urls, err := own_url.AvailableURLs("http", listen)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to figure out my own URL")
	}

	ssdp := makeAutoDiscoverable(urls)

	// Construct the services.
	persist := openDB(*configService)

	// Disabled for now. `VACUUM` locks the database, which means that other
	// queries can fail with a "database is locked (5) (SQLITE_BUSY)" error. This
	// situation should be handled gracefully before reinstating the vacuum loop.
	//
	// go persist.PeriodicMaintenanceLoop(mainCtx)

	timeService := clock.New()
	webUpdater := webupdates.New()

	// TODO: the local storage now is hard-coded to use the same sub-directory as the task log storage.
	// This should be refactored so that the task logs storage uses the localStorage object as well.
	localStorage := local_storage.NewNextToExe("task-logs")
	logStorage := task_logs.NewStorage(configService.Get().TaskLogsPath, timeService, webUpdater)

	taskStateMachine := task_state_machine.NewStateMachine(persist, webUpdater, logStorage)
	lastRender := last_rendered.New(localStorage)
	flamenco := buildFlamencoAPI(timeService, configService, persist, taskStateMachine, logStorage, webUpdater, lastRender)
	e := buildWebService(flamenco, persist, ssdp, webUpdater, urls)

	timeoutChecker := timeout_checker.New(
		configService.Get().TaskTimeout,
		configService.Get().WorkerTimeout,
		timeService, persist, taskStateMachine, logStorage, webUpdater)

	installSignalHandler(mainCtxCancel)

	// Before doing anything new, clean up in case we made a mess in an earlier run.
	taskStateMachine.CheckStuck(mainCtx)

	// All main goroutines should sync with this waitgroup. Once the waitgroup is
	// done, the main() function will return and the process will stop.
	wg := new(sync.WaitGroup)

	// Run the "last rendered image" processor.
	wg.Add(1)
	go func() {
		defer wg.Done()
		lastRender.Run(mainCtx)
	}()

	// Start the web server.
	wg.Add(1)
	go func() {
		defer wg.Done()

		// No matter how this function ends, if the HTTP server goes down, so does
		// the application.
		defer mainCtxCancel()

		err := runWebService(mainCtx, e, listen)
		if err != nil {
			log.Error().Err(err).Msg("HTTP server error, shutting down the application")
		}
	}()

	// Start the UPnP/SSDP server.
	if ssdp != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ssdp.Run(mainCtx)
		}()
	}

	// Start the timeout checker.
	wg.Add(1)
	go func() {
		defer wg.Done()
		timeoutChecker.Run(mainCtx)
	}()

	wg.Wait()
	log.Info().Msg("shutdown complete")
}

func buildFlamencoAPI(
	timeService clock.Clock,
	configService *config.Service,
	persist *persistence.DB,
	taskStateMachine *task_state_machine.StateMachine,
	logStorage *task_logs.Storage,
	webUpdater *webupdates.BiDirComms,
	lastRender *last_rendered.LastRenderedProcessor,
) api.ServerInterface {
	compiler, err := job_compilers.Load(timeService)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading job compilers")
	}
	shamanServer := shaman.NewServer(configService.Get().Shaman, nil)
	flamenco := api_impl.NewFlamenco(
		compiler, persist, webUpdater, logStorage, configService,
		taskStateMachine, shamanServer, timeService, lastRender)
	return flamenco
}

func buildWebService(
	flamenco api.ServerInterface,
	persist api_impl.PersistenceService,
	ssdp *upnp_ssdp.Server,
	webUpdater *webupdates.BiDirComms,
	ownURLs []url.URL,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// The request should come in fairly quickly, given that Flamenco is intended
	// to run on a local network.
	e.Server.ReadHeaderTimeout = 1 * time.Second
	// e.Server.ReadTimeout is not set, as this is quite specific per request.
	// Shaman file uploads and websocket connections should be allowed to run
	// quite long, whereas other queries should be relatively short.
	//
	// See https://github.com/golang/go/issues/16100 for more info about current
	// limitations in Go that get in our way here.

	// Hook Zerolog onto Echo:
	e.Use(lecho.Middleware(lecho.Config{
		Logger: lecho.From(log.Logger),
	}))

	// Ensure panics when serving a web request won't bring down the server.
	e.Use(middleware.Recover())

	// For development of the web interface, to get a less predictable order of asynchronous requests.
	e.Use(randomDelayMiddleware)

	// Disabled, as it causes issues with "204 No Content" responses.
	// TODO: investigate & file a bug report. Adding the check on an empty slice
	// seems to fix it:
	//
	// func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	// 	if len(b) == 0 {
	// 		return 0, nil
	// 	}
	// 	... original code of the function ...
	// }
	// e.Use(middleware.Gzip())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: corsOrigins(ownURLs),

		// List taken from https://www.bacancytechnology.com/blog/real-time-chat-application-using-socketio-golang-vuejs/
		AllowHeaders: []string{
			echo.HeaderAccept,
			echo.HeaderAcceptEncoding,
			echo.HeaderAccessControlAllowOrigin,
			echo.HeaderAccessControlRequestHeaders,
			echo.HeaderAccessControlRequestMethod,
			echo.HeaderAuthorization,
			echo.HeaderContentLength,
			echo.HeaderContentType,
			echo.HeaderOrigin,
			echo.HeaderXCSRFToken,
			echo.HeaderXRequestedWith,
			"Cache-Control",
			"Connection",
			"Host",
			"Referer",
			"User-Agent",
			"X-header",
		},
		AllowMethods: []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
	}))

	// Load the API definition and enable validation & authentication checks.
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get swagger")
	}
	validator := api_impl.SwaggerValidator(swagger, persist)
	e.Use(validator)
	registerOAPIBodyDecoders()

	// Register routes.
	api.RegisterHandlers(e, flamenco)
	webUpdater.RegisterHandlers(e)
	swagger_ui.RegisterSwaggerUIStaticFiles(e)
	e.GET("/api/openapi3.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, swagger)
	})

	// Temporarily redirect the index page to the Swagger UI, so that at least you
	// can see something.
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/api/swagger-ui/")
	})

	// Serve UPnP service descriptions.
	if ssdp != nil {
		e.GET(ssdp.DescriptionPath(), func(c echo.Context) error {
			return c.XMLPretty(http.StatusOK, ssdp.Description(), "  ")
		})
	}

	// Log available routes
	routeLogger := log.Level(zerolog.TraceLevel)
	routeLogger.Trace().Msg("available routes:")
	for _, route := range e.Routes() {
		routeLogger.Trace().Msgf("%7s %s", route.Method, route.Path)
	}

	return e
}

// runWebService runs the Echo server, shutting it down when the context closes.
// If there was any other error, it is returned and the entire server should go down.
func runWebService(ctx context.Context, e *echo.Echo, listen string) error {
	serverStopped := make(chan struct{})
	var httpStartErr error = nil
	var httpShutdownErr error = nil

	go func() {
		defer close(serverStopped)
		err := e.Start(listen)
		if err == http.ErrServerClosed {
			log.Info().Msg("HTTP server shut down")
		} else {
			log.Warn().Err(err).Msg("HTTP server unexpectedly shut down")
			httpStartErr = err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("HTTP server stopping because application is shutting down")

		// Do a clean shutdown of the HTTP server.
		err := e.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("error shutting down HTTP server")
			httpShutdownErr = err
		}

		// Wait until the above goroutine has stopped.
		<-serverStopped

		// Return any error that occurred.
		if httpStartErr != nil {
			return httpStartErr
		}
		return httpShutdownErr

	case <-serverStopped:
		// The HTTP server stopped before the application shutdown was signalled.
		// This is unexpected, so take the entire application down with us.
		if httpStartErr != nil {
			return httpStartErr
		}
		return errors.New("unexpected and unexplained shutdown of HTTP server")
	}
}

func parseCliArgs() {
	var quiet, debug, trace bool

	flag.BoolVar(&cliArgs.version, "version", false, "Shows the application version, then exits.")
	flag.BoolVar(&quiet, "quiet", false, "Only log warning-level and worse.")
	flag.BoolVar(&debug, "debug", false, "Enable debug-level logging.")
	flag.BoolVar(&trace, "trace", false, "Enable trace-level logging.")
	flag.BoolVar(&cliArgs.writeConfig, "write-config", false, "Writes configuration to flamenco-manager.yaml, then exits.")
	flag.BoolVar(&cliArgs.delayResponses, "delay", false,
		"Add a random delay to any HTTP responses. This aids in development of Flamenco Manager's web frontend.")

	flag.Parse()

	var logLevel zerolog.Level
	switch {
	case trace:
		logLevel = zerolog.TraceLevel
	case debug:
		logLevel = zerolog.DebugLevel
	case quiet:
		logLevel = zerolog.WarnLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
}

// openDB opens the database or dies.
func openDB(configService config.Service) *persistence.DB {
	dsn := configService.Get().DatabaseDSN
	if dsn == "" {
		log.Fatal().Msg("configure the database in flamenco-manager.yaml")
	}

	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCtxCancel()
	persist, err := persistence.OpenDB(dbCtx, dsn)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("dsn", dsn).
			Msg("error opening database")
	}

	return persist
}

// installSignalHandler spawns a goroutine that handles incoming POSIX signals.
func installSignalHandler(cancelFunc context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		for signum := range signals {
			log.Info().Str("signal", signum.String()).Msg("signal received, shutting down")
			cancelFunc()
		}
	}()
}

func makeAutoDiscoverable(urls []url.URL) *upnp_ssdp.Server {
	ssdp, err := upnp_ssdp.NewServer(log.Logger)
	if err != nil {
		log.Error().Err(err).Msg("error creating UPnP/SSDP server")
		return nil
	}

	ssdp.AddAdvertisementURLs(urls)
	return ssdp
}

// corsOrigins strips everything from the URL that follows the hostname:port, so
// that it's suitable for checking Origin headers of CORS OPTIONS requests.
func corsOrigins(urls []url.URL) []string {
	origins := make([]string, len(urls))

	// TODO: find a way to allow CORS requests during development, but not when
	// running in production.

	for i, url := range urls {
		// Allow the `yarn run dev` webserver do cross-origin requests to this Manager.
		url.Path = ""
		url.Fragment = ""
		url.Host = fmt.Sprintf("%s:%d", url.Hostname(), developmentWebInterfacePort)
		origins[i] = url.String()
	}
	log.Debug().Str("origins", strings.Join(origins, " ")).Msg("acceptable CORS origins")
	return origins
}

// randomDelayMiddleware sleeps for a random period of time, as a development tool for frontend work.
func randomDelayMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)

		// Delay the response a bit.
		var duration int64 = int64(rand.NormFloat64()*250 + 125) // in msec
		if duration > 0 {
			if duration > 1000 {
				duration = 1000 // Cap at one second.
			}
			time.Sleep(time.Duration(duration) * time.Millisecond)
		}
		return err
	}
}

func registerOAPIBodyDecoders() {
	// Register "decoders" so that binary data other than
	// "application/octet-stream" can be handled by our OpenAPI library.
	openapi3filter.RegisterBodyDecoder("image/jpeg", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/png", openapi3filter.FileBodyDecoder)
}
