package api_impl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/api_impl/mocks"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
	"gorm.io/gorm"
)

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

type mockedFlamenco struct {
	flamenco    *Flamenco
	jobCompiler *mocks.MockJobCompiler
	persistence *mocks.MockPersistenceService
}

func newMockedFlamenco(mockCtrl *gomock.Controller) mockedFlamenco {
	jc := mocks.NewMockJobCompiler(mockCtrl)
	ps := mocks.NewMockPersistenceService(mockCtrl)
	f := NewFlamenco(jc, ps)

	return mockedFlamenco{
		flamenco:    f,
		jobCompiler: jc,
		persistence: ps,
	}
}

func (mf *mockedFlamenco) prepareMockedJSONRequest(worker *persistence.Worker, requestBody interface{}) echo.Context {

	e := echo.New()

	bodyBytes, err := json.MarshalIndent(requestBody, "", "    ")
	if err != nil {
		panic(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(bodyBytes))
	req.Header.Add(echo.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	requestWorkerStore(c, worker)

	return c
}

func testWorker() persistence.Worker {
	return persistence.Worker{
		Model:              gorm.Model{ID: 1},
		UUID:               "e7632d62-c3b8-4af0-9e78-01752928952c",
		Name:               "дрон",
		Address:            "fe80::5054:ff:fede:2ad7",
		LastActivity:       "",
		Platform:           "linux",
		Software:           "3.0",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management,misc",
	}
}
