package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
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
				Id:              worker1.UUID,
				Nickname:        worker1.Name,
				Status:          worker1.Status,
				StatusRequested: nil,
			},
			{
				Id:              worker2.UUID,
				Nickname:        worker2.Name,
				Status:          worker2.Status,
				StatusRequested: &worker2.StatusRequested,
			},
		},
	}
	assertResponseJSON(t, echo, http.StatusOK, workers)
	resp := getRecordedResponse(echo)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
