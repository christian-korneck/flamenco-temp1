// Package api_impl implements the OpenAPI API from pkg/api/flamenco-manager.yaml.
package api_impl

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
