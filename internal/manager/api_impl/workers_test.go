package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/internal/manager/last_rendered"
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
		Commands: []persistence.Command{
			{Name: "test", Parameters: map[string]interface{}{
				"param": "prefix-{variable}-suffix",
			}},
		},
	}

	ctx := echo.Request().Context()
	bgCtx := gomock.Not(ctx)
	mf.persistence.EXPECT().ScheduleTask(ctx, &worker).Return(&task, nil)
	mf.persistence.EXPECT().TaskTouchedByWorker(bgCtx, &task)
	mf.persistence.EXPECT().WorkerSeen(bgCtx, &worker)
	mf.expectExpandVariables(t,
		config.VariableAudienceWorkers,
		config.VariablePlatform(worker.Platform),
		map[string]string{"variable": "value"},
	)

	mf.logStorage.EXPECT().WriteTimestamped(bgCtx, job.UUID, task.UUID,
		"Task assigned to worker дрон (e7632d62-c3b8-4af0-9e78-01752928952c)")

	mf.stateMachine.EXPECT().TaskStatusChange(bgCtx, &task, api.TaskStatusActive)
	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(gomock.Any())

	err := mf.flamenco.ScheduleTask(echo)
	assert.NoError(t, err)

	// Check the response
	assignedTask := api.AssignedTask{
		Uuid: task.UUID,
		Job:  job.UUID,
		Commands: []api.Command{
			{Name: "test", Parameters: map[string]interface{}{
				"param": "prefix-value-suffix",
			}},
		},
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
	bgCtx := gomock.Not(ctx)
	mf.persistence.EXPECT().ScheduleTask(ctx, &worker).Return(nil, nil)

	// This call should still trigger a "worker seen" call, as the worker is
	// actively asking for tasks.
	mf.persistence.EXPECT().WorkerSeen(bgCtx, &worker)

	err := mf.flamenco.ScheduleTask(echo)
	assert.NoError(t, err)
	assertResponseNoContent(t, echo)
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
	bgCtx := gomock.Not(echoCtx.Request().Context())
	mf.persistence.EXPECT().WorkerSeen(bgCtx, &worker)

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
	bgCtx := gomock.Not(echoCtx.Request().Context())
	mf.persistence.EXPECT().WorkerSeen(bgCtx, &worker)

	err := mf.flamenco.ScheduleTask(echoCtx)
	assert.NoError(t, err)

	expectBody := api.WorkerStateChange{StatusRequested: api.WorkerStatusAsleep}
	assertResponseJSON(t, echoCtx, http.StatusLocked, expectBody)
}

func TestTaskScheduleOtherStatusRequestedAndBadState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	// Even when the worker is in a state that doesn't allow execution, if there
	// is a status change requested, this should be communicated to the worker.
	worker.Status = api.WorkerStatusError
	worker.StatusChangeRequest(api.WorkerStatusAwake, false)

	echoCtx := mf.prepareMockedRequest(nil)
	requestWorkerStore(echoCtx, &worker)

	// The worker should be marked as 'seen', even when it's in a state that
	// doesn't allow task execution.
	bgCtx := gomock.Not(echoCtx.Request().Context())
	mf.persistence.EXPECT().WorkerSeen(bgCtx, &worker)

	err := mf.flamenco.ScheduleTask(echoCtx)
	assert.NoError(t, err)

	expectBody := api.WorkerStateChange{StatusRequested: api.WorkerStatusAwake}
	assertResponseJSON(t, echoCtx, http.StatusLocked, expectBody)
}

