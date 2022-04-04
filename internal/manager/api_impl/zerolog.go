package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func requestLogger(e echo.Context) zerolog.Logger {
	logger := log.Ctx(e.Request().Context())
	return *logger
}
