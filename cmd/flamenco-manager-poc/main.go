package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v3"

	"gitlab.com/blender/flamenco-goja-test/internal/appinfo"
	"gitlab.com/blender/flamenco-goja-test/internal/job_compilers"
	"gitlab.com/blender/flamenco-goja-test/pkg/api"
)

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting %v", appinfo.ApplicationName)

	gojaPoC()
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
	e.Use(lecho.Middleware(lecho.Config{
		Logger: lecho.From(log.Logger),
	}))
	e.Use(middleware.Recover())

	e.GET("/ping", func(c echo.Context) error {
		logger := log.Level(zerolog.InfoLevel)
		logger.Debug().Msg("debug debug")
		logger.Info().Msg("Info Info")

		return c.JSON(http.StatusOK, echo.Map{
			"message": "pong",
		})
	})

	api.RegisterSwaggerUIStaticFiles(e)

	// Adjust the OpenAPI3/Swagger spec to match the port we're listening on.
	swagger, err := api.GetSwagger()
	swagger.Servers = []*openapi3.Server{
		{
			URL: fmt.Sprintf("http://0.0.0.0:%s/", port),
		},
	}
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get swagger")
	}
	e.GET("/api/openapi3.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, swagger)
	})

	flamenco := api.NewFlamenco()
	api.RegisterHandlers(e, flamenco)

	finalErr := e.Start(listen)
	log.Warn().Err(finalErr).Msg("shutting down")
}
