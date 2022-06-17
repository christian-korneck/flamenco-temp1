package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func requestLogger(e echo.Context) zerolog.Logger {
	logger := zerolog.Ctx(e.Request().Context())
	return *logger
}
