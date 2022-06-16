package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/config"
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

	ctx := echo.Request().Context()
	mf.persistence.EXPECT().ScheduleTask(ctx, &worker).Return(&task, nil)
	mf.persistence.EXPECT().TaskTouchedByWorker(ctx, &task)
	mf.persistence.EXPECT().WorkerSeen(ctx, &worker)

	mf.logStorage.EXPECT().WriteTimestamped(gomock.Any(), job.UUID, task.UUID,
		"Task assigned to worker дрон (e7632d62-c3b8-4af0-9e78-01752928952c)")

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

func TestTaskScheduleNoTaskAvailable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	echo := mf.prepareMockedRequest(nil)
	requestWorkerStore(echo, &worker)

	// Expect a call into the persistence layer, which should return nil.
	ctx := echo.Request().Context()
	mf.persistence.EXPECT().ScheduleTask(ctx, &worker).Return(nil, nil)

	// This call should still trigger a "worker seen" call, as the worker is
	// actively asking for tasks.
	mf.persistence.EXPECT().WorkerSeen(ctx, &worker)

	err := mf.flamenco.ScheduleTask(echo)
	assert.NoError(t, err)
	assertResponseEmpty(t, echo)
}

func TestTaskScheduleNonActiveStatus(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusAsleep

	echoCtx := mf.prepareMockedRequest(nil)
	requestWorkerStore(echoCtx, &worker)

	// The worker should be marked as 'seen', even when it's in a state that
	// doesn't allow task execution.
	mf.persistence.EXPECT().WorkerSeen(echoCtx.Request().Context(), &worker)

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
	worker.StatusChangeRequest(api.WorkerStatusAsleep, false)

	echoCtx := mf.prepareMockedRequest(nil)
	requestWorkerStore(echoCtx, &worker)

	// The worker should be marked as 'seen', even when it's in a state that
	// doesn't allow task execution.
	mf.persistence.EXPECT().WorkerSeen(echoCtx.Request().Context(), &worker)

	err := mf.flamenco.ScheduleTask(echoCtx)
	assert.NoError(t, err)

	expectBody := api.WorkerStateChange{StatusRequested: api.WorkerStatusAsleep}
	assertResponseJSON(t, echoCtx, http.StatusLocked, expectBody)
}

func TestWorkerSignOn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusOffline
	prevStatus := worker.Status

	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:             worker.UUID,
		Nickname:       "Lazy Boi",
		PreviousStatus: &prevStatus,
		Status:         api.WorkerStatusStarting,
		Updated:        worker.UpdatedAt,
		Version:        "3.0-testing",
	})

	mf.persistence.EXPECT().SaveWorker(gomock.Any(), &worker).Return(nil)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

	echo := mf.prepareMockedJSONRequest(api.WorkerSignOn{
		Nickname:           "Lazy Boi",
		SoftwareVersion:    "3.0-testing",
		SupportedTaskTypes: []string{"testing", "sleeping", "snoozing"},
	})
	requestWorkerStore(echo, &worker)
	err := mf.flamenco.SignOn(echo)
	assert.NoError(t, err)

	assertResponseJSON(t, echo, http.StatusOK, api.WorkerStateChange{
		StatusRequested: api.WorkerStatusAwake,
	})
}

func TestWorkerSignoffTaskRequeue(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	// Signing off should be handled completely, even when the HTTP connection
	// breaks. This means using a different context than the one passed by Echo.
	echo := mf.prepareMockedRequest(nil)
	requestWorkerStore(echo, &worker)
	expectCtx := gomock.Not(gomock.Eq(echo.Request().Context()))

	// Expect worker's tasks to be re-queued.
	mf.stateMachine.EXPECT().RequeueTasksOfWorker(expectCtx, &worker, "worker signed off").Return(nil)
	mf.persistence.EXPECT().WorkerSeen(expectCtx, &worker)

	// Expect worker to be saved as 'offline'.
	mf.persistence.EXPECT().
		SaveWorkerStatus(expectCtx, &worker).
		Do(func(ctx context.Context, w *persistence.Worker) error {
			assert.Equal(t, api.WorkerStatusOffline, w.Status)
			return nil
		})

	prevStatus := api.WorkerStatusAwake
	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:             worker.UUID,
		Nickname:       worker.Name,
		PreviousStatus: &prevStatus,
		Status:         api.WorkerStatusOffline,
		Updated:        worker.UpdatedAt,
		Version:        worker.Software,
	})

	err := mf.flamenco.SignOff(echo)
	assert.NoError(t, err)

	resp := getRecordedResponse(echo)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestWorkerSignoffStatusChangeRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusAwake
	worker.StatusChangeRequest(api.WorkerStatusOffline, true)

	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:             worker.UUID,
		Nickname:       worker.Name,
		PreviousStatus: ptr(api.WorkerStatusAwake),
		Status:         api.WorkerStatusOffline,
		Updated:        worker.UpdatedAt,
		Version:        worker.Software,
	})

	// Expect the Worker to be saved with the status change removed.
	savedWorker := worker
	savedWorker.Status = api.WorkerStatusOffline
	savedWorker.StatusChangeClear()
	mf.persistence.EXPECT().SaveWorkerStatus(gomock.Any(), &savedWorker).Return(nil)

	mf.stateMachine.EXPECT().RequeueTasksOfWorker(gomock.Any(), &worker, "worker signed off").Return(nil)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

	// Perform the request
	echo := mf.prepareMockedRequest(nil)
	requestWorkerStore(echo, &worker)
	err := mf.flamenco.SignOff(echo)
	assert.NoError(t, err)
	assertResponseEmpty(t, echo)
}

