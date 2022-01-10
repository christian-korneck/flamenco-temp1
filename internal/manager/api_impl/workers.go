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
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"gitlab.com/blender/flamenco-goja-test/pkg/api"
)

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
		User:        "",
	})
}
