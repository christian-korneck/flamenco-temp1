package task_state_machine

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/task_state_machine/mocks"
	"git.blender.org/flamenco/pkg/api"
)

type StateMachineMocks struct {
	persist     *mocks.MockPersistenceService
	broadcaster *mocks.MockChangeBroadcaster
}

// In the comments below, "T" indicates the performed task status change, and
// "J" the expected resulting job status change.

func TestTaskStatusChangeQueuedToActive(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: queued > active  --> J: queued > active
	task := taskWithStatus(api.JobStatusQueued, api.TaskStatusQueued)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusActive)
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusActive)
	mocks.expectBroadcastJobChange(task.Job, api.JobStatusQueued, api.JobStatusActive)
	mocks.expectBroadcastTaskChange(task, api.TaskStatusQueued, api.TaskStatusActive)

	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusActive))
}

func TestTaskStatusChangeSaveTaskAfterJobChangeFailure(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// A task status change should be saved, even when triggering the job change errors somehow.
	task := taskWithStatus(api.JobStatusQueued, api.TaskStatusQueued)

	jobSaveErr := errors.New("hypothetical job save error")
	mocks.persist.EXPECT().
		SaveJobStatus(gomock.Any(), task.Job).
		Return(jobSaveErr)

	// Expect a call to save the task in the persistence layer, regardless of the above error.
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusActive)
	mocks.expectBroadcastTaskChange(task, api.TaskStatusQueued, api.TaskStatusActive)

	returnedErr := sm.TaskStatusChange(ctx, task, api.TaskStatusActive)
	assert.ErrorIs(t, returnedErr, jobSaveErr, "the returned error should wrap the persistence layer error")
}

func TestTaskStatusChangeActiveToCompleted(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// Job has three tasks.
	task := taskWithStatus(api.JobStatusActive, api.TaskStatusActive)
	task2 := taskOfSameJob(task, api.TaskStatusActive)
	task3 := taskOfSameJob(task, api.TaskStatusActive)

	// First task completing: T: active > completed --> J: active > active
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusCompleted)
	mocks.expectBroadcastTaskChange(task, api.TaskStatusActive, api.TaskStatusCompleted)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusCompleted).Return(1, 3, nil) // 1 of 3 complete.
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusCompleted))

	// Second task hickup: T: active > soft-failed --> J: active > active
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusSoftFailed)
	mocks.expectBroadcastTaskChange(task2, api.TaskStatusActive, api.TaskStatusSoftFailed)
	assert.NoError(t, sm.TaskStatusChange(ctx, task2, api.TaskStatusSoftFailed))

	// Second task completing: T: soft-failed > completed --> J: active > active
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusCompleted)
	mocks.expectBroadcastTaskChange(task2, api.TaskStatusSoftFailed, api.TaskStatusCompleted)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusCompleted).Return(2, 3, nil) // 2 of 3 complete.
	assert.NoError(t, sm.TaskStatusChange(ctx, task2, api.TaskStatusCompleted))

	// Third task completing: T: active > completed --> J: active > completed
	mocks.expectSaveTaskWithStatus(t, task3, api.TaskStatusCompleted)
	mocks.expectBroadcastTaskChange(task3, api.TaskStatusActive, api.TaskStatusCompleted)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusCompleted).Return(3, 3, nil) // 3 of 3 complete.
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusCompleted)
	mocks.expectBroadcastJobChange(task.Job, api.JobStatusActive, api.JobStatusCompleted)

	assert.NoError(t, sm.TaskStatusChange(ctx, task3, api.TaskStatusCompleted))
}

func TestTaskStatusChangeQueuedToFailed(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: queued > failed (1% task failure) --> J: queued > active
	task := taskWithStatus(api.JobStatusQueued, api.TaskStatusQueued)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusFailed)
	mocks.expectBroadcastTaskChange(task, api.TaskStatusQueued, api.TaskStatusFailed)
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusActive)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusFailed).Return(1, 100, nil) // 1 out of 100 failed.
	mocks.expectBroadcastJobChange(task.Job, api.JobStatusQueued, api.JobStatusActive)

	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusFailed))
}

