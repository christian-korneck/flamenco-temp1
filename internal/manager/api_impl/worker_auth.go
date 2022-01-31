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

	oapi_middle "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type workerContextKey string

const (
	workerKey = workerContextKey("worker")
)

var (
	errAuthBad = errors.New("no such worker known")
)

// OpenAPI authentication function for authing workers.
// The worker will be fetched from the database and stored in the request context.
func WorkerAuth(ctx context.Context, authInfo *openapi3filter.AuthenticationInput, persist PersistenceService) error {
	echo := ctx.Value(oapi_middle.EchoContextKey).(echo.Context)
	req := echo.Request()
	logger := requestLogger(echo)

	// Fetch username & password from the HTTP header.
	u, p, ok := req.BasicAuth()
	logger.Debug().Interface("scheme", authInfo.SecuritySchemeName).Str("user", u).Msg("authenticator")
	if !ok {
		return authInfo.NewError(errors.New("no auth header found"))
	}

	// Fetch the Worker that has this username, making sure there is always _some_
	// secret to check. This helps in making this a constant-time operation.
	var hashedSecret string
	w, err := persist.FetchWorker(ctx, u)
	if err == nil {
		hashedSecret = w.Secret
	} else {
		hashedSecret = "this is not a BCrypt hash, so it'll fail"
	}

	// Check the password.
	err = bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(p))
	if err != nil {
		logger.Warn().Str("username", u).Msg("authentication error")
		return authInfo.NewError(errAuthBad)
	}

	// Store the Worker in the request context, so that it doesn't need to be fetched again later.
	reqCtx := context.WithValue(req.Context(), workerKey, w)
	echo.SetRequest(req.WithContext(reqCtx))

	return nil
}
