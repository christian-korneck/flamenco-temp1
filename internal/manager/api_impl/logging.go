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

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type loggerContextKey string

const (
	loggerKey = loggerContextKey("logger")
)

// MiddleWareRequestLogger is Echo middleware that puts a Zerolog logger in the request context, for endpoints to use.
func MiddleWareRequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		remoteIP := c.RealIP()
		logger := log.With().Str("remoteAddr", remoteIP).Logger()
		ctx := context.WithValue(c.Request().Context(), loggerKey, logger)
		c.SetRequest(c.Request().WithContext(ctx))

		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

func requestLogger(e echo.Context) zerolog.Logger {
	ctx := e.Request().Context()
	logger, ok := ctx.Value(loggerKey).(zerolog.Logger)
	if ok {
		return logger
	}

	log.Error().Msg("no logger found in request context, returning default logger")
	return log.With().Logger()
}
