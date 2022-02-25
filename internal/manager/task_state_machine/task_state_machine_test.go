package task_state_machine

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/task_state_machine/mocks"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type StateMachineMocks struct {
	persist *mocks.MockPersistenceService
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
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusActive))
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
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusCompleted).Return(1, 3, nil) // 1 of 3 complete.
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusCompleted))

	// Second task hickup: T: active > soft-failed --> J: active > active
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusSoftFailed)
	assert.NoError(t, sm.TaskStatusChange(ctx, task2, api.TaskStatusSoftFailed))

	// Second task completing: T: soft-failed > completed --> J: active > active
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusCompleted)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusCompleted).Return(2, 3, nil) // 2 of 3 complete.
	assert.NoError(t, sm.TaskStatusChange(ctx, task2, api.TaskStatusCompleted))

	// Third task completing: T: active > completed --> J: active > completed
	mocks.expectSaveTaskWithStatus(t, task3, api.TaskStatusCompleted)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusCompleted).Return(3, 3, nil) // 3 of 3 complete.
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusCompleted)
	assert.NoError(t, sm.TaskStatusChange(ctx, task3, api.TaskStatusCompleted))
}

func TestTaskStatusChangeQueuedToFailed(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: queued > failed (1% task failure) --> J: queued > active
	task := taskWithStatus(api.JobStatusQueued, api.TaskStatusQueued)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusFailed)
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusActive)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusFailed).Return(1, 100, nil) // 1 out of 100 failed.
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusFailed))
}

func TestTaskStatusChangeActiveToFailedFailJob(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: active > failed (10% task failure) --> J: active > failed
	task := taskWithStatus(api.JobStatusActive, api.TaskStatusActive)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusFailed)
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusFailed)
	mocks.persist.EXPECT().CountTasksOfJobInStatus(ctx, task.Job, api.TaskStatusFailed).Return(10, 100, nil) // 10 out of 100 failed.
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusFailed))
}

func TestTaskStatusChangeRequeueOnCompletedJob(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: completed > queued --> J: completed > queued
	task := taskWithStatus(api.JobStatusCompleted, api.TaskStatusCompleted)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusQueued)
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusQueued)
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusQueued))
}

func TestTaskStatusChangeCancelSingleTask(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	task := taskWithStatus(api.JobStatusCancelRequested, api.TaskStatusCancelRequested)
	task2 := taskOfSameJob(task, api.TaskStatusCancelRequested)
	job := task.Job

	// T1: cancel-requested > cancelled --> J: cancel-requested > cancel-requested
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusCanceled)
	mocks.persist.EXPECT().JobHasTasksInStatus(ctx, job, api.TaskStatusCancelRequested).Return(true, nil)
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusCanceled))

	// T2: cancel-requested > cancelled --> J: cancel-requested > canceled
	mocks.expectSaveTaskWithStatus(t, task2, api.TaskStatusCanceled)
	mocks.persist.EXPECT().JobHasTasksInStatus(ctx, job, api.TaskStatusCancelRequested).Return(false, nil)
	mocks.expectSaveJobWithStatus(t, job, api.JobStatusCanceled)
	assert.NoError(t, sm.TaskStatusChange(ctx, task2, api.TaskStatusCanceled))
}

func TestTaskStatusChangeUnknownStatus(t *testing.T) {
	mockCtrl, ctx, sm, mocks := taskStateMachineTestFixtures(t)
	defer mockCtrl.Finish()

	// T: queued > borked --> saved to DB but otherwise ignored
	task := taskWithStatus(api.JobStatusQueued, api.TaskStatusQueued)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatus("borked"))
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatus("borked")))
}

func mockedTaskStateMachine(mockCtrl *gomock.Controller) (*StateMachine, *StateMachineMocks) {
	mocks := StateMachineMocks{
		persist: mocks.NewMockPersistenceService(mockCtrl),
	}
	sm := NewStateMachine(mocks.persist)
	return sm, &mocks
}

func (m *StateMachineMocks) expectSaveTaskWithStatus(
	t *testing.T,
	task *persistence.Task,
	expectTaskStatus api.TaskStatus,
) {
	m.persist.EXPECT().
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
) {
	m.persist.EXPECT().
		SaveJobStatus(gomock.Any(), job).
		DoAndReturn(func(ctx context.Context, savedJob *persistence.Job) error {
			assert.Equal(t, expectJobStatus, savedJob.Status)
			return nil
		})
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
