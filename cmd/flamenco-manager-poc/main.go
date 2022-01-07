package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-goja-test/internal/appinfo"
	"gitlab.com/blender/flamenco-goja-test/internal/job_compilers"
	"gitlab.com/blender/flamenco-goja-test/pkg/api"
)

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting %v", appinfo.ApplicationName)

	gojaPoC()
	ginOpenAPIPoC()
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

// Proof of concept of a REST API with Gin and OpenAPI.
func ginOpenAPIPoC() {
	listen := ":8080"
	_, port, _ := net.SplitHostPort(listen)
	log.Info().Str("port", port).Msg("listening")

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	api.RegisterSwaggerUIStaticFiles(r)

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
	r.GET("/api/openapi3.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, swagger)
	})

	flamenco := api.NewFlamenco()
	r = api.RegisterHandlers(r, flamenco)

	finalErr := r.Run(listen)
	log.Warn().Err(finalErr).Msg("shutting down")
}
