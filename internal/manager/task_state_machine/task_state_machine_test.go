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
	task2 := taskOfSameJob(task1, api.TaskStatusFailed)
	task3 := taskOfSameJob(task2, api.TaskStatusSoftFailed)
	remainingTasks := []*persistence.Task{task2, task3}

	mocks.expectSaveTaskWithStatus(t, task1, api.TaskStatusFailed)
	mocks.expectBroadcastTaskChange(task1, api.TaskStatusActive, api.TaskStatusFailed)
	mocks.expectSaveJobWithStatus(t, task1.Job, api.JobStatusFailed)
	mocks.expectBroadcastJobChange(task1.Job, api.JobStatusActive, api.JobStatusFailed)

	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task1.Job, api.TaskStatusFailed).Return(10, 100, nil) // 10 out of 100 failed.

	// Expect failure of the job to trigger cancellation of remaining tasks.
	mocks.persist.EXPECT().FetchTasksOfJobInStatus(ctx, task1.Job,
		api.TaskStatusActive,
		api.TaskStatusQueued,
		api.TaskStatusSoftFailed,
	).Return(remainingTasks, nil)
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusCanceled)
	mocks.expectSaveTaskWithStatus(t, task3, api.TaskStatusCanceled)

	mocks.expectBroadcastTaskChange(task2, api.TaskStatusFailed, api.TaskStatusCanceled)
	mocks.expectBroadcastTaskChange(task3, api.TaskStatusSoftFailed, api.TaskStatusCanceled)

	assert.NoError(t, sm.TaskStatusChange(ctx, task1, api.TaskStatusFailed))
}

func TestTaskStatusChangeRequeueOnCompletedJob(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: completed > queued --> J: completed > requeued > queued
	task1 := taskWithStatus(api.JobStatusCompleted, api.TaskStatusCompleted)
	task2 := taskOfSameJob(task1, api.TaskStatusCompleted)
	task3 := taskOfSameJob(task2, api.TaskStatusCompleted)
	allTaskIDs := []*persistence.Task{task1, task2, task3}

	mocks.expectSaveTaskWithStatus(t, task1, api.TaskStatusQueued)
	mocks.expectBroadcastTaskChange(task1, api.TaskStatusCompleted, api.TaskStatusQueued)
	mocks.expectSaveJobWithStatus(t, task1.Job, api.JobStatusRequeued)
	mocks.expectBroadcastJobChange(task1.Job, api.JobStatusCompleted, api.JobStatusRequeued)
	mocks.expectBroadcastJobChange(task1.Job, api.JobStatusRequeued, api.JobStatusQueued)

	// Expect queueing of the job to trigger queueing of all its tasks, if those tasks were all completed before.
	// 2 out of 3 completed, because one was just queued.
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task1.Job, api.TaskStatusCompleted).Return(2, 3, nil)
	fetchCall := mocks.persist.EXPECT().FetchTasksOfJob(ctx, task1.Job).Return(allTaskIDs, nil)
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusQueued).After(fetchCall)
	mocks.expectSaveTaskWithStatus(t, task3, api.TaskStatusQueued).After(fetchCall)

	mocks.expectBroadcastTaskChange(task2, api.TaskStatusCompleted, api.TaskStatusQueued)
	mocks.expectBroadcastTaskChange(task3, api.TaskStatusCompleted, api.TaskStatusQueued)

	mocks.expectSaveJobWithStatus(t, task1.Job, api.JobStatusQueued)

	assert.NoError(t, sm.TaskStatusChange(ctx, task1, api.TaskStatusQueued))
}

func TestTaskStatusChangeCancelSingleTask(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task := taskWithStatus(api.JobStatusCancelRequested, api.TaskStatusCancelRequested)
	task2 := taskOfSameJob(task, api.TaskStatusCancelRequested)
	job := task.Job

	// T1: cancel-requested > cancelled --> J: cancel-requested > cancel-requested
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusCanceled)
	mocks.expectBroadcastTaskChange(task, api.TaskStatusCancelRequested, api.TaskStatusCanceled)
	mocks.persist.EXPECT().JobHasTasksInStatus(ctx, job, api.TaskStatusCancelRequested).Return(true, nil)
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusCanceled))

	// T2: cancel-requested > cancelled --> J: cancel-requested > canceled
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusCanceled)
	mocks.expectBroadcastTaskChange(task2, api.TaskStatusCancelRequested, api.TaskStatusCanceled)
	mocks.persist.EXPECT().JobHasTasksInStatus(ctx, job, api.TaskStatusCancelRequested).Return(false, nil)
	mocks.expectSaveJobWithStatus(t, job, api.JobStatusCanceled)
	mocks.expectBroadcastJobChange(task.Job, api.JobStatusCancelRequested, api.JobStatusCanceled)

	assert.NoError(t, sm.TaskStatusChange(ctx, task2, api.TaskStatusCanceled))
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
	task2 := taskOfSameJob(task1, api.TaskStatusFailed)
	task3 := taskOfSameJob(task2, api.TaskStatusSoftFailed)
	notCompleteTasks := []*persistence.Task{task2, task3}
	job := task1.Job

	mocks.expectSaveJobWithStatus(t, job, api.JobStatusRequeued)

	// Expect queueing of the job to trigger queueing of all its not-yet-completed tasks.
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, job, api.TaskStatusCompleted).Return(1, 3, nil)
	mocks.persist.EXPECT().FetchTasksOfJobInStatus(ctx, job,
		api.TaskStatusCancelRequested,
		api.TaskStatusCanceled,
		api.TaskStatusFailed,
		api.TaskStatusPaused,
		api.TaskStatusSoftFailed,
	).Return(notCompleteTasks, nil)

	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusQueued)
	mocks.expectSaveTaskWithStatus(t, task3, api.TaskStatusQueued)
	mocks.expectSaveJobWithStatus(t, job, api.JobStatusQueued)

	mocks.expectBroadcastJobChange(job, api.JobStatusActive, api.JobStatusRequeued)
	mocks.expectBroadcastJobChange(job, api.JobStatusRequeued, api.JobStatusQueued)

	mocks.expectBroadcastTaskChange(task2, api.TaskStatusFailed, api.TaskStatusQueued)
	mocks.expectBroadcastTaskChange(task3, api.TaskStatusSoftFailed, api.TaskStatusQueued)

	assert.NoError(t, sm.JobStatusChange(ctx, job, api.JobStatusRequeued, "someone wrote a unittest"))
}

