// Package api_impl implements the OpenAPI API from pkg/api/flamenco-manager.yaml.
package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"regexp"

	oapi_middle "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Regexp to find variables in `/path/:task_id` or `/path/:variable/subpath`.
var urlVariablesReplacer = regexp.MustCompile("/:([^/]+)(/?)")

// SwaggerValidator constructs the OpenAPI validator, which also handles authentication.
func SwaggerValidator(swagger *openapi3.T, persist PersistenceService) echo.MiddlewareFunc {
	options := oapi_middle.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, authInfo *openapi3filter.AuthenticationInput) error {
				return authenticator(ctx, authInfo, persist)
			},
		},

		// Skip OAPI validation when the request is not for the OAPI interface.
		Skipper: func(e echo.Context) bool {
			isOapi := isOpenAPIPath(swagger, e.Path())
			log.Trace().
				Bool("isOpenAPI", isOapi).
				Str("path", e.Path()).
				Msg("checking validation skipperoo")
			return !isOapi
		},
	}

	validator := oapi_middle.OapiRequestValidatorWithOptions(swagger, &options)
	return validator
}

// authenticator runs the appropriate authentication function given the security
// scheme name.
func authenticator(ctx context.Context, authInfo *openapi3filter.AuthenticationInput, persist PersistenceService) error {
	switch authInfo.SecuritySchemeName {
	case "worker_auth":
		return WorkerAuth(ctx, authInfo, persist)
	default:
		log.Warn().Str("scheme", authInfo.SecuritySchemeName).Msg("unknown security scheme")
		return errors.New("unknown security scheme")
	}
}

func isOpenAPIPath(swagger *openapi3.T, urlPath string) bool {
	oapiPath := replaceURLPathVariables(urlPath)
	found := swagger.Paths.Find(oapiPath) != nil
	return found
}

// replaceURLPathVariables replaces variable style (`:task_id` to `{task_id}`)
// to suit the swagger.Paths path list.
func replaceURLPathVariables(urlPath string) string {
	pathBytes := []byte(urlPath)
	replacement := []byte("/{$1}$2")
	oapiPathBytes := urlVariablesReplacer.ReplaceAll(pathBytes, replacement)
	return string(oapiPathBytes)
}
