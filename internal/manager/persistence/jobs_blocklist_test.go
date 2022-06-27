package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// SPDX-License-Identifier: GPL-3.0-or-later

func TestAddWorkerToJobBlocklist(t *testing.T) {
	ctx, close, db, job, _ := jobTasksTestFixtures(t)
	defer close()

	worker := createWorker(ctx, t, db)

	{
		// Add a worker to the block list.
		err := db.AddWorkerToJobBlocklist(ctx, job, worker, "blender")
		assert.NoError(t, err)

		list := []JobBlock{}
		tx := db.gormDB.Model(&JobBlock{}).Scan(&list)
		assert.NoError(t, tx.Error)
		if assert.Len(t, list, 1) {
			entry := list[0]
			assert.Equal(t, entry.JobID, job.ID)
			assert.Equal(t, entry.WorkerID, worker.ID)
			assert.Equal(t, entry.TaskType, "blender")
		}
	}

	{
		// Adding the same worker again should be a no-op.
		err := db.AddWorkerToJobBlocklist(ctx, job, worker, "blender")
		assert.NoError(t, err)

		list := []JobBlock{}
		tx := db.gormDB.Model(&JobBlock{}).Scan(&list)
		assert.NoError(t, tx.Error)
		assert.Len(t, list, 1, "No new entry should have been created")
	}
}

func TestFetchJobBlocklist(t *testing.T) {
	ctx, close, db, job, _ := jobTasksTestFixtures(t)
	defer close()

	// Add a worker to the block list.
	worker := createWorker(ctx, t, db)
	err := db.AddWorkerToJobBlocklist(ctx, job, worker, "blender")
	assert.NoError(t, err)

	list, err := db.FetchJobBlocklist(ctx, job.UUID)
	assert.NoError(t, err)

	if assert.Len(t, list, 1) {
		entry := list[0]
		assert.Equal(t, entry.JobID, job.ID)
		assert.Equal(t, entry.WorkerID, worker.ID)
		assert.Equal(t, entry.TaskType, "blender")

		assert.Nil(t, entry.Job, "should NOT fetch the entire job")
		assert.NotNil(t, entry.Worker, "SHOULD fetch the entire worker")
	}
}

func TestRemoveFromJobBlocklist(t *testing.T) {
	ctx, close, db, job, _ := jobTasksTestFixtures(t)
	defer close()

	// Add a worker and some entries to the block list.
	worker := createWorker(ctx, t, db)
	err := db.AddWorkerToJobBlocklist(ctx, job, worker, "blender")
	assert.NoError(t, err)
	err = db.AddWorkerToJobBlocklist(ctx, job, worker, "ffmpeg")
	assert.NoError(t, err)

	// Remove an entry.
	err = db.RemoveFromJobBlocklist(ctx, job.UUID, worker.UUID, "ffmpeg")
	assert.NoError(t, err)

	// Check that the other entry is still there.
	list, err := db.FetchJobBlocklist(ctx, job.UUID)
	assert.NoError(t, err)

	if assert.Len(t, list, 1) {
		entry := list[0]
		assert.Equal(t, entry.JobID, job.ID)
		assert.Equal(t, entry.WorkerID, worker.ID)
		assert.Equal(t, entry.TaskType, "blender")
	}
}
func TestWorkersLeftToRun(t *testing.T) {
	ctx, close, db, job, _ := jobTasksTestFixtures(t)
	defer close()

	// No workers.
	left, err := db.WorkersLeftToRun(ctx, job, "blender")
	assert.NoError(t, err)
	assert.Empty(t, left)

	worker1 := createWorker(ctx, t, db)
	worker2 := createWorkerFrom(ctx, t, db, *worker1)

	uuidMap := func(workers ...*Worker) map[string]bool {
		theMap := map[string]bool{}
		for _, worker := range workers {
			theMap[worker.UUID] = true
		}
		return theMap
	}

	// Two workers, no blocklist.
	left, err = db.WorkersLeftToRun(ctx, job, "blender")
	if assert.NoError(t, err) {
		assert.Equal(t, uuidMap(worker1, worker2), left)
	}

	// Two workers, one blocked.
	_ = db.AddWorkerToJobBlocklist(ctx, job, worker1, "blender")
	left, err = db.WorkersLeftToRun(ctx, job, "blender")
	if assert.NoError(t, err) {
		assert.Equal(t, uuidMap(worker2), left)
	}

	// Two workers, both blocked.
	_ = db.AddWorkerToJobBlocklist(ctx, job, worker2, "blender")
	left, err = db.WorkersLeftToRun(ctx, job, "blender")
	assert.NoError(t, err)
	assert.Empty(t, left)

	// Two workers, unknown job.
	fakeJob := Job{Model: Model{ID: 327}}
	left, err = db.WorkersLeftToRun(ctx, &fakeJob, "blender")
	if assert.NoError(t, err) {
		assert.Equal(t, uuidMap(worker1, worker2), left)
	}
}

func TestCountTaskFailuresOfWorker(t *testing.T) {
	ctx, close, db, dbJob, authoredJob := jobTasksTestFixtures(t)
	defer close()

	task0, _ := db.FetchTask(ctx, authoredJob.Tasks[0].UUID)
	task1, _ := db.FetchTask(ctx, authoredJob.Tasks[1].UUID)
	task2, _ := db.FetchTask(ctx, authoredJob.Tasks[2].UUID)

	// Sanity check on the test data.
	assert.Equal(t, "blender", task0.Type)
	assert.Equal(t, "blender", task1.Type)
	assert.Equal(t, "ffmpeg", task2.Type)

	worker1 := createWorker(ctx, t, db)
	worker2 := createWorkerFrom(ctx, t, db, *worker1)

	// Store some failures for different tasks
	_, _ = db.AddWorkerToTaskFailedList(ctx, task0, worker1)
	_, _ = db.AddWorkerToTaskFailedList(ctx, task1, worker1)
	_, _ = db.AddWorkerToTaskFailedList(ctx, task1, worker2)
	_, _ = db.AddWorkerToTaskFailedList(ctx, task2, worker1)

	// Multiple failures.
	numBlender1, err := db.CountTaskFailuresOfWorker(ctx, dbJob, worker1, "blender")
	if assert.NoError(t, err) {
		assert.Equal(t, 2, numBlender1)
	}

	// Single failure, but multiple tasks exist of this type.
	numBlender2, err := db.CountTaskFailuresOfWorker(ctx, dbJob, worker2, "blender")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, numBlender2)
	}

	// Single failure, only one task of this type exists.
	numFFMpeg1, err := db.CountTaskFailuresOfWorker(ctx, dbJob, worker1, "ffmpeg")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, numFFMpeg1)
	}

	// No failure.
	numFFMpeg2, err := db.CountTaskFailuresOfWorker(ctx, dbJob, worker2, "ffmpeg")
	if assert.NoError(t, err) {
		assert.Equal(t, 0, numFFMpeg2)
	}
}