func TestTaskStatusChangeActiveToFailedFailJob(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: active > failed (10% task1 failure) --> J: active > failed + cancellation of any runnable tasks.
	task1 := taskWithStatus(api.JobStatusActive, api.TaskStatusActive)
	mocks.expectSaveTaskWithStatus(t, task1, api.TaskStatusFailed)
	// The change to the failed task should be broadcast.
	mocks.expectBroadcastTaskChange(task1, api.TaskStatusActive, api.TaskStatusFailed)
	mocks.expectSaveJobWithStatus(t, task1.Job, api.JobStatusFailed)
	// The resulting cancellation of the other tasks should be communicated as mass-task-update in the job update broadcast.
	mocks.expectBroadcastJobChangeWithTaskRefresh(task1.Job, api.JobStatusActive, api.JobStatusFailed)

	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task1.Job, api.TaskStatusFailed).Return(10, 100, nil) // 10 out of 100 failed.

	// Expect failure of the job to trigger cancellation of remaining tasks.
	taskStatusesToCancel := []api.TaskStatus{
		api.TaskStatusActive,
		api.TaskStatusQueued,
		api.TaskStatusSoftFailed,
	}

	mocks.persist.EXPECT().UpdateJobsTaskStatusesConditional(ctx, task1.Job, taskStatusesToCancel, api.TaskStatusCanceled,
		"Manager cancelled this task because the job got status \"failed\".",
	)

	assert.NoError(t, sm.TaskStatusChange(ctx, task1, api.TaskStatusFailed))
}

func TestTaskStatusChangeRequeueOnCompletedJob(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: completed > queued --> J: completed > requeueing > queued
	task := taskWithStatus(api.JobStatusCompleted, api.TaskStatusCompleted)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusQueued)
	mocks.expectBroadcastTaskChange(task, api.TaskStatusCompleted, api.TaskStatusQueued)
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusRequeueing)
	mocks.expectBroadcastJobChangeWithTaskRefresh(task.Job, api.JobStatusCompleted, api.JobStatusRequeueing)
	mocks.expectBroadcastJobChangeWithTaskRefresh(task.Job, api.JobStatusRequeueing, api.JobStatusQueued)

	// Expect queueing of the job to trigger queueing of all its tasks, if those tasks were all completed before.
	// 2 out of 3 completed, because one was just queued.
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusCompleted).Return(2, 3, nil)
	mocks.persist.EXPECT().UpdateJobsTaskStatuses(ctx, task.Job, api.TaskStatusQueued,
		"Queued because job transitioned status from \"completed\" to \"requeueing\"",
	)
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusQueued)

	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusQueued))
}

func TestTaskStatusChangeCancelSingleTask(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task := taskWithStatus(api.JobStatusCancelRequested, api.TaskStatusActive)
	task2 := taskOfSameJob(task, api.TaskStatusQueued)
	job := task.Job

	// T1: active > cancelled --> J: cancel-requested > cancel-requested
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusCanceled)
	mocks.expectBroadcastTaskChange(task, api.TaskStatusActive, api.TaskStatusCanceled)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, job,
		api.TaskStatusActive, api.TaskStatusQueued, api.TaskStatusSoftFailed).
		Return(1, 2, nil)
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusCanceled))

	// T2: queued > cancelled --> J: cancel-requested > canceled
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusCanceled)
	mocks.expectBroadcastTaskChange(task2, api.TaskStatusQueued, api.TaskStatusCanceled)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, job,
		api.TaskStatusActive, api.TaskStatusQueued, api.TaskStatusSoftFailed).
		Return(0, 2, nil)
	mocks.expectSaveJobWithStatus(t, job, api.JobStatusCanceled)
	mocks.expectBroadcastJobChange(task.Job, api.JobStatusCancelRequested, api.JobStatusCanceled)

	assert.NoError(t, sm.TaskStatusChange(ctx, task2, api.TaskStatusCanceled))
}

