package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

func TestFetchWorkers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker1 := testWorker()
	worker2 := worker1
	worker2.ID = 4
	worker2.UUID = "f07b6d53-16ec-40a8-a7b4-a9cc8547f790"
	worker2.Status = api.WorkerStatusAwake
	worker2.StatusChangeRequest(api.WorkerStatusAsleep, false)

	mf.persistence.EXPECT().FetchWorkers(gomock.Any()).
		Return([]*persistence.Worker{&worker1, &worker2}, nil)

	echo := mf.prepareMockedRequest(nil)
	err := mf.flamenco.FetchWorkers(echo)
	assert.NoError(t, err)

	// Check the response
	workers := api.WorkerList{
		Workers: []api.WorkerSummary{
			{
				Id:      worker1.UUID,
				Name:    worker1.Name,
				Status:  worker1.Status,
				Version: worker1.Software,
			},
			{
				Id:      worker2.UUID,
				Name:    worker2.Name,
				Status:  worker2.Status,
				Version: worker2.Software,
				StatusChange: &api.WorkerStatusChangeRequest{
					Status: worker2.StatusRequested,
					IsLazy: false,
				},
			},
		},
	}
	assertResponseJSON(t, echo, http.StatusOK, workers)
	resp := getRecordedResponse(echo)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestFetchWorker(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	workerUUID := worker.UUID

	// Test without worker in the database.
	mf.persistence.EXPECT().FetchWorker(gomock.Any(), workerUUID).
		Return(nil, fmt.Errorf("wrapped: %w", persistence.ErrWorkerNotFound))
	echo := mf.prepareMockedRequest(nil)
	err := mf.flamenco.FetchWorker(echo, workerUUID)
	assert.NoError(t, err)
	assertResponseAPIError(t, echo, http.StatusNotFound, fmt.Sprintf("worker %q not found", workerUUID))

	// Test database error fetching worker.
	mf.persistence.EXPECT().FetchWorker(gomock.Any(), workerUUID).
		Return(nil, errors.New("some unknown error"))
	echo = mf.prepareMockedRequest(nil)
	err = mf.flamenco.FetchWorker(echo, workerUUID)
	assert.NoError(t, err)
	assertResponseAPIError(t, echo, http.StatusInternalServerError, "error fetching worker: some unknown error")

	// Test with worker that does NOT have a status change requested, and DOES have an assigned task.
	mf.persistence.EXPECT().FetchWorker(gomock.Any(), workerUUID).Return(&worker, nil)
	assignedTask := persistence.Task{
		UUID:   "806057d5-759a-4e75-86a4-356d43f28cff",
		Name:   "test task",
		Job:    &persistence.Job{UUID: "f0e25ee4-0d13-4291-afc3-e9446b555aaf"},
		Status: api.TaskStatusActive,
	}
	mf.persistence.EXPECT().FetchWorkerTask(gomock.Any(), &worker).Return(&assignedTask, nil)

	echo = mf.prepareMockedRequest(nil)
	err = mf.flamenco.FetchWorker(echo, workerUUID)
	assert.NoError(t, err)
	assertResponseJSON(t, echo, http.StatusOK, api.Worker{
		WorkerSummary: api.WorkerSummary{
			Id:      workerUUID,
			Name:    "дрон",
			Version: "3.0",
			Status:  api.WorkerStatusAwake,
		},
		IpAddress:          "fe80::5054:ff:fede:2ad7",
		Platform:           "linux",
		SupportedTaskTypes: []string{"blender", "ffmpeg", "file-management", "misc"},
		Task: &api.WorkerTask{
			TaskSummary: api.TaskSummary{
				Id:     assignedTask.UUID,
				Name:   assignedTask.Name,
				Status: assignedTask.Status,
			},
			JobId: assignedTask.Job.UUID,
		},
	})

	// Test with worker that does have a status change requested, but does NOT Have an assigned task.
	requestedStatus := api.WorkerStatusAsleep
	worker.StatusChangeRequest(requestedStatus, false)
	mf.persistence.EXPECT().FetchWorker(gomock.Any(), workerUUID).Return(&worker, nil)
	mf.persistence.EXPECT().FetchWorkerTask(gomock.Any(), &worker).Return(nil, nil)

	echo = mf.prepareMockedRequest(nil)
	err = mf.flamenco.FetchWorker(echo, worker.UUID)
	assert.NoError(t, err)
	assertResponseJSON(t, echo, http.StatusOK, api.Worker{
		WorkerSummary: api.WorkerSummary{
			Id:           workerUUID,
			Name:         "дрон",
			Version:      "3.0",
			Status:       api.WorkerStatusAwake,
			StatusChange: &api.WorkerStatusChangeRequest{Status: requestedStatus},
		},
		IpAddress:          "fe80::5054:ff:fede:2ad7",
		Platform:           "linux",
		SupportedTaskTypes: []string{"blender", "ffmpeg", "file-management", "misc"},
		Task:               nil,
	})
}

func TestRequestWorkerStatusChange(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	workerUUID := worker.UUID
	prevStatus := worker.Status

	mf.persistence.EXPECT().FetchWorker(gomock.Any(), workerUUID).Return(&worker, nil)

	requestStatus := api.WorkerStatusAsleep
	savedWorker := worker
	savedWorker.StatusChangeRequest(requestStatus, true)
	mf.persistence.EXPECT().SaveWorker(gomock.Any(), &savedWorker).Return(nil)

	// Expect a broadcast of the change
	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:      worker.UUID,
		Name:    worker.Name,
		Status:  prevStatus,
		Updated: worker.UpdatedAt,
		Version: worker.Software,
		StatusChange: &api.WorkerStatusChangeRequest{
			Status: requestStatus,
			IsLazy: true,
		},
	})

	echo := mf.prepareMockedJSONRequest(api.WorkerStatusChangeRequest{
		Status: requestStatus,
		IsLazy: true,
	})
	err := mf.flamenco.RequestWorkerStatusChange(echo, workerUUID)
	assert.NoError(t, err)
	assertResponseNoContent(t, echo)
}

func TestRequestWorkerStatusChangeRevert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	// Mimick that a status change request to 'asleep' was already performed.
	worker.StatusChangeRequest(api.WorkerStatusAsleep, true)

	workerUUID := worker.UUID
	currentStatus := worker.Status

	mf.persistence.EXPECT().FetchWorker(gomock.Any(), workerUUID).Return(&worker, nil)

	// Perform a request to go to the current worker status. This should cancel
	// the previous status change request.
	requestStatus := currentStatus
	savedWorker := worker
	savedWorker.StatusChangeClear()
	mf.persistence.EXPECT().SaveWorker(gomock.Any(), &savedWorker).Return(nil)

	// Expect a broadcast of the change
	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:           worker.UUID,
		Name:         worker.Name,
		Status:       currentStatus,
		Updated:      worker.UpdatedAt,
		Version:      worker.Software,
		StatusChange: nil,
	})

	echo := mf.prepareMockedJSONRequest(api.WorkerStatusChangeRequest{
		Status: requestStatus,

		// This shouldn't matter; requesting the current status should simply erase
		// the previous status change request.
		IsLazy: true,
	})
	err := mf.flamenco.RequestWorkerStatusChange(echo, workerUUID)
	assert.NoError(t, err)
	assertResponseNoContent(t, echo)
}
