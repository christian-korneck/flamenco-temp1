package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"git.blender.org/flamenco/internal/manager/api_impl/mocks"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

type mockedFlamenco struct {
	flamenco     *Flamenco
	jobCompiler  *mocks.MockJobCompiler
	persistence  *mocks.MockPersistenceService
	broadcaster  *mocks.MockChangeBroadcaster
	logStorage   *mocks.MockLogStorage
	config       *mocks.MockConfigService
	stateMachine *mocks.MockTaskStateMachine
	shaman       *mocks.MockShaman
}

func newMockedFlamenco(mockCtrl *gomock.Controller) mockedFlamenco {
	jc := mocks.NewMockJobCompiler(mockCtrl)
	ps := mocks.NewMockPersistenceService(mockCtrl)
	cb := mocks.NewMockChangeBroadcaster(mockCtrl)
	ls := mocks.NewMockLogStorage(mockCtrl)
	cs := mocks.NewMockConfigService(mockCtrl)
	sm := mocks.NewMockTaskStateMachine(mockCtrl)
	sha := mocks.NewMockShaman(mockCtrl)
	f := NewFlamenco(jc, ps, cb, ls, cs, sm, sha)

	return mockedFlamenco{
		flamenco:     f,
		jobCompiler:  jc,
		persistence:  ps,
		broadcaster:  cb,
		logStorage:   ls,
		config:       cs,
		stateMachine: sm,
	}
}

// prepareMockedJSONRequest returns an `echo.Context` that has a JSON request body attached to it.
func (mf *mockedFlamenco) prepareMockedJSONRequest(requestBody interface{}) echo.Context {
	bodyBytes, err := json.MarshalIndent(requestBody, "", "    ")
	if err != nil {
		panic(err)
	}

	c := mf.prepareMockedRequest(bytes.NewBuffer(bodyBytes))
	c.Request().Header.Add(echo.HeaderContentType, "application/json")

	return c
}

// prepareMockedJSONRequest returns an `echo.Context` that has an empty request body attached to it.
func (mf *mockedFlamenco) prepareMockedRequest(body io.Reader) echo.Context {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

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