func TestWorkerSignOn(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusOffline
	prevStatus := worker.Status

	mf.sleepScheduler.EXPECT().WorkerStatus(gomock.Any(), worker.UUID).
		Return(api.WorkerStatusAsleep, nil)

	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:             worker.UUID,
		Name:           "Lazy Boi",
		PreviousStatus: &prevStatus,
		Status:         api.WorkerStatusStarting,
		Updated:        worker.UpdatedAt,
		Version:        "3.0-testing",
	})

	mf.persistence.EXPECT().SaveWorker(gomock.Any(), &worker).Return(nil)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

	echo := mf.prepareMockedJSONRequest(api.WorkerSignOn{
		Name:               "Lazy Boi",
		SoftwareVersion:    "3.0-testing",
		SupportedTaskTypes: []string{"testing", "sleeping", "snoozing"},
	})
	requestWorkerStore(echo, &worker)
	err := mf.flamenco.SignOn(echo)
	assert.NoError(t, err)

	assertResponseJSON(t, echo, http.StatusOK, api.WorkerStateChange{
		StatusRequested: api.WorkerStatusAsleep,
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
	mf.stateMachine.EXPECT().RequeueActiveTasksOfWorker(expectCtx, &worker, "worker signed off").Return(nil)
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
		Name:           worker.Name,
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
		Name:           worker.Name,
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

	mf.stateMachine.EXPECT().RequeueActiveTasksOfWorker(gomock.Any(), &worker, "worker signed off").Return(nil)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

	// Perform the request
	echo := mf.prepareMockedRequest(nil)
	requestWorkerStore(echo, &worker)
	err := mf.flamenco.SignOff(echo)
	assert.NoError(t, err)
	assertResponseNoContent(t, echo)
}

func TestWorkerState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker).Times(2)

	// No state change requested.
	{
		echo := mf.prepareMockedRequest(nil)
		requestWorkerStore(echo, &worker)
		err := mf.flamenco.WorkerState(echo)
		if assert.NoError(t, err) {
			assertResponseNoContent(t, echo)
		}
	}

	// State change requested.
	{
		requestStatus := api.WorkerStatusAsleep
		worker.StatusChangeRequest(requestStatus, false)

		echo := mf.prepareMockedRequest(nil)
		requestWorkerStore(echo, &worker)

		err := mf.flamenco.WorkerState(echo)
		if assert.NoError(t, err) {
			assertResponseJSON(t, echo, http.StatusOK, api.WorkerStateChange{
				StatusRequested: requestStatus,
			})
		}
	}
}