func TestTaskStatusChangeCancelSingleTaskWithOtherFailed(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task1 := taskWithStatus(api.JobStatusCancelRequested, api.TaskStatusActive)
	task2 := taskOfSameJob(task1, api.TaskStatusFailed)
	taskOfSameJob(task2, api.TaskStatusPaused)
	job := task1.Job

	// T1: active > cancelled --> J: cancel-requested > canceled because T2 already failed and cannot run anyway.
	mocks.expectSaveTaskWithStatus(t, task1, api.TaskStatusCanceled)
	mocks.expectBroadcastTaskChange(task1, api.TaskStatusActive, api.TaskStatusCanceled)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, job,
		api.TaskStatusActive, api.TaskStatusQueued, api.TaskStatusSoftFailed).
		Return(0, 3, nil)
	mocks.expectSaveJobWithStatus(t, job, api.JobStatusCanceled)
	mocks.expectBroadcastJobChange(task1.Job, api.JobStatusCancelRequested, api.JobStatusCanceled)

	// The paused task just stays paused, so don't expectBroadcastTaskChange(task3).

	assert.NoError(t, sm.TaskStatusChange(ctx, task1, api.TaskStatusCanceled))
}

func TestTaskStatusChangeUnknownStatus(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: queued > borked --> saved to DB but otherwise ignored w.r.t. job status changes.
	task := taskWithStatus(api.JobStatusQueued, api.TaskStatusQueued)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatus("borked"))
	mocks.expectBroadcastTaskChange(task, api.TaskStatusQueued, api.TaskStatus("borked"))

	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatus("borked")))
}

func TestJobRequeueWithSomeCompletedTasks(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task1 := taskWithStatus(api.JobStatusActive, api.TaskStatusCompleted)
	// These are not necessary to create for this test, but just imagine these tasks are there too.
	// This is mimicked by returning (1, 3, nil) when counting the tasks (1 of 3 completed).
	// task2 := taskOfSameJob(task1, api.TaskStatusFailed)
	// task3 := taskOfSameJob(task2, api.TaskStatusSoftFailed)
	job := task1.Job

	mocks.expectSaveJobWithStatus(t, job, api.JobStatusRequeueing)

	// Expect queueing of the job to trigger queueing of all its not-yet-completed tasks.
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, job, api.TaskStatusCompleted).Return(1, 3, nil)
	mocks.persist.EXPECT().UpdateJobsTaskStatusesConditional(ctx, job,
		[]api.TaskStatus{
			api.TaskStatusCanceled,
			api.TaskStatusFailed,
			api.TaskStatusPaused,
			api.TaskStatusSoftFailed,
		},
		api.TaskStatusQueued,
		"Queued because job transitioned status from \"active\" to \"requeueing\"",
	)

	mocks.expectSaveJobWithStatus(t, job, api.JobStatusQueued)

	mocks.expectBroadcastJobChangeWithTaskRefresh(job, api.JobStatusActive, api.JobStatusRequeueing)
	mocks.expectBroadcastJobChangeWithTaskRefresh(job, api.JobStatusRequeueing, api.JobStatusQueued)

	assert.NoError(t, sm.JobStatusChange(ctx, job, api.JobStatusRequeueing, "someone wrote a unittest"))
}

func TestJobRequeueWithAllCompletedTasks(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task1 := taskWithStatus(api.JobStatusCompleted, api.TaskStatusCompleted)
	// These are not necessary to create for this test, but just imagine these tasks are there too.
	// This is mimicked by returning (3, 3, nil) when counting the tasks (3 of 3 completed).
	// task2 := taskOfSameJob(task1, api.TaskStatusCompleted)
	// task3 := taskOfSameJob(task2, api.TaskStatusCompleted)
	job := task1.Job

	call1 := mocks.expectSaveJobWithStatus(t, job, api.JobStatusRequeueing)

	// Expect queueing of the job to trigger queueing of all its not-yet-completed tasks.
	updateCall := mocks.persist.EXPECT().
		UpdateJobsTaskStatuses(ctx, job, api.TaskStatusQueued,
			"Queued because job transitioned status from \"completed\" to \"requeueing\"").
		After(call1)

	saveJobCall := mocks.expectSaveJobWithStatus(t, job, api.JobStatusQueued).After(updateCall)

	mocks.persist.EXPECT().
		CountTasksOfJobInStatus(ctx, job, api.TaskStatusCompleted).
		Return(0, 3, nil). // By now all tasks are queued.
		After(saveJobCall)

	mocks.expectBroadcastJobChangeWithTaskRefresh(job, api.JobStatusCompleted, api.JobStatusRequeueing)
	mocks.expectBroadcastJobChangeWithTaskRefresh(job, api.JobStatusRequeueing, api.JobStatusQueued)

	assert.NoError(t, sm.JobStatusChange(ctx, job, api.JobStatusRequeueing, "someone wrote a unit test"))
}

