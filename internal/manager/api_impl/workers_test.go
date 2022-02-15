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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

func TestTaskScheduleHappy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	// Expect a call into the persistence layer, which should return a scheduled task.
	job := persistence.Job{
		UUID: "583a7d59-887a-4c6c-b3e4-a753018f71b0",
	}
	task := persistence.Task{
		UUID: "4107c7aa-e86d-4244-858b-6c4fce2af503",
		Job:  &job,
	}
	mf.persistence.EXPECT().ScheduleTask(&worker).Return(&task, nil)

	echoCtx := mf.prepareMockedRequest(&worker, nil)
	err := mf.flamenco.ScheduleTask(echoCtx)
	assert.NoError(t, err)

	resp := echoCtx.Response().Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusOK, resp.Code)
	// TODO: check that the returned JSON actually matches what we expect.
}

func TestTaskScheduleNonActiveStatus(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusAsleep

	// Explicitly NO expected calls to the persistence layer. Since the worker is
	// not in a state that allows task execution, there should be no DB queries.

	echoCtx := mf.prepareMockedRequest(&worker, nil)
	err := mf.flamenco.ScheduleTask(echoCtx)
	assert.NoError(t, err)

	resp := echoCtx.Response().Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestTaskScheduleOtherStatusRequested(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.StatusRequested = api.WorkerStatusAsleep

	// Explicitly NO expected calls to the persistence layer. Since the worker is
	// not in a state that allows task execution, there should be no DB queries.

	echoCtx := mf.prepareMockedRequest(&worker, nil)
	err := mf.flamenco.ScheduleTask(echoCtx)
	assert.NoError(t, err)

	resp := echoCtx.Response().Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusLocked, resp.Code)

	responseBody := api.WorkerStateChange{}
	err = json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, worker.StatusRequested, responseBody.StatusRequested)
}
