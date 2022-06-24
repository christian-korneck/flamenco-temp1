package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

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
		Type:     "misc",
	}

	conf := config.Conf{
		Base: config.Base{
			TaskFailAfterSoftFailCount: 3,
			BlocklistThreshold:         65535, // This test doesn't cover blocklisting.
		},
	}
	mf.config.EXPECT().Get().Return(&conf).AnyTimes()

	const numSubTests = 2
	// Expect the task to be fetched for each sub-test:
	mf.persistence.EXPECT().FetchTask(gomock.Any(), taskID).Return(&mockTask, nil).Times(numSubTests)

	// Expect a 'touch' of the task for each sub-test:
	mf.persistence.EXPECT().TaskTouchedByWorker(gomock.Any(), &mockTask).Times(numSubTests)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker).Times(numSubTests)

	// Mimick that this is always first failure of this worker/job/tasktype combo:
	mf.persistence.EXPECT().CountTaskFailuresOfWorker(gomock.Any(), &mockJob, &worker, "misc").Return(0, nil).Times(numSubTests)

	{
		// Expect the Worker to be added to the list of failed workers.
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
		assertResponseNoContent(t, echoCtx)
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
		assertResponseNoContent(t, echoCtx)
	}
}

func TestBlockingAfterFailure(t *testing.T) {
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
		Type:     "misc",
	}

	conf := config.Conf{
		Base: config.Base{
			TaskFailAfterSoftFailCount: 3,
			BlocklistThreshold:         3,
		},
	}
	mf.config.EXPECT().Get().Return(&conf).AnyTimes()

	const numSubTests = 3
	// Expect the task to be fetched for each sub-test:
	mf.persistence.EXPECT().FetchTask(gomock.Any(), taskID).Return(&mockTask, nil).Times(numSubTests)

	// Expect a 'touch' of the task for each sub-test:
	mf.persistence.EXPECT().TaskTouchedByWorker(gomock.Any(), &mockTask).Times(numSubTests)
	mf.persistence.EXPECT().WorkerSeen(gomock.Any(), &worker).Times(numSubTests)

	// Mimick that this is the 3rd of this worker/job/tasktype combo, and thus should trigger a block.
	// Returns 2 because there have been 2 previous failures.
	mf.persistence.EXPECT().
		CountTaskFailuresOfWorker(gomock.Any(), &mockJob, &worker, "misc").
		Return(2, nil).
		Times(numSubTests)

	// Expect the worker to be blocked.
	mf.persistence.EXPECT().
		AddWorkerToJobBlocklist(gomock.Any(), &mockJob, &worker, "misc").
		Times(numSubTests)

	{
		// Mimick that there is another worker to work on this task, so the job should continue happily.
		mf.persistence.EXPECT().WorkersLeftToRun(gomock.Any(), &mockJob, "misc").
			Return(map[string]bool{"60453eec-5a26-43e9-9da2-d00506d492cc": true}, nil)
		mf.persistence.EXPECT().FetchTaskFailureList(gomock.Any(), &mockTask).
			Return([]*persistence.Worker{ /* It shouldn't matter whether the failing worker is here or not. */ }, nil)

		// Expect the Worker to be added to the list of failed workers for this task.
		// This returns 1, which is less than the failure threshold -> soft failure.
		mf.persistence.EXPECT().AddWorkerToTaskFailedList(gomock.Any(), &mockTask, &worker).Return(1, nil)

		// Expect soft failure of the task.
		mf.stateMachine.EXPECT().TaskStatusChange(gomock.Any(), &mockTask, api.TaskStatusSoftFailed)
		mf.logStorage.EXPECT().WriteTimestamped(gomock.Any(), jobID, taskID,
			"Task failed by 1 worker, Manager will mark it as soft failure. 2 more failures will cause hard failure.")

		// Because the job didn't fail in its entirety, the tasks previously failed
		// by the Worker should be requeued so they can be picked up by another.
		mf.stateMachine.EXPECT().RequeueFailedTasksOfWorkerOfJob(
			gomock.Any(), &worker, &mockJob,
			"worker дрон was blocked from tasks of type \"misc\"")

		// Do the call.
		echoCtx := mf.prepareMockedJSONRequest(taskUpdate)
		requestWorkerStore(echoCtx, &worker)
		err := mf.flamenco.TaskUpdate(echoCtx, taskID)
		assert.NoError(t, err)
		assertResponseNoContent(t, echoCtx)
	}

	{
		// Test without any workers left to run these tasks on this job due to blocklisting. This should fail the entire job.
		mf.persistence.EXPECT().WorkersLeftToRun(gomock.Any(), &mockJob, "misc").
			Return(map[string]bool{}, nil)
		mf.persistence.EXPECT().FetchTaskFailureList(gomock.Any(), &mockTask).
			Return([]*persistence.Worker{ /* It shouldn't matter whether the failing worker is here or not. */ }, nil)

		// Expect the Worker to be added to the list of failed workers for this task.
		// This returns 1, which is less than the failure threshold -> soft failure if it were only based on this metric.
		mf.persistence.EXPECT().AddWorkerToTaskFailedList(gomock.Any(), &mockTask, &worker).Return(1, nil)

		// Expect hard failure of the task, because there are no workers left to perfom it.
		mf.stateMachine.EXPECT().TaskStatusChange(gomock.Any(), &mockTask, api.TaskStatusFailed)
		mf.logStorage.EXPECT().WriteTimestamped(gomock.Any(), jobID, taskID,
			"Task failed by worker дрон (e7632d62-c3b8-4af0-9e78-01752928952c), Manager will fail the entire job "+
				"as there are no more workers left for tasks of type \"misc\".")

		// Expect failure of the job.
		mf.stateMachine.EXPECT().
			JobStatusChange(gomock.Any(), &mockJob, api.JobStatusFailed, "no more workers left to run tasks of type \"misc\"")

		// Because the job failed, there is no need to re-queue any tasks previously failed by this worker.

		// Do the call.
		echoCtx := mf.prepareMockedJSONRequest(taskUpdate)
		requestWorkerStore(echoCtx, &worker)
		err := mf.flamenco.TaskUpdate(echoCtx, taskID)
		assert.NoError(t, err)
		assertResponseNoContent(t, echoCtx)
	}

	{
		// Test that no worker has been blocklisted, but the one available one did fail this task.
		// This also makes the task impossible to run, and should just fail the entire job.
		theOtherFailingWorker := persistence.Worker{
			UUID: "ce312357-29cd-4389-81ab-4d43e30945f8",
		}
		mf.persistence.EXPECT().WorkersLeftToRun(gomock.Any(), &mockJob, "misc").
			Return(map[string]bool{theOtherFailingWorker.UUID: true}, nil)
		mf.persistence.EXPECT().FetchTaskFailureList(gomock.Any(), &mockTask).
			Return([]*persistence.Worker{&theOtherFailingWorker}, nil)

		// Expect the Worker to be added to the list of failed workers for this task.
		// This returns 1, which is less than the failure threshold -> soft failure if it were only based on this metric.
		mf.persistence.EXPECT().AddWorkerToTaskFailedList(gomock.Any(), &mockTask, &worker).Return(1, nil)

		// Expect hard failure of the task, because there are no workers left to perfom it.
		mf.stateMachine.EXPECT().TaskStatusChange(gomock.Any(), &mockTask, api.TaskStatusFailed)
		mf.logStorage.EXPECT().WriteTimestamped(gomock.Any(), jobID, taskID,
			"Task failed by worker дрон (e7632d62-c3b8-4af0-9e78-01752928952c), Manager will fail the entire job "+
				"as there are no more workers left for tasks of type \"misc\".")

		// Expect failure of the job.
		mf.stateMachine.EXPECT().
			JobStatusChange(gomock.Any(), &mockJob, api.JobStatusFailed, "no more workers left to run tasks of type \"misc\"")

		// Because the job failed, there is no need to re-queue any tasks previously failed by this worker.

		// Do the call.
		echoCtx := mf.prepareMockedJSONRequest(taskUpdate)
		requestWorkerStore(echoCtx, &worker)
		err := mf.flamenco.TaskUpdate(echoCtx, taskID)
		assert.NoError(t, err)
		assertResponseNoContent(t, echoCtx)
	}
}