func TestJobRequeueWithAllCompletedTasks(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task1 := taskWithStatus(api.JobStatusCompleted, api.TaskStatusCompleted)
	task2 := taskOfSameJob(task1, api.TaskStatusCompleted)
	task3 := taskOfSameJob(task2, api.TaskStatusCompleted)
	allTasks := []*persistence.Task{task1, task2, task3}
	job := task1.Job

	call1 := mocks.expectSaveJobWithStatus(t, job, api.JobStatusRequeued)

	// Expect queueing of the job to trigger queueing of all its not-yet-completed tasks.
	fetchCall := mocks.persist.EXPECT().FetchTasksOfJob(ctx, job).
		Return(allTasks, nil).
		After(call1)

	mocks.expectSaveTaskWithStatus(t, task1, api.TaskStatusQueued).After(fetchCall)
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusQueued).After(fetchCall)
	mocks.expectSaveTaskWithStatus(t, task3, api.TaskStatusQueued).After(fetchCall)

	saveJobCall := mocks.expectSaveJobWithStatus(t, job, api.JobStatusQueued).After(fetchCall)

	mocks.persist.EXPECT().
		CountTasksOfJobInStatus(ctx, job, api.TaskStatusCompleted).
		Return(0, 3, nil). // By now all tasks are queued.
		After(saveJobCall)

	mocks.expectBroadcastJobChange(job, api.JobStatusCompleted, api.JobStatusRequeued)
	mocks.expectBroadcastJobChange(job, api.JobStatusRequeued, api.JobStatusQueued)

	mocks.expectBroadcastTaskChange(task1, api.TaskStatusCompleted, api.TaskStatusQueued)
	mocks.expectBroadcastTaskChange(task2, api.TaskStatusCompleted, api.TaskStatusQueued)
	mocks.expectBroadcastTaskChange(task3, api.TaskStatusCompleted, api.TaskStatusQueued)

	assert.NoError(t, sm.JobStatusChange(ctx, job, api.JobStatusRequeued, "someone wrote a unit test"))
}

func TestJobCancelWithSomeCompletedTasks(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task1 := taskWithStatus(api.JobStatusActive, api.TaskStatusCompleted)
	task2 := taskOfSameJob(task1, api.TaskStatusFailed)
	task3 := taskOfSameJob(task2, api.TaskStatusSoftFailed)
	job := task1.Job
	potentialRunTasks := []*persistence.Task{task2, task3}

	mocks.expectSaveJobWithStatus(t, job, api.JobStatusCancelRequested)

	// Expect cancelling of the job to trigger cancelling of all its could-potentially-still-run tasks.
	fetchCall := mocks.persist.EXPECT().FetchTasksOfJobInStatus(ctx, job,
		api.TaskStatusActive,
		api.TaskStatusQueued,
		api.TaskStatusSoftFailed,
	).Return(potentialRunTasks, nil)
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusCanceled).After(fetchCall)
	mocks.expectSaveTaskWithStatus(t, task3, api.TaskStatusCanceled).After(fetchCall)
	mocks.expectSaveJobWithStatus(t, job, api.JobStatusCanceled).After(fetchCall)

	mocks.expectBroadcastJobChange(job, api.JobStatusActive, api.JobStatusCancelRequested)
	mocks.expectBroadcastJobChange(job, api.JobStatusCancelRequested, api.JobStatusCanceled)

	mocks.expectBroadcastTaskChange(task2, api.TaskStatusFailed, api.TaskStatusCanceled)
	mocks.expectBroadcastTaskChange(task3, api.TaskStatusSoftFailed, api.TaskStatusCanceled)

	assert.NoError(t, sm.JobStatusChange(ctx, job, api.JobStatusCancelRequested, "someone wrote a unittest"))
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
	expectUpdate := api.JobUpdate{
		Id:             job.UUID,
		Name:           &job.Name,
		Updated:        job.UpdatedAt,
		PreviousStatus: &fromStatus,
		Status:         toStatus,
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
