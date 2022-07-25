package swagger_ui

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

//go:embed static
var swaggerUI embed.FS

const swaggerURL = "/api/v3/swagger-ui/"

func RegisterSwaggerUIStaticFiles(router *echo.Echo) {
	files, err := fs.Sub(swaggerUI, "static")
	if err != nil {
		log.Fatal().Err(err).Msg("error preparing embedded files for serving over HTTP")
	}

	httpHandler := http.FileServer(http.FS(files))
	router.GET(swaggerURL+"*", echo.WrapHandler(http.StripPrefix(swaggerURL, httpHandler)))
}
