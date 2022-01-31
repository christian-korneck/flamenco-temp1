package api_impl

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