func TestJobCancelWithSomeCompletedTasks(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task1 := taskWithStatus(api.JobStatusActive, api.TaskStatusCompleted)
	// task2 := taskOfSameJob(task1, api.TaskStatusFailed)
	// task3 := taskOfSameJob(task2, api.TaskStatusSoftFailed)
	job := task1.Job

	mocks.expectSaveJobWithStatus(t, job, api.JobStatusCancelRequested)

	// Expect cancelling of the job to trigger cancelling of all its could-potentially-still-run tasks.
	mocks.persist.EXPECT().UpdateJobsTaskStatusesConditional(ctx, job,
		[]api.TaskStatus{
			api.TaskStatusActive,
			api.TaskStatusQueued,
			// TODO: add api.TaskStatusPaused as well, as those should get cancelled too,
			api.TaskStatusSoftFailed,
		},
		api.TaskStatusCanceled,
		"Manager cancelled this task because the job got status \"cancel-requested\".",
	)

	mocks.expectSaveJobWithStatus(t, job, api.JobStatusCanceled)

	mocks.expectBroadcastJobChangeWithTaskRefresh(job, api.JobStatusActive, api.JobStatusCancelRequested)
	mocks.expectBroadcastJobChange(job, api.JobStatusCancelRequested, api.JobStatusCanceled)

	assert.NoError(t, sm.JobStatusChange(ctx, job, api.JobStatusCancelRequested, "someone wrote a unittest"))
}

func TestCheckStuck(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task1 := taskWithStatus(api.JobStatusActive, api.TaskStatusCompleted)
	// task2 := taskOfSameJob(task1, api.TaskStatusFailed)
	// task3 := taskOfSameJob(task2, api.TaskStatusSoftFailed)
	job := task1.Job
	job.Status = api.JobStatusRequeueing

	mocks.persist.EXPECT().FetchJobsInStatus(ctx, api.JobStatusCancelRequested, api.JobStatusRequeueing).
		Return([]*persistence.Job{job}, nil)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, job, api.TaskStatusCompleted).Return(1, 3, nil)

	mocks.persist.EXPECT().UpdateJobsTaskStatusesConditional(ctx, job,
		[]api.TaskStatus{
			api.TaskStatusCanceled,
			api.TaskStatusFailed,
			api.TaskStatusPaused,
			api.TaskStatusSoftFailed,
		},
		api.TaskStatusQueued,
		fmt.Sprintf("Queued because job transitioned status from %q to %q", job.Status, job.Status),
	)

	// Expect Job -> Queued and non-completed tasks -> Queued.
	mocks.expectSaveJobWithStatus(t, job, api.JobStatusRequeueing) // should be called once for the current status
	mocks.expectSaveJobWithStatus(t, job, api.JobStatusQueued)     // and then with the new status

	mocks.expectBroadcastJobChangeWithTaskRefresh(job, api.JobStatusRequeueing, api.JobStatusRequeueing)
	mocks.expectBroadcastJobChangeWithTaskRefresh(job, api.JobStatusRequeueing, api.JobStatusQueued)

	sm.CheckStuck(ctx)
}

func mockedTaskStateMachine(mockCtrl *gomock.Controller) (*StateMachine, *StateMachineMocks) {
	mocks := StateMachineMocks{
		persist:     mocks.NewMockPersistenceService(mockCtrl),
		broadcaster: mocks.NewMockChangeBroadcaster(mockCtrl),
	}
	sm := NewStateMachine(mocks.persist, mocks.broadcaster)
	return sm, &mocks
}

