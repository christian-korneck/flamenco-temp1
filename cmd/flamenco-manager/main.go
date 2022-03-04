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
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
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
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/swagger_ui"
	"git.blender.org/flamenco/internal/manager/task_logs"
	"git.blender.org/flamenco/internal/manager/task_state_machine"
	"git.blender.org/flamenco/pkg/api"
)

var cliArgs struct {
	version bool
}

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
	configService.Load()

	// TODO: enable TLS via Let's Encrypt.
	listen := configService.Get().Listen
	_, port, _ := net.SplitHostPort(listen)
	log.Info().Str("port", port).Msg("listening")

	// Construct the services.
	persist := openDB(*configService)
	flamenco := buildFlamencoAPI(configService, persist)
	e := buildWebService(flamenco, persist)

	// Handle Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for signum := range c {
			log.Info().Str("signal", signum.String()).Msg("signal received, shutting down")
			mainCtxCancel()
		}
	}()

	// All main goroutines should sync with this waitgroup. Once the waitgroup is
	// done, the main() function will return and the process will stop.
	wg := new(sync.WaitGroup)

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

	wg.Wait()
	log.Info().Msg("shutdown complete")
}

func buildFlamencoAPI(configService *config.Service, persist *persistence.DB) api.ServerInterface {
	timeService := clock.New()
	compiler, err := job_compilers.Load(timeService)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading job compilers")
	}
	logStorage := task_logs.NewStorage(configService.Get().TaskLogsPath)
	taskStateMachine := task_state_machine.NewStateMachine(persist)
	flamenco := api_impl.NewFlamenco(compiler, persist, logStorage, configService, taskStateMachine)
	return flamenco
}

func buildWebService(flamenco api.ServerInterface, persist api_impl.PersistenceService) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Hook Zerolog onto Echo:
	e.Use(lecho.Middleware(lecho.Config{
		Logger: lecho.From(log.Logger),
	}))

	// Ensure panics when serving a web request won't bring down the server.
	e.Use(middleware.Recover())

	// Load the API definition and enable validation & authentication checks.
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get swagger")
	}
	validator := api_impl.SwaggerValidator(swagger, persist)
	e.Use(validator)

	// Register routes.
	api.RegisterHandlers(e, flamenco)
	swagger_ui.RegisterSwaggerUIStaticFiles(e)
	e.GET("/api/openapi3.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, swagger)
	})

	// Temporarily redirect the index page to the Swagger UI, so that at least you
	// can see something.
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, "/api/swagger-ui/")
	})

	// Log available routes
	routeLogger := log.Level(zerolog.DebugLevel)
	routeLogger.Debug().Msg("available routes:")
	for _, route := range e.Routes() {
		routeLogger.Debug().Msgf("%7s %s", route.Method, route.Path)
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
