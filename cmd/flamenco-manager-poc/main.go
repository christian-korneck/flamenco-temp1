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
	"net"
	"net/http"
	"time"

	oapi_middle "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v3"

	"gitlab.com/blender/flamenco-goja-test/internal/appinfo"
	"gitlab.com/blender/flamenco-goja-test/internal/manager/api_impl"
	"gitlab.com/blender/flamenco-goja-test/internal/manager/job_compilers"
	"gitlab.com/blender/flamenco-goja-test/internal/manager/swagger_ui"
	"gitlab.com/blender/flamenco-goja-test/pkg/api"
)

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting %v", appinfo.ApplicationName)

	echoOpenAPIPoC()
}

// Proof of concept of job compiler in JavaScript.
func gojaPoC() {
	compiler, err := job_compilers.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading job compilers")
	}

	if err := compiler.Run("simple-blender-render"); err != nil {
		log.Fatal().Err(err).Msg("error running job compiler")
	}
}

// Proof of concept of a REST API with Echo and OpenAPI.
func echoOpenAPIPoC() {
	listen := ":8080"
	_, port, _ := net.SplitHostPort(listen)
	log.Info().Str("port", port).Msg("listening")

	e := echo.New()
	e.HideBanner = true
	e.Use(lecho.Middleware(lecho.Config{
		Logger: lecho.From(log.Logger),
	}))
	e.Use(middleware.Recover())

	swagger_ui.RegisterSwaggerUIStaticFiles(e)

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get swagger")
	}
	e.GET("/api/openapi3.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, swagger)
	})

	e.GET("/api/ping", func(c echo.Context) error {
		logger := log.Level(zerolog.InfoLevel)
		logger.Debug().Msg("debug debug")
		logger.Info().Msg("Info Info")

		return c.JSON(http.StatusOK, echo.Map{
			"message": "pong",
		})
	})

	validator := oapi_middle.OapiRequestValidatorWithOptions(swagger,
		&oapi_middle.Options{
			Options: openapi3filter.Options{
				AuthenticationFunc: authenticator,
			},

			// Skip OAPI validation when the request is not for the OAPI interface.
			Skipper: func(e echo.Context) bool {
				path := e.Path()
				skip := swagger.Paths.Find(path) == nil
				return skip
			},
		})
	e.Use(validator)

	compiler, err := job_compilers.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading job compilers")
	}

	flamenco := api_impl.NewFlamenco(compiler)
	api.RegisterHandlers(e, flamenco)

	// Log available routes
	routeLogger := log.Level(zerolog.DebugLevel)
	routeLogger.Debug().Msg("available routes:")
	for _, route := range e.Routes() {
		routeLogger.Debug().Msgf("%7s %s", route.Method, route.Path)
	}

	finalErr := e.Start(listen)
	log.Warn().Err(finalErr).Msg("shutting down")
}

func authenticator(ctx context.Context, authInfo *openapi3filter.AuthenticationInput) error {
	switch authInfo.SecuritySchemeName {
	case "worker_auth":
		return workerAuth(ctx, authInfo)
	default:
		log.Warn().Str("scheme", authInfo.SecuritySchemeName).Msg("unknown security scheme")
		return errors.New("unknown security scheme")
	}
}

func workerAuth(ctx context.Context, authInfo *openapi3filter.AuthenticationInput) error {
	echo := ctx.Value(oapi_middle.EchoContextKey).(echo.Context)
	req := echo.Request()
	u, p, ok := req.BasicAuth()

	// TODO: stop logging passwords.
	log.Debug().Interface("scheme", authInfo.SecuritySchemeName).Str("user", u).Str("password", p).Msg("authenticator")
	if !ok {
		return authInfo.NewError(errors.New("no auth header found"))
	}

	// TODO: check username/password against worker database.
	return nil
}
