package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

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
	clock        *clock.Mock
	lastRender   *mocks.MockLastRendered
	localStorage *mocks.MockLocalStorage

	// Place for some tests to store a temporary directory.
	tempdir string
}

func newMockedFlamenco(mockCtrl *gomock.Controller) mockedFlamenco {
	jc := mocks.NewMockJobCompiler(mockCtrl)
	ps := mocks.NewMockPersistenceService(mockCtrl)
	cb := mocks.NewMockChangeBroadcaster(mockCtrl)
	logStore := mocks.NewMockLogStorage(mockCtrl)
	cs := mocks.NewMockConfigService(mockCtrl)
	sm := mocks.NewMockTaskStateMachine(mockCtrl)
	sha := mocks.NewMockShaman(mockCtrl)
	lr := mocks.NewMockLastRendered(mockCtrl)
	localStore := mocks.NewMockLocalStorage(mockCtrl)

	clock := clock.NewMock()
	mockedNow, err := time.Parse(time.RFC3339, "2022-06-09T11:14:41+02:00")
	if err != nil {
		panic(err)
	}
	clock.Set(mockedNow)

	f := NewFlamenco(jc, ps, cb, logStore, cs, sm, sha, clock, lr, localStore)

	return mockedFlamenco{
		flamenco:     f,
		jobCompiler:  jc,
		persistence:  ps,
		broadcaster:  cb,
		logStorage:   logStore,
		config:       cs,
		stateMachine: sm,
		clock:        clock,
		lastRender:   lr,
		localStorage: localStore,
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
// `body` may be `nil` to indicate "no body".
func (mf *mockedFlamenco) prepareMockedRequest(body io.Reader) echo.Context {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c
}

func getRecordedResponseRecorder(echoCtx echo.Context) *httptest.ResponseRecorder {
	writer := echoCtx.Response().Writer
	resp, ok := writer.(*httptest.ResponseRecorder)
	if !ok {
		panic(fmt.Sprintf("response writer was not a `*httptest.ResponseRecorder` but a %T", writer))
	}
	return resp
}

func getRecordedResponse(echoCtx echo.Context) *http.Response {
	return getRecordedResponseRecorder(echoCtx).Result()
}

func getResponseJSON(t *testing.T, echoCtx echo.Context, expectStatusCode int, actualPayloadPtr interface{}) {
	resp := getRecordedResponse(echoCtx)
	assert.Equal(t, expectStatusCode, resp.StatusCode)
	contentType := resp.Header.Get(echo.HeaderContentType)

	if !assert.Equal(t, "application/json; charset=UTF-8", contentType) {
		t.Fatalf("response not JSON but %q, not going to compare body", contentType)
		return
	}

	actualJSON, err := io.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = json.Unmarshal(actualJSON, actualPayloadPtr)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

// assertResponseJSON asserts that a recorded response is JSON with the given HTTP status code.
func assertResponseJSON(t *testing.T, echoCtx echo.Context, expectStatusCode int, expectBody interface{}) {
	resp := getRecordedResponse(echoCtx)
	assert.Equal(t, expectStatusCode, resp.StatusCode)
	contentType := resp.Header.Get(echo.HeaderContentType)

	if !assert.Equal(t, "application/json; charset=UTF-8", contentType) {
		t.Fatalf("response not JSON but %q, not going to compare body", contentType)
		return
	}

	expectJSON, err := json.Marshal(expectBody)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	actualJSON, err := io.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.JSONEq(t, string(expectJSON), string(actualJSON))
}

func assertResponseAPIError(t *testing.T, echoCtx echo.Context, expectStatusCode int, expectMessage string, fmtArgs ...interface{}) {
	if len(fmtArgs) > 0 {
		expectMessage = fmt.Sprintf(expectMessage, fmtArgs...)
	}

	assertResponseJSON(t, echoCtx, expectStatusCode, api.Error{
		Code:    int32(expectStatusCode),
		Message: expectMessage,
	})
}

// assertResponseNoContent asserts the response has no body and the given
func assertResponseNoContent(t *testing.T, echoCtx echo.Context) {
	resp := getRecordedResponseRecorder(echoCtx)
	assert.Equal(t, http.StatusNoContent, resp.Code, "Unexpected status: %v", resp.Result().Status)
	assert.Zero(t, resp.Body.Len(), "HTTP 204 No Content should have no content, got %v", resp.Body.String())
}

// assertResponseNoBody asserts the response has no body and the given status.
func assertResponseNoBody(t *testing.T, echoCtx echo.Context, expectStatus int) {
	resp := getRecordedResponseRecorder(echoCtx)
	assert.Equal(t, expectStatus, resp.Code, "Unexpected status: %v", resp.Result().Status)
	assert.Zero(t, resp.Body.Len(), "HTTP response have no content, got %v", resp.Body.String())
}

func testWorker() persistence.Worker {
	return persistence.Worker{
		Model:              persistence.Model{ID: 1},
		UUID:               "e7632d62-c3b8-4af0-9e78-01752928952c",
		Name:               "дрон",
		Address:            "fe80::5054:ff:fede:2ad7",
		Platform:           "linux",
		Software:           "3.0",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management,misc",
	}
}
