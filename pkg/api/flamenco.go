//go:generate oapi-codegen -generate types -o openapi_types.gen.go -package api flamenco-manager.yaml
//go:generate oapi-codegen -generate gin   -o openapi_gin.gen.go   -package api flamenco-manager.yaml
//go:generate oapi-codegen -generate spec  -o openapi_spec.gen.go  -package api flamenco-manager.yaml

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Flamenco struct {
}

var _ ServerInterface = (*Flamenco)(nil)

func NewFlamenco() *Flamenco {
	return &Flamenco{}
}

func (f *Flamenco) RegisterWorker(c *gin.Context) {
	remoteIP, isTrustedProxy := c.RemoteIP()

	logger := log.With().
		Str("ip", remoteIP.String()).
		Bool("trustedProxy", isTrustedProxy).
		Logger()

	var req RegisterWorkerJSONBody
	err := c.Bind(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		sendAPIError(c, http.StatusBadRequest, "invalid format")
		return
	}

	logger.Info().Str("nickname", req.Nickname).Msg("registering new worker")

	c.JSON(http.StatusOK, &RegisteredWorker{
		Id:       uuid.New().String(),
		Nickname: req.Nickname,
		Platform: req.Platform,
		Address:  remoteIP.String(),
	})
}

func (f *Flamenco) PostTask(c *gin.Context) {
	c.JSON(http.StatusOK, &AssignedTask{
		Id: uuid.New().String(),
		Commands: []Command{
			{"echo", gin.H{"payload": "Simon says \"Shaders!\""}},
			{"blender", gin.H{"blender_cmd": "/shared/bin/blender"}},
		},
		Job:         uuid.New().String(),
		JobPriority: 50,
		JobType:     "blender-render",
		Name:        "A1032",
		Priority:    50,
		Status:      "active",
		TaskType:    "blender-render",
	})
}

// sendPetstoreError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendAPIError(c *gin.Context, code int, message string) {
	petErr := Error{
		Code:    int32(code),
		Message: message,
	}
	c.JSON(code, petErr)
}