func TestWorkerStateChanged(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()
	worker.Status = api.WorkerStatusAwake
	prevStatus := worker.Status

	// Expect a broadcast of the change
	mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:             worker.UUID,
		Name:           worker.Name,
		PreviousStatus: &prevStatus,
		Status:         api.WorkerStatusAsleep,
		Updated:        worker.UpdatedAt,
		Version:        worker.Software,
	})

	// Expect the Worker to be saved with the new status
	savedWorker := worker
	savedWorker.Status = api.WorkerStatusAsleep
	mf.persistence.EXPECT().SaveWorkerStatus(gomock.Any(), &savedWorker).Return(nil)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)
	mf.stateMachine.EXPECT().RequeueActiveTasksOfWorker(gomock.Any(), &worker,
		"worker дрон (e7632d62-c3b8-4af0-9e78-01752928952c) changed status to 'asleep'")

	// Perform the request
	echo := mf.prepareMockedJSONRequest(api.WorkerStateChanged{
		Status: api.WorkerStatusAsleep,
	})
	requestWorkerStore(echo, &worker)
	err := mf.flamenco.WorkerStateChanged(echo)
	assert.NoError(t, err)
	assertResponseNoContent(t, echo)
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
			Name:           worker.Name,
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
		assertResponseNoContent(t, echo)
	}

	// Do another status change, which does meet the requested state.
	{
		// Expect a broadcast.
		mf.broadcaster.EXPECT().BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
			Id:             worker.UUID,
			Name:           worker.Name,
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
		assertResponseNoContent(t, echo)
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

func TestTaskOutputProduced(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	prepareRequest := func(body io.Reader) echo.Context {
		echo := mf.prepareMockedRequest(body)
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

	// Mock body to use in the request.
	bodyBytes := []byte("JPEG file contents")

	mf.lastRender.EXPECT().ThumbSpecs().Return([]last_rendered.Thumbspec{
		{Filename: "big'un.jpg", MaxWidth: 1920, MaxHeight: 1080},
		{Filename: "tiny.jpg", MaxWidth: 320, MaxHeight: 240},
	}).AnyTimes()
	mf.lastRender.EXPECT().PathForJob(job.UUID).Return("/path/to/job").AnyTimes()
	mf.localStorage.EXPECT().RelPath("/path/to/job").Return("relative/path/to/job", nil).AnyTimes()

	// Test: unhappy, missing Content-Length header.
	{
		mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

		echo := prepareRequest(nil)
		err := mf.flamenco.TaskOutputProduced(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseAPIError(t, echo, http.StatusLengthRequired, "Content-Length header required")
	}

	// Test: unhappy, too large Content-Length header.
	{
		mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)

		bodyBytes := bytes.Repeat([]byte("x"), int(last_rendered.MaxImageSizeBytes+1))
		if int64(len(bodyBytes)) != last_rendered.MaxImageSizeBytes+1 {
			panic("cannot generate enough bytes")
		}

		echo := prepareRequest(bytes.NewReader(bodyBytes))
		err := mf.flamenco.TaskOutputProduced(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseAPIError(t, echo, http.StatusRequestEntityTooLarge,
			"image too large; should be max %v bytes", last_rendered.MaxImageSizeBytes)
	}

	// Test: unhappy, wrong mime type
	{
		mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)
		mf.persistence.EXPECT().FetchTask(gomock.Any(), task.UUID).Return(&task, nil)

		echo := prepareRequest(bytes.NewReader(bodyBytes))
		echo.Request().Header.Set("Content-Type", "image/openexr")
		mf.lastRender.EXPECT().QueueImage(gomock.Any()).Return(last_rendered.ErrMimeTypeUnsupported)

		err := mf.flamenco.TaskOutputProduced(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseAPIError(t, echo, http.StatusUnsupportedMediaType, `unsupported mime type "image/openexr"`)
	}

	// Test: unhappy, queue full
	{
		mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)
		mf.persistence.EXPECT().FetchTask(gomock.Any(), task.UUID).Return(&task, nil)

		echo := prepareRequest(bytes.NewReader(bodyBytes))
		mf.lastRender.EXPECT().QueueImage(gomock.Any()).Return(last_rendered.ErrQueueFull)

		err := mf.flamenco.TaskOutputProduced(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseAPIError(t, echo, http.StatusTooManyRequests, "image processing queue is full")
	}

	// Test: happy
	{
		mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker)
		mf.persistence.EXPECT().FetchTask(gomock.Any(), task.UUID).Return(&task, nil)
		// Don't expect persistence.SetLastRendered(...) quite yet. That should be
		// called after the image processing is done.

		echo := prepareRequest(bytes.NewReader(bodyBytes))
		echo.Request().Header.Set("Content-Type", "image/jpeg")
		expectPayload := last_rendered.Payload{
			JobUUID:    job.UUID,
			WorkerUUID: worker.UUID,
			MimeType:   "image/jpeg",
			Image:      bodyBytes,
		}
		var actualPayload *last_rendered.Payload
		mf.lastRender.EXPECT().QueueImage(gomock.Any()).DoAndReturn(func(payload last_rendered.Payload) error {
			actualPayload = &payload
			return nil
		})

		err := mf.flamenco.TaskOutputProduced(echo, task.UUID)
		assert.NoError(t, err)
		assertResponseNoBody(t, echo, http.StatusAccepted)

		if assert.NotNil(t, actualPayload) {
			ctx := context.Background()
			mf.persistence.EXPECT().SetLastRendered(ctx, &job)

			expectBroadcast := api.SocketIOLastRenderedUpdate{
				JobId: job.UUID,
				Thumbnail: api.JobLastRenderedImageInfo{
					Base:     "/job-files/relative/path/to/job",
					Suffixes: []string{"big'un.jpg", "tiny.jpg"},
				},
			}
			mf.broadcaster.EXPECT().BroadcastLastRenderedImage(expectBroadcast)

			// Calling the callback function is normally done by the last-rendered image processor.
			actualPayload.Callback(ctx)

			// Compare the parameter to `QueueImage()` in a way that ignores the callback function.
			actualPayload.Callback = nil
			assert.Equal(t, expectPayload, *actualPayload)
		}
	}

}
