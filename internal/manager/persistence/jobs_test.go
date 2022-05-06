// Package persistence provides the database interface for Flamenco Manager.
package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/pkg/api"
)

func TestStoreAuthoredJob(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, 1*time.Second)
	defer cancel()

	job := createTestAuthoredJobWithTasks()
	err := db.StoreAuthoredJob(ctx, job)
	assert.NoError(t, err)

	fetchedJob, err := db.FetchJob(ctx, job.JobID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedJob)

	// Test contents of fetched job
	assert.Equal(t, job.JobID, fetchedJob.UUID)
	assert.Equal(t, job.Name, fetchedJob.Name)
	assert.Equal(t, job.JobType, fetchedJob.JobType)
	assert.Equal(t, job.Priority, fetchedJob.Priority)
	assert.Equal(t, api.JobStatusUnderConstruction, fetchedJob.Status)
	assert.EqualValues(t, map[string]interface{}(job.Settings), fetchedJob.Settings)
	assert.EqualValues(t, map[string]string(job.Metadata), fetchedJob.Metadata)

	// Fetch tasks of job.
	var dbJob Job
	tx := db.gormDB.Where(&Job{UUID: job.JobID}).Find(&dbJob)
	assert.NoError(t, tx.Error)
	var tasks []Task
	tx = db.gormDB.Where("job_id = ?", dbJob.ID).Find(&tasks)
	assert.NoError(t, tx.Error)

	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// TODO: test task contents.
	assert.Equal(t, api.TaskStatusQueued, tasks[0].Status)
	assert.Equal(t, api.TaskStatusQueued, tasks[1].Status)
	assert.Equal(t, api.TaskStatusQueued, tasks[2].Status)
}

func TestJobHasTasksInStatus(t *testing.T) {
	ctx, close, db, job, _ := jobTasksTestFixtures(t)
	defer close()

	hasTasks, err := db.JobHasTasksInStatus(ctx, job, api.TaskStatusQueued)
	assert.NoError(t, err)
	assert.True(t, hasTasks, "expected freshly-created job to have queued tasks")

	hasTasks, err = db.JobHasTasksInStatus(ctx, job, api.TaskStatusActive)
	assert.NoError(t, err)
	assert.False(t, hasTasks, "expected freshly-created job to have no active tasks")
}

func TestCountTasksOfJobInStatus(t *testing.T) {
	ctx, close, db, job, authoredJob := jobTasksTestFixtures(t)
	defer close()

	numQueued, numTotal, err := db.CountTasksOfJobInStatus(ctx, job, api.TaskStatusQueued)
	assert.NoError(t, err)
	assert.Equal(t, 3, numQueued)
	assert.Equal(t, 3, numTotal)

	// Make one task failed.
	task, err := db.FetchTask(ctx, authoredJob.Tasks[0].UUID)
	assert.NoError(t, err)
	task.Status = api.TaskStatusFailed
	assert.NoError(t, db.SaveTask(ctx, task))

	numQueued, numTotal, err = db.CountTasksOfJobInStatus(ctx, job, api.TaskStatusQueued)
	assert.NoError(t, err)
	assert.Equal(t, 2, numQueued)
	assert.Equal(t, 3, numTotal)

	numFailed, numTotal, err := db.CountTasksOfJobInStatus(ctx, job, api.TaskStatusFailed)
	assert.NoError(t, err)
	assert.Equal(t, 1, numFailed)
	assert.Equal(t, 3, numTotal)

	numActive, numTotal, err := db.CountTasksOfJobInStatus(ctx, job, api.TaskStatusActive)
	assert.NoError(t, err)
	assert.Equal(t, 0, numActive)
	assert.Equal(t, 3, numTotal)
}

