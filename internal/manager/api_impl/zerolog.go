package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func requestLogger(e echo.Context) zerolog.Logger {
	logCtx := log.With().
		Str("remoteAddr", e.RealIP()).
		Str("userAgent", e.Request().UserAgent())

	worker := requestWorker(e)
	if worker != nil {
		logCtx = logCtx.
			Str("wUUID", worker.UUID).
			Str("wName", worker.Name)
	}

	return logCtx.Logger()
}
