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
	"net"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v3"

	"gitlab.com/blender/flamenco-ng-poc/internal/appinfo"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/api_impl"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/config"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/swagger_ui"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/task_logs"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

var cliArgs struct {
	version bool
	initDB  bool
}

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)
	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting %v", appinfo.ApplicationName)

	parseCliArgs()
	if cliArgs.version {
		return
	}

	// Load configuration.
	configService := config.NewService()

	if cliArgs.initDB {
		log.Info().Msg("creating databases")
		err := persistence.InitialSetup()
		if err != nil {
			log.Fatal().Err(err).Msg("problem performing initial setup")
		}
		return
	}

	// Open the database.
	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCtxCancel()
	persist, err := persistence.OpenDB(dbCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("error opening database")
	}

	// TODO: load port number from the configuration in the database.
	// TODO: enable TLS via Let's Encrypt.
	listen := ":8080"
	_, port, _ := net.SplitHostPort(listen)
	log.Info().Str("port", port).Msg("listening")

	// Construct the services.
	timeService := clock.New()
	compiler, err := job_compilers.Load(timeService)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading job compilers")
	}
	logStorage := task_logs.NewStorage("./task-logs") // TODO: load job storage path from configuration.
	flamenco := api_impl.NewFlamenco(compiler, persist, logStorage, configService)
	e := buildWebService(flamenco, persist)

	// Start the web server.
	finalErr := e.Start(listen)
	log.Warn().Err(finalErr).Msg("shutting down")
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

	// Log available routes
	routeLogger := log.Level(zerolog.DebugLevel)
	routeLogger.Debug().Msg("available routes:")
	for _, route := range e.Routes() {
		routeLogger.Debug().Msgf("%7s %s", route.Method, route.Path)
	}

	return e
}

func parseCliArgs() {
	var quiet, debug, trace bool

	flag.BoolVar(&cliArgs.version, "version", false, "Shows the application version, then exits.")
	flag.BoolVar(&cliArgs.initDB, "initdb", false, "Create the database; requires admin access to PostgreSQL.")
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
