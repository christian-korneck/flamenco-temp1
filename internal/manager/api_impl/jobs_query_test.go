// SPDX-License-Identifier: GPL-3.0-or-later
package api_impl

import (
	"net/http"
	"testing"
	"time"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFetchTask(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)

	taskUUID := "19b62e32-564f-43a3-84fb-06e80ad36f16"
	workerUUID := "b5725bb3-d540-4070-a2b6-7b4b26925f94"
	jobUUID := "8b179118-0189-478a-b463-73798409898c"

	taskWorker := persistence.Worker{UUID: workerUUID, Name: "Radnik", Address: "Slapić"}

	dbTask := persistence.Task{
		Model: persistence.Model{
			ID:        327,
			CreatedAt: mf.clock.Now().Add(-30 * time.Second),
			UpdatedAt: mf.clock.Now(),
		},
		UUID:         taskUUID,
		Name:         "симпатичная задача",
		Type:         "misc",
		JobID:        0,
		Job:          &persistence.Job{UUID: jobUUID},
		Priority:     47,
		Status:       api.TaskStatusQueued,
		WorkerID:     new(uint),
		Worker:       &taskWorker,
		Dependencies: []*persistence.Task{},
		Activity:     "used in unit test",

		Commands: []persistence.Command{
			{Name: "move-directory",
				Parameters: map[string]interface{}{
					"dest": "/render/_flamenco/tests/renders/2022-04-29 Weekly/2022-04-29_140531",
					"src":  "/render/_flamenco/tests/renders/2022-04-29 Weekly/2022-04-29_140531__intermediate-2022-04-29_140531",
				}},
		},
	}

	expectAPITask := api.Task{
		Activity: "used in unit test",
		Created:  dbTask.CreatedAt,
		Id:       taskUUID,
		JobId:    jobUUID,
		Name:     "симпатичная задача",
		Priority: 47,
		Status:   api.TaskStatusQueued,
		TaskType: "misc",
		Updated:  dbTask.UpdatedAt,
		Worker:   &api.TaskWorker{Id: workerUUID, Name: "Radnik", Address: "Slapić"},

		Commands: []api.Command{
			{Name: "move-directory",
				Parameters: map[string]interface{}{
					"dest": "/render/_flamenco/tests/renders/2022-04-29 Weekly/2022-04-29_140531",
					"src":  "/render/_flamenco/tests/renders/2022-04-29 Weekly/2022-04-29_140531__intermediate-2022-04-29_140531",
				}},
		},

		FailedByWorkers: ptr([]api.TaskWorker{
			{Id: workerUUID, Name: "Radnik", Address: "Slapić"},
		}),
	}

	echoCtx := mf.prepareMockedRequest(nil)
	ctx := echoCtx.Request().Context()
	mf.persistence.EXPECT().FetchTask(ctx, taskUUID).Return(&dbTask, nil)
	mf.persistence.EXPECT().FetchTaskFailureList(ctx, &dbTask).
		Return([]*persistence.Worker{&taskWorker}, nil)

	err := mf.flamenco.FetchTask(echoCtx, taskUUID)
	assert.NoError(t, err)

	assertResponseJSON(t, echoCtx, http.StatusOK, expectAPITask)
}
