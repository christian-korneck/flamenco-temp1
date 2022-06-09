package persistence

import (
	"testing"
	"time"

	"git.blender.org/flamenco/pkg/api"
	"github.com/stretchr/testify/assert"
)

// SPDX-License-Identifier: GPL-3.0-or-later

func TestFetchTimedOutTasks(t *testing.T) {
	ctx, close, db, job, _ := jobTasksTestFixtures(t)
	defer close()

	tasks, err := db.FetchTasksOfJob(ctx, job)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	now := db.gormDB.NowFunc()
	deadline := now.Add(-5 * time.Minute)

	// Mark the task as last touched before the deadline, i.e. old enough for a timeout.
	task := tasks[0]
	task.LastTouchedAt = deadline.Add(-1 * time.Minute)
	assert.NoError(t, db.SaveTask(ctx, task))

	w := createWorker(ctx, t, db)
	assert.NoError(t, db.TaskAssignToWorker(ctx, task, w))

	// The task should still not be returned, as it's not in 'active' state.
	timedout, err := db.FetchTimedOutTasks(ctx, deadline)
	assert.NoError(t, err)
	assert.Empty(t, timedout)

	// Mark as Active:
	task.Status = api.TaskStatusActive
	assert.NoError(t, db.SaveTask(ctx, task))

	// Now it should time out:
	timedout, err = db.FetchTimedOutTasks(ctx, deadline)
	assert.NoError(t, err)
	if assert.Len(t, timedout, 1) {
		// Other fields will be different, like the 'UpdatedAt' field -- this just
		// tests that the expected task is returned.
		assert.Equal(t, task.UUID, timedout[0].UUID)
		assert.Equal(t, job, task.Job, "the job should be included in the result as well")
	}
}
