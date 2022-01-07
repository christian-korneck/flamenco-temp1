package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

//go:embed swagger-ui
var swaggerUI embed.FS

func RegisterSwaggerUIStaticFiles(router *echo.Echo) {
	files, err := fs.Sub(swaggerUI, "swagger-ui")
	if err != nil {
		log.Fatal().Err(err).Msg("error preparing embedded files for serving over HTTP")
	}

	httpHandler := http.FileServer(http.FS(files))
	router.GET("/api/swagger-ui/*", echo.WrapHandler(http.StripPrefix("/api/swagger-ui/", httpHandler)))
}
