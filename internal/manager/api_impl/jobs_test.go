package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"testing"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](value T) *T {
	return &value
}

func TestTaskUpdate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	// Construct the JSON request object.
	taskUpdate := api.TaskUpdateJSONRequestBody{
		Activity:   ptr("testing"),
		Log:        ptr("line1\nline2\n"),
		TaskStatus: ptr(api.TaskStatusFailed),
	}

	// Construct the task that's supposed to be updated.
	taskID := "181eab68-1123-4790-93b1-94309a899411"
	jobID := "e4719398-7cfa-4877-9bab-97c2d6c158b5"
	mockJob := persistence.Job{UUID: jobID}
	mockTask := persistence.Task{
		UUID:     taskID,
		Worker:   &worker,
		WorkerID: &worker.ID,
		Job:      &mockJob,
		Activity: "pre-update activity",
	}

	// Expect the task to be fetched.
	mf.persistence.EXPECT().FetchTask(gomock.Any(), taskID).Return(&mockTask, nil)

	// Expect the task status change to be handed to the state machine.
	var statusChangedtask persistence.Task
	mf.stateMachine.EXPECT().TaskStatusChange(gomock.Any(), gomock.AssignableToTypeOf(&persistence.Task{}), api.TaskStatusFailed).
		DoAndReturn(func(ctx context.Context, task *persistence.Task, newStatus api.TaskStatus) error {
			statusChangedtask = *task
			return nil
		})

	// Expect the activity to be updated.
	var actUpdatedTask persistence.Task
	mf.persistence.EXPECT().SaveTaskActivity(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task *persistence.Task) error {
			actUpdatedTask = *task
			return nil
		})

	// Expect the log to be written.
	mf.logStorage.EXPECT().Write(gomock.Any(), jobID, taskID, "line1\nline2\n")

	// Do the call.
	echoCtx := mf.prepareMockedJSONRequest(&worker, taskUpdate)
	err := mf.flamenco.TaskUpdate(echoCtx, taskID)

	// Check the saved task.
	assert.NoError(t, err)
	assert.Equal(t, mockTask.UUID, statusChangedtask.UUID)
	assert.Equal(t, mockTask.UUID, actUpdatedTask.UUID)
	assert.Equal(t, "pre-update activity", statusChangedtask.Activity) // the 'save' should come from the change in status.
	assert.Equal(t, "testing", actUpdatedTask.Activity)                // the activity should be saved separately.
}