func TestWorkerStateChanged(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusStarting
	prevStatus := worker.Status

	// Expect a broadcast of the change
	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:             worker.UUID,
		Nickname:       worker.Name,
		PreviousStatus: &prevStatus,
		Status:         api.WorkerStatusAwake,
		Updated:        worker.UpdatedAt,
		Version:        worker.Software,
	})

	// Expect the Worker to be saved with the new status
	savedWorker := worker
	savedWorker.Status = api.WorkerStatusAwake
	mf.persistence.EXPECT().SaveWorkerStatus(gomock.Any(), &savedWorker).Return(nil)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

	// Perform the request
	echo := mf.prepareMockedJSONRequest(api.WorkerStateChanged{
		Status: api.WorkerStatusAwake,
	})
	requestWorkerStore(echo, &worker)
	err := mf.flamenco.WorkerStateChanged(echo)
	assert.NoError(t, err)
	assertResponseEmpty(t, echo)
}

func TestWorkerStateChangedAfterChangeRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusOffline
	worker.StatusChangeRequest(api.WorkerStatusAsleep, false)

	{
		// Expect a broadcast of the change, even though it's not the state that was requested.
		// This is to allow some flexibility, for example when a worker has to go
		// asleep but would do so via `offline → starting → asleep`.
		mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
			Id:             worker.UUID,
			Nickname:       worker.Name,
			PreviousStatus: ptr(api.WorkerStatusOffline),
			Status:         api.WorkerStatusStarting,
			Updated:        worker.UpdatedAt,
			Version:        worker.Software,
			StatusChange: &api.WorkerStatusChangeRequest{
				Status: api.WorkerStatusAsleep,
				IsLazy: false,
			},
		})

		// Expect the Worker to be saved with the new status, but with the status
		// request still in place as it hasn't been met yet.
		savedWorker := worker
		savedWorker.Status = api.WorkerStatusStarting
		mf.persistence.EXPECT().SaveWorkerStatus(gomock.Any(), &savedWorker).Return(nil)
		mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

		// Perform the request
		echo := mf.prepareMockedJSONRequest(api.WorkerStateChanged{
			Status: api.WorkerStatusStarting,
		})
		requestWorkerStore(echo, &worker)
		err := mf.flamenco.WorkerStateChanged(echo)
		assert.NoError(t, err)
		assertResponseEmpty(t, echo)
	}

	// Do another status change, which does meet the requested state.
	{
		// Expect a broadcast.
		mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
			Id:             worker.UUID,
			Nickname:       worker.Name,
			PreviousStatus: ptr(api.WorkerStatusStarting),
			Status:         api.WorkerStatusAsleep,
			Updated:        worker.UpdatedAt,
			Version:        worker.Software,
		})

		// Expect the Worker to be saved with the new status and the status request
		// erased.
		savedWorker := worker
		savedWorker.Status = api.WorkerStatusAsleep
		savedWorker.StatusChangeClear()
		mf.persistence.EXPECT().SaveWorkerStatus(gomock.Any(), &savedWorker).Return(nil)
		mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

		// Perform the request
		echo := mf.prepareMockedJSONRequest(api.WorkerStateChanged{
			Status: api.WorkerStatusAsleep,
		})
		requestWorkerStore(echo, &worker)
		err := mf.flamenco.WorkerStateChanged(echo)
		assert.NoError(t, err)
		assertResponseEmpty(t, echo)
	}
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
		TaskStatus: ptr(api.TaskStatusCompleted),
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
	mf.stateMachine.EXPECT().TaskStatusChange(gomock.Any(), gomock.AssignableToTypeOf(&persistence.Task{}), api.TaskStatusCompleted).
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

	// Expect the log to be written and broadcast over SocketIO.
	mf.logStorage.EXPECT().Write(gomock.Any(), jobID, taskID, "line1\nline2\n")

	// Expect a 'touch' of the task.
	var touchedTask persistence.Task
	mf.persistence.EXPECT().TaskTouchedByWorker(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task *persistence.Task) error {
			touchedTask = *task
			return nil
		})
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

	// Do the call.
	echoCtx := mf.prepareMockedJSONRequest(taskUpdate)
	requestWorkerStore(echoCtx, &worker)
	err := mf.flamenco.TaskUpdate(echoCtx, taskID)

	// Check the saved task.
	assert.NoError(t, err)
	assert.Equal(t, mockTask.UUID, statusChangedtask.UUID)
	assert.Equal(t, mockTask.UUID, actUpdatedTask.UUID)
	assert.Equal(t, mockTask.UUID, touchedTask.UUID)
	assert.Equal(t, "testing", statusChangedtask.Activity)
	assert.Equal(t, "testing", actUpdatedTask.Activity)
}

