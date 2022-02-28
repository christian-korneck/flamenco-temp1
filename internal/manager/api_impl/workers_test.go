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
	"context"
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

	echo := mf.prepareMockedRequest(&worker, nil)

	// Expect a call into the persistence layer, which should return a scheduled task.
	job := persistence.Job{
		UUID: "583a7d59-887a-4c6c-b3e4-a753018f71b0",
	}
	task := persistence.Task{
		UUID: "4107c7aa-e86d-4244-858b-6c4fce2af503",
		Job:  &job,
	}
	mf.persistence.EXPECT().ScheduleTask(echo.Request().Context(), &worker).Return(&task, nil)

	err := mf.flamenco.ScheduleTask(echo)
	assert.NoError(t, err)

	resp := echo.Response().Writer.(*httptest.ResponseRecorder)
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

func TestWorkerSignoffTaskRequeue(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	job := persistence.Job{
		UUID: "583a7d59-887a-4c6c-b3e4-a753018f71b0",
	}
	// Mock that the worker has two active tasks. It shouldn't happen, but even
	// when it does, both should be requeued when the worker signs off.
	task1 := persistence.Task{
		UUID:   "4107c7aa-e86d-4244-858b-6c4fce2af503",
		Job:    &job,
		Status: api.TaskStatusActive,
	}
	task2 := persistence.Task{
		UUID:   "beb3f39b-57a5-44bf-a0ad-533e3513a0b6",
		Job:    &job,
		Status: api.TaskStatusActive,
	}
	workerTasks := []*persistence.Task{&task1, &task2}

	// Signing off should be handled completely, even when the HTTP connection
	// breaks. This means using a different context than the one passed by Echo.
	echo := mf.prepareMockedRequest(&worker, nil)
	expectCtx := gomock.Not(gomock.Eq(echo.Request().Context()))

	// Expect worker's tasks to be re-queued.
	mf.persistence.EXPECT().
		FetchTasksOfWorkerInStatus(expectCtx, &worker, api.TaskStatusActive).
		Return(workerTasks, nil)
	mf.stateMachine.EXPECT().TaskStatusChange(expectCtx, &task1, api.TaskStatusQueued)
	mf.stateMachine.EXPECT().TaskStatusChange(expectCtx, &task2, api.TaskStatusQueued)

	// Expect worker to be saved as 'offline'.
	mf.persistence.EXPECT().
		SaveWorkerStatus(expectCtx, &worker).
		Do(func(ctx context.Context, w *persistence.Worker) error {
			assert.Equal(t, api.WorkerStatusOffline, w.Status)
			return nil
		})

	err := mf.flamenco.SignOff(echo)
	assert.NoError(t, err)

	resp := echo.Response().Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusNoContent, resp.Code)
}
