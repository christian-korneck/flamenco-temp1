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
		assert.Equal(t, job, timedout[0].Job, "the job should be included in the result as well")
		assert.Equal(t, w, timedout[0].Worker, "the worker should be included in the result as well")
	}
}

func TestFetchTimedOutWorkers(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, 1*time.Second)
	defer cancel()

	timeoutDeadline := mustParseTime("2022-06-07T11:14:47+02:00")
	beforeDeadline := timeoutDeadline.Add(-10 * time.Second)
	afterDeadline := timeoutDeadline.Add(10 * time.Second)

	worker0 := Worker{ // Offline, so should not time out.
		UUID:       "c7b4d1d5-0a96-4e19-993f-028786d3d2c1",
		Name:       "дрон 0",
		Status:     api.WorkerStatusOffline,
		LastSeenAt: beforeDeadline,
	}
	worker1 := Worker{ // Awake and timed out.
		UUID:       "bafc098f-2760-40c6-9a45-a4f980389a9a",
		Name:       "дрон 1",
		Status:     api.WorkerStatusAwake,
		LastSeenAt: beforeDeadline,
	}
	worker2 := Worker{ // Starting and timed out.
		UUID:       "67afa6e6-406d-4224-87d9-82abde7f9d6a",
		Name:       "дрон 2",
		Status:     api.WorkerStatusStarting,
		LastSeenAt: beforeDeadline,
	}
	worker3 := Worker{ // Asleep and timed out.
		UUID:       "12a0bb9a-515b-440a-922a-fd6765fd89a4",
		Name:       "дрон 3",
		Status:     api.WorkerStatusAsleep,
		LastSeenAt: beforeDeadline,
	}
	worker4 := Worker{ // Awake and not timed out.
		UUID:       "aecfc9c8-ebf5-4be3-9091-99b6961a8b6e",
		Name:       "дрон 4",
		Status:     api.WorkerStatusAwake,
		LastSeenAt: afterDeadline,
	}
	workers := []*Worker{&worker0, &worker1, &worker2, &worker3, &worker4}
	for _, worker := range workers {
		err := db.CreateWorker(ctx, worker)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	timedout, err := db.FetchTimedOutWorkers(ctx, timeoutDeadline)
	if assert.NoError(t, err) && assert.Len(t, timedout, 3) {
		assert.Equal(t, worker1.UUID, timedout[0].UUID)
		assert.Equal(t, worker2.UUID, timedout[1].UUID)
		assert.Equal(t, worker3.UUID, timedout[2].UUID)
	}
}
