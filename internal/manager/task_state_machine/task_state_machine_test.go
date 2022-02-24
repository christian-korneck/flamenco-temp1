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
		UUID:  "testtask-f474-4e28-aeea-8cbaf2fc96a5",

		JobID: job.ID,
		Job:   &job,

		Status: taskStatus,
	}

	return &task
}

func TestTaskStatusChange(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := context.Background()

	sm, mocks := mockedTaskStateMachine(mockCtrl)

	task := taskWithStatus(api.JobStatusQueued, api.TaskStatusQueued)
	mocks.expectSaveTaskWithStatus(t, task, api.TaskStatusActive)
	mocks.expectSaveJobWithStatus(t, task.Job, api.JobStatusActive)
	assert.NoError(t, sm.TaskStatusChange(ctx, task, api.TaskStatusActive))
}
