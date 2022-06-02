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
	worker2.StatusRequested = api.WorkerStatusAsleep

	mf.persistence.EXPECT().FetchWorkers(gomock.Any()).
		Return([]*persistence.Worker{&worker1, &worker2}, nil)

	echo := mf.prepareMockedRequest(nil)
	err := mf.flamenco.FetchWorkers(echo)
	assert.NoError(t, err)

	// Check the response
	workers := api.WorkerList{
		Workers: []api.WorkerSummary{
			{
				Id:       worker1.UUID,
				Nickname: worker1.Name,
				Status:   worker1.Status,
				Version:  worker1.Software,
			},
			{
				Id:       worker2.UUID,
				Nickname: worker2.Name,
				Status:   worker2.Status,
				Version:  worker2.Software,
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

	// Test with worker that doesn't have a status change requested.
	mf.persistence.EXPECT().FetchWorker(gomock.Any(), workerUUID).Return(&worker, nil)

	echo = mf.prepareMockedRequest(nil)
	err = mf.flamenco.FetchWorker(echo, workerUUID)
	assert.NoError(t, err)
	assertResponseJSON(t, echo, http.StatusOK, api.Worker{
		WorkerSummary: api.WorkerSummary{
			Id:       workerUUID,
			Nickname: "дрон",
			Version:  "3.0",
			Status:   api.WorkerStatusAwake,
		},
		IpAddress:          "fe80::5054:ff:fede:2ad7",
		Platform:           "linux",
		SupportedTaskTypes: []string{"blender", "ffmpeg", "file-management", "misc"},
	})

	// Test with worker that does have a status change requested.
	requestedStatus := api.WorkerStatusAsleep
	worker.StatusRequested = requestedStatus
	mf.persistence.EXPECT().FetchWorker(gomock.Any(), workerUUID).Return(&worker, nil)

	echo = mf.prepareMockedRequest(nil)
	err = mf.flamenco.FetchWorker(echo, worker.UUID)
	assert.NoError(t, err)
	assertResponseJSON(t, echo, http.StatusOK, api.Worker{
		WorkerSummary: api.WorkerSummary{
			Id:           workerUUID,
			Nickname:     "дрон",
			Version:      "3.0",
			Status:       api.WorkerStatusAwake,
			StatusChange: &api.WorkerStatusChangeRequest{Status: requestedStatus},
		},
		IpAddress:          "fe80::5054:ff:fede:2ad7",
		Platform:           "linux",
		SupportedTaskTypes: []string{"blender", "ffmpeg", "file-management", "misc"},
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
	savedWorker.StatusRequested = requestStatus
	savedWorker.LazyStatusRequest = true
	mf.persistence.EXPECT().SaveWorker(gomock.Any(), &savedWorker).Return(nil)

	// Expect a broadcast of the change
	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:       worker.UUID,
		Nickname: worker.Name,
		Status:   prevStatus,
		Updated:  worker.UpdatedAt,
		Version:  worker.Software,
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
	assertResponseEmpty(t, echo)
}