func TestFetchJobsInStatus(t *testing.T) {
	ctx, close, db, job1, _ := jobTasksTestFixtures(t)
	defer close()

	ajob2 := createTestAuthoredJob("1f08e20b-ce24-41c2-b237-36120bd69fc6")
	ajob3 := createTestAuthoredJob("3ac2dbb4-0c34-410e-ad3b-652e6d7e65a5")
	job2 := persistAuthoredJob(t, ctx, db, ajob2)
	job3 := persistAuthoredJob(t, ctx, db, ajob3)

	// Sanity check
	if !assert.Equal(t, api.JobStatusUnderConstruction, job1.Status) {
		return
	}

	// Query single status
	jobs, err := db.FetchJobsInStatus(ctx, api.JobStatusUnderConstruction)
	assert.NoError(t, err)
	assert.Equal(t, []*Job{job1, job2, job3}, jobs)

	// Query two statuses, where only one matches all jobs.
	jobs, err = db.FetchJobsInStatus(ctx, api.JobStatusCanceled, api.JobStatusUnderConstruction)
	assert.NoError(t, err)
	assert.Equal(t, []*Job{job1, job2, job3}, jobs)

	// Update a job status, query for two of the three used statuses.
	job1.Status = api.JobStatusQueued
	assert.NoError(t, db.SaveJobStatus(ctx, job1))
	job2.Status = api.JobStatusRequeued
	assert.NoError(t, db.SaveJobStatus(ctx, job2))

	jobs, err = db.FetchJobsInStatus(ctx, api.JobStatusQueued, api.JobStatusUnderConstruction)
	assert.NoError(t, err)
	assert.Equal(t, []*Job{job1, job3}, jobs)
}

func TestFetchTasksOfJobInStatus(t *testing.T) {
	ctx, close, db, job, authoredJob := jobTasksTestFixtures(t)
	defer close()

	allTasks, err := db.FetchTasksOfJob(ctx, job)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, job, allTasks[0].Job, "FetchTasksOfJob should set job pointer")

	tasks, err := db.FetchTasksOfJobInStatus(ctx, job, api.TaskStatusQueued)
	assert.NoError(t, err)
	assert.Equal(t, allTasks, tasks)
	assert.Equal(t, job, tasks[0].Job, "FetchTasksOfJobInStatus should set job pointer")

	// Make one task failed.
	task, err := db.FetchTask(ctx, authoredJob.Tasks[0].UUID)
	assert.NoError(t, err)
	task.Status = api.TaskStatusFailed
	assert.NoError(t, db.SaveTask(ctx, task))

	tasks, err = db.FetchTasksOfJobInStatus(ctx, job, api.TaskStatusQueued)
	assert.NoError(t, err)
	assert.Equal(t, []*Task{allTasks[1], allTasks[2]}, tasks)

	// Check the failed task. This cannot directly compare to `allTasks[0]`
	// because saving the task above changed some of its fields.
	tasks, err = db.FetchTasksOfJobInStatus(ctx, job, api.TaskStatusFailed)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, allTasks[0].ID, tasks[0].ID)

	tasks, err = db.FetchTasksOfJobInStatus(ctx, job, api.TaskStatusActive)
	assert.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestTaskAssignToWorker(t *testing.T) {
	ctx, close, db, _, authoredJob := jobTasksTestFixtures(t)
	defer close()

	task, err := db.FetchTask(ctx, authoredJob.Tasks[1].UUID)
	assert.NoError(t, err)

	w := createWorker(ctx, t, db)
	assert.NoError(t, db.TaskAssignToWorker(ctx, task, w))

	if task.Worker == nil {
		t.Error("task.Worker == nil")
	} else {
		assert.Equal(t, w, task.Worker)
	}
	if task.WorkerID == nil {
		t.Error("task.WorkerID == nil")
	} else {
		assert.Equal(t, w.ID, *task.WorkerID)
	}
}

func TestFetchTasksOfWorkerInStatus(t *testing.T) {
	ctx, close, db, _, authoredJob := jobTasksTestFixtures(t)
	defer close()

	task, err := db.FetchTask(ctx, authoredJob.Tasks[1].UUID)
	assert.NoError(t, err)

	w := createWorker(ctx, t, db)
	assert.NoError(t, db.TaskAssignToWorker(ctx, task, w))

	tasks, err := db.FetchTasksOfWorkerInStatus(ctx, w, task.Status)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1, "worker should have one task in status %q", task.Status)
	assert.Equal(t, task.ID, tasks[0].ID)
	assert.Equal(t, task.UUID, tasks[0].UUID)

	assert.NotEqual(t, api.TaskStatusCanceled, task.Status)
	tasks, err = db.FetchTasksOfWorkerInStatus(ctx, w, api.TaskStatusCanceled)
	assert.NoError(t, err)
	assert.Empty(t, tasks, "worker should have no task in status %q", w)
}