func TestTaskUpdateFailed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	// Construct the JSON request object.
	taskUpdate := api.TaskUpdateJSONRequestBody{
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

	conf := config.Conf{
		Base: config.Base{
			TaskFailAfterSoftFailCount: 3,
		},
	}
	mf.config.EXPECT().Get().Return(&conf).AnyTimes()

	// Expect the task to be fetched for each sub-test:
	mf.persistence.EXPECT().FetchTask(gomock.Any(), taskID).Return(&mockTask, nil).Times(2)

	// Expect a 'touch' of the task for each sub-test:
	mf.persistence.EXPECT().TaskTouchedByWorker(gomock.Any(), &mockTask).Times(2)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker).Times(2)

	{
		// Expect the Worker to be added to the list of workers.
		// This returns 1, which is less than the failure threshold -> soft failure expected.
		mf.persistence.EXPECT().AddWorkerToTaskFailedList(gomock.Any(), &mockTask, &worker).Return(1, nil)

		// Expect soft failure.
		mf.stateMachine.EXPECT().TaskStatusChange(gomock.Any(), &mockTask, api.TaskStatusSoftFailed)
		mf.logStorage.EXPECT().WriteTimestamped(gomock.Any(), jobID, taskID,
			"Task failed by 1 worker, Manager will mark it as soft failure. 2 more failures will cause hard failure.")

		// Do the call.
		echoCtx := mf.prepareMockedJSONRequest(taskUpdate)
		requestWorkerStore(echoCtx, &worker)
		err := mf.flamenco.TaskUpdate(echoCtx, taskID)
		assert.NoError(t, err)
		assertResponseEmpty(t, echoCtx)
	}

	{
		// Test with more (mocked) failures in the past, pushing the task over the threshold.
		mf.persistence.EXPECT().AddWorkerToTaskFailedList(gomock.Any(), &mockTask, &worker).
			Return(conf.TaskFailAfterSoftFailCount, nil)
		mf.stateMachine.EXPECT().TaskStatusChange(gomock.Any(), &mockTask, api.TaskStatusFailed)
		mf.logStorage.EXPECT().WriteTimestamped(gomock.Any(), jobID, taskID,
			"Task failed by 3 workers, Manager will mark it as hard failure")

		// Do the call.
		echoCtx := mf.prepareMockedJSONRequest(taskUpdate)
		requestWorkerStore(echoCtx, &worker)
		err := mf.flamenco.TaskUpdate(echoCtx, taskID)
		assert.NoError(t, err)
		assertResponseEmpty(t, echoCtx)
	}
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

	// Expect the worker to be marked as 'seen' regardless of whether it may run
	// its current task or not, so equal to the number of calls to
	// `MayWorkerRun()` below.
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker).Times(4)

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
		// Expect a 'touch' of the task.
		mf.persistence.EXPECT().TaskTouchedByWorker(gomock.Any(), &task).Return(nil)

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
		worker.StatusChangeRequest(api.WorkerStatusAsleep, false)
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
