package api_impl

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-goja-test/pkg/api"
)

type Flamenco struct {
}

var _ api.ServerInterface = (*Flamenco)(nil)

func NewFlamenco() *Flamenco {
	return &Flamenco{}
}

func (f *Flamenco) RegisterWorker(e echo.Context) error {
	remoteIP := e.RealIP()

	logger := log.With().
		Str("ip", remoteIP).
		Logger()

	var req api.RegisterWorkerJSONBody
	err := e.Bind(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	logger.Info().Str("nickname", req.Nickname).Msg("registering new worker")

	return e.JSON(http.StatusOK, &api.RegisteredWorker{
		Id:       uuid.New().String(),
		Nickname: req.Nickname,
		Platform: req.Platform,
		Address:  remoteIP,
	})
}

func (f *Flamenco) ScheduleTask(e echo.Context) error {
	return e.JSON(http.StatusOK, &api.AssignedTask{
		Id: uuid.New().String(),
		Commands: []api.Command{
			{"echo", echo.Map{"payload": "Simon says \"Shaders!\""}},
			{"blender", echo.Map{"blender_cmd": "/shared/bin/blender"}},
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
func sendAPIError(e echo.Context, code int, message string) error {
	petErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	return e.JSON(code, petErr)
}
