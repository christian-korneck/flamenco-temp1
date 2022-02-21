package api_impl

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

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

func TestTaskUpdate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	// Construct the JSON request object.
	s := func(value string) *string { return &value }
	ts := func(value api.TaskStatus) *api.TaskStatus { return &value }
	taskUpdate := api.TaskUpdateJSONRequestBody{
		Activity:   s("testing"),
		Log:        s("line1\nline2\n"),
		TaskStatus: ts(api.TaskStatusFailed),
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
	}

	// Expect the task to be fetched.
	mf.persistence.EXPECT().FetchTask(gomock.Any(), taskID).Return(&mockTask, nil)

	// Expect the task to be saved.
	var savedTask persistence.Task
	mf.persistence.EXPECT().SaveTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task *persistence.Task) error {
			savedTask = *task
			return nil
		})
	// Expect the log to be written.
	mf.logStorage.EXPECT().Write(gomock.Any(), jobID, taskID, "line1\nline2\n")

	// Do the call.
	echoCtx := mf.prepareMockedJSONRequest(&worker, taskUpdate)
	err := mf.flamenco.TaskUpdate(echoCtx, taskID)

	// Check the saved task.
	assert.NoError(t, err)
	assert.Equal(t, mockTask.UUID, savedTask.UUID)
}