func (m *StateMachineMocks) expectSaveTaskWithStatus(
	t *testing.T,
	task *persistence.Task,
	expectTaskStatus api.TaskStatus,
) *gomock.Call {
	return m.persist.EXPECT().
		SaveTask(gomock.Any(), task).
		DoAndReturn(func(ctx context.Context, savedTask *persistence.Task) error {
			assert.Equal(t, expectTaskStatus, savedTask.Status)
			return nil
		})
}

func (m *StateMachineMocks) expectSaveJobWithStatus(
	t *testing.T,
	job *persistence.Job,
	expectJobStatus api.JobStatus,
) *gomock.Call {
	return m.persist.EXPECT().
		SaveJobStatus(gomock.Any(), job).
		DoAndReturn(func(ctx context.Context, savedJob *persistence.Job) error {
			assert.Equal(t, expectJobStatus, savedJob.Status)
			return nil
		})
}

func (m *StateMachineMocks) expectBroadcastJobChange(
	job *persistence.Job,
	fromStatus, toStatus api.JobStatus,
) *gomock.Call {
	expectUpdate := api.SocketIOJobUpdate{
		Id:             job.UUID,
		Name:           &job.Name,
		PreviousStatus: &fromStatus,
		RefreshTasks:   false,
		Status:         toStatus,
		Updated:        job.UpdatedAt,
	}
	return m.broadcaster.EXPECT().BroadcastJobUpdate(expectUpdate)
}

func (m *StateMachineMocks) expectBroadcastJobChangeWithTaskRefresh(
	job *persistence.Job,
	fromStatus, toStatus api.JobStatus,
) *gomock.Call {
	expectUpdate := api.SocketIOJobUpdate{
		Id:             job.UUID,
		Name:           &job.Name,
		PreviousStatus: &fromStatus,
		RefreshTasks:   true,
		Status:         toStatus,
		Updated:        job.UpdatedAt,
	}
	return m.broadcaster.EXPECT().BroadcastJobUpdate(expectUpdate)
}

func (m *StateMachineMocks) expectBroadcastTaskChange(
	task *persistence.Task,
	fromStatus, toStatus api.TaskStatus,
) *gomock.Call {
	expectUpdate := api.SocketIOTaskUpdate{
		Id:             task.UUID,
		JobId:          task.Job.UUID,
		Name:           task.Name,
		Updated:        task.UpdatedAt,
		PreviousStatus: &fromStatus,
		Status:         toStatus,
	}
	return m.broadcaster.EXPECT().BroadcastTaskUpdate(expectUpdate)
}

/* taskWithStatus() creates a task of a certain status, with a job of a certain status. */
func taskWithStatus(jobStatus api.JobStatus, taskStatus api.TaskStatus) *persistence.Task {
	job := persistence.Job{
		Model: gorm.Model{ID: 47},
		UUID:  "test-job-f3f5-4cef-9cd7-e67eb28eaf3e",

		Status: jobStatus,
	}
	task := persistence.Task{
		Model: gorm.Model{ID: 327},
		UUID:  "testtask-0001-4e28-aeea-8cbaf2fc96a5",

		JobID: job.ID,
		Job:   &job,

		Status: taskStatus,
	}

	return &task
}

/* taskOfSameJob() creates a task of a certain status, on the same job as the given task. */
func taskOfSameJob(task *persistence.Task, taskStatus api.TaskStatus) *persistence.Task {
	newTaskID := task.ID + 1
	return &persistence.Task{
		Model:  gorm.Model{ID: newTaskID},
		UUID:   fmt.Sprintf("testtask-%04d-4e28-aeea-8cbaf2fc96a5", newTaskID),
		JobID:  task.JobID,
		Job:    task.Job,
		Status: taskStatus,
	}
}

func taskStateMachineTestFixtures(t *testing.T) (*gomock.Controller, context.Context, *StateMachine, *StateMachineMocks) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	sm, mocks := mockedTaskStateMachine(mockCtrl)
	return mockCtrl, ctx, sm, mocks
}