func createTestAuthoredJobWithTasks() job_compilers.AuthoredJob {
	task1 := job_compilers.AuthoredTask{
		Name: "render-1-3",
		Type: "blender",
		UUID: "db1f5481-4ef5-4084-8571-8460c547ecaa",
		Commands: []job_compilers.AuthoredCommand{
			{
				Name: "blender-render",
				Parameters: job_compilers.AuthoredCommandParameters{
					"exe":       "{blender}",
					"blendfile": "/path/to/file.blend",
					"args": []interface{}{
						"--render-output", "/path/to/output/######.png",
						"--render-format", "PNG",
						"--render-frame", "1-3",
					},
				}},
		},
	}

	task2 := task1
	task2.Name = "render-4-6"
	task2.UUID = "d75ac779-151b-4bc2-b8f1-d153a9c4ac69"
	task2.Commands[0].Parameters["frames"] = "4-6"

	task3 := job_compilers.AuthoredTask{
		Name: "preview-video",
		Type: "ffmpeg",
		UUID: "4915fb05-72f5-463e-a2f4-7efdb2584a1e",
		Commands: []job_compilers.AuthoredCommand{
			{
				Name: "merge-frames-to-video",
				Parameters: job_compilers.AuthoredCommandParameters{
					"images":       "/path/to/output/######.png",
					"output":       "/path/to/output/preview.mkv",
					"ffmpegParams": "-c:v hevc -crf 31",
				}},
		},
		Dependencies: []*job_compilers.AuthoredTask{&task1, &task2},
	}

	return createTestAuthoredJob("263fd47e-b9f8-4637-b726-fd7e47ecfdae", task1, task2, task3)
}

func createTestAuthoredJob(jobID string, tasks ...job_compilers.AuthoredTask) job_compilers.AuthoredJob {
	job := job_compilers.AuthoredJob{
		JobID:    jobID,
		Name:     "Test job",
		Status:   api.JobStatusUnderConstruction,
		Priority: 50,
		Settings: job_compilers.JobSettings{
			"frames":     "1-6",
			"chunk_size": 3.0, // The roundtrip to JSON in the database can make this a float.
		},
		Metadata: job_compilers.JobMetadata{
			"author":  "Sybren",
			"project": "Sprite Fright",
		},
		Tasks: tasks,
	}

	return job
}

func persistAuthoredJob(t *testing.T, ctx context.Context, db *DB, authoredJob job_compilers.AuthoredJob) *Job {
	err := db.StoreAuthoredJob(ctx, authoredJob)
	if err != nil {
		t.Fatalf("error storing authored job in DB: %v", err)
	}

	dbJob, err := db.FetchJob(ctx, authoredJob.JobID)
	if err != nil {
		t.Fatalf("error fetching job from DB: %v", err)
	}
	if dbJob == nil {
		t.Fatalf("nil job obtained from DB but with no error!")
	}
	return dbJob
}

func jobTasksTestFixtures(t *testing.T) (context.Context, context.CancelFunc, *DB, *Job, job_compilers.AuthoredJob) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)

	authoredJob := createTestAuthoredJobWithTasks()
	dbJob := persistAuthoredJob(t, ctx, db, authoredJob)

	return ctx, cancel, db, dbJob, authoredJob
}

func createWorker(ctx context.Context, t *testing.T, db *DB) *Worker {
	w := Worker{
		UUID:               "f0a123a9-ab05-4ce2-8577-94802cfe74a4",
		Name:               "дрон",
		Address:            "fe80::5054:ff:fede:2ad7",
		LastActivity:       "",
		Platform:           "linux",
		Software:           "3.0",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management",
	}

	err := db.CreateWorker(ctx, &w)
	if err != nil {
		t.Fatalf("error creating worker: %v", err)
	}
	assert.NoError(t, err)

	fetchedWorker, err := db.FetchWorker(ctx, w.UUID)
	if err != nil {
		t.Fatalf("error fetching worker: %v", err)
	}
	if fetchedWorker == nil {
		t.Fatal("fetched worker is nil, but no error returned")
	}

	return fetchedWorker
}
