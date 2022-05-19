package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

func TestTaskScheduleHappy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	echo := mf.prepareMockedRequest(nil)
	requestWorkerStore(echo, &worker)

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

	// Check the response
	assignedTask := api.AssignedTask{
		Uuid:     task.UUID,
		Job:      job.UUID,
		Commands: []api.Command{},
	}
	assertResponseJSON(t, echo, http.StatusOK, assignedTask)
	resp := getRecordedResponse(echo)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTaskScheduleNonActiveStatus(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusAsleep

	// Explicitly NO expected calls to the persistence layer. Since the worker is
	// not in a state that allows task execution, there should be no DB queries.

	echoCtx := mf.prepareMockedRequest(nil)
	requestWorkerStore(echoCtx, &worker)
	err := mf.flamenco.ScheduleTask(echoCtx)
	assert.NoError(t, err)

	resp := getRecordedResponse(echoCtx)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestTaskScheduleOtherStatusRequested(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.StatusRequested = api.WorkerStatusAsleep

	// Explicitly NO expected calls to the persistence layer. Since the worker is
	// not in a state that allows task execution, there should be no DB queries.

	echoCtx := mf.prepareMockedRequest(nil)
	requestWorkerStore(echoCtx, &worker)
	err := mf.flamenco.ScheduleTask(echoCtx)
	assert.NoError(t, err)

	expectBody := api.WorkerStateChange{StatusRequested: api.WorkerStatusAsleep}
	assertResponseJSON(t, echoCtx, http.StatusLocked, expectBody)
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
	echo := mf.prepareMockedRequest(nil)
	requestWorkerStore(echo, &worker)
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

	resp := getRecordedResponse(echo)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
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
	echoCtx := mf.prepareMockedJSONRequest(taskUpdate)
	requestWorkerStore(echoCtx, &worker)
	err := mf.flamenco.TaskUpdate(echoCtx, taskID)

	// Check the saved task.
	assert.NoError(t, err)
	assert.Equal(t, mockTask.UUID, statusChangedtask.UUID)
	assert.Equal(t, mockTask.UUID, actUpdatedTask.UUID)
	assert.Equal(t, "pre-update activity", statusChangedtask.Activity) // the 'save' should come from the change in status.
	assert.Equal(t, "testing", actUpdatedTask.Activity)                // the activity should be saved separately.
}

func TestMayWorkerRun(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	prepareRequest := func() echo.Context {
		echo := mf.prepareMockedRequest(nil)
		requestWorkerStore(echo, &worker)
		return echo
	}

	job := persistence.Job{
		UUID: "583a7d59-887a-4c6c-b3e4-a753018f71b0",
	}

	task := persistence.Task{
		UUID:   "4107c7aa-e86d-4244-858b-6c4fce2af503",
		Job:    &job,
		Status: api.TaskStatusActive,
	}

	mf.persistence.EXPECT().FetchTask(gomock.Any(), task.UUID).Return(&task, nil).AnyTimes()

	// Test: unhappy, task unassigned
	{
		echo := prepareRequest()
		err := mf.flamenco.MayWorkerRun(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseJSON(t, echo, http.StatusOK, api.MayKeepRunning{
			MayKeepRunning: false,
			Reason:         "task not assigned to this worker",
		})
	}

	// Test: happy, task assigned to this worker.
	{
		echo := prepareRequest()
		task.WorkerID = &worker.ID
		err := mf.flamenco.MayWorkerRun(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseJSON(t, echo, http.StatusOK, api.MayKeepRunning{
			MayKeepRunning: true,
		})
	}

	// Test: unhappy, assigned but cancelled.
	{
		echo := prepareRequest()
		task.WorkerID = &worker.ID
		task.Status = api.TaskStatusCanceled
		err := mf.flamenco.MayWorkerRun(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseJSON(t, echo, http.StatusOK, api.MayKeepRunning{
			MayKeepRunning: false,
			Reason:         "task is in non-runnable status \"canceled\"",
		})
	}

	// Test: unhappy, assigned and runnable but worker should go to bed.
	{
		worker.StatusRequested = api.WorkerStatusAsleep
		echo := prepareRequest()
		task.WorkerID = &worker.ID
		task.Status = api.TaskStatusActive
		err := mf.flamenco.MayWorkerRun(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseJSON(t, echo, http.StatusOK, api.MayKeepRunning{
			MayKeepRunning:        false,
			Reason:                "worker status change requested",
			StatusChangeRequested: true,
		})
	}

}
