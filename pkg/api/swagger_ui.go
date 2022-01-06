package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

//go:embed swagger-ui
var swaggerUI embed.FS

func RegisterSwaggerUIStaticFiles(router *gin.Engine) {
	files, err := fs.Sub(swaggerUI, "swagger-ui")
	if err != nil {
		log.Fatal().Err(err).Msg("error preparing embedded files for serving over HTTP")
	}
	router.StaticFS("/api/swagger-ui/", http.FS(files))
}
