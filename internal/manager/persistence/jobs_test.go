// Package persistence provides the database interface for Flamenco Manager.
package persistence

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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
	"golang.org/x/net/context"
)

func TestStoreAuthoredJob(t *testing.T) {
	db := CreateTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
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
	ctx, db, job, _ := jobTasksTestFixtures(t)

	hasTasks, err := db.JobHasTasksInStatus(ctx, job, api.TaskStatusQueued)
	assert.NoError(t, err)
	assert.True(t, hasTasks, "expected freshly-created job to have queued tasks")

	hasTasks, err = db.JobHasTasksInStatus(ctx, job, api.TaskStatusActive)
	assert.NoError(t, err)
	assert.False(t, hasTasks, "expected freshly-created job to have no active tasks")
}

func TestCountTasksOfJobInStatus(t *testing.T) {
	ctx, db, job, authoredJob := jobTasksTestFixtures(t)

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

func TestUpdateJobsTaskStatuses(t *testing.T) {
	ctx, db, job, authoredJob := jobTasksTestFixtures(t)

	err := db.UpdateJobsTaskStatuses(ctx, job, api.TaskStatusSoftFailed, "testing æctivity")
	assert.NoError(t, err)

	numSoftFailed, numTotal, err := db.CountTasksOfJobInStatus(ctx, job, api.TaskStatusSoftFailed)
	assert.NoError(t, err)
	assert.Equal(t, 3, numSoftFailed, "all tasks should have had their status changed")
	assert.Equal(t, 3, numTotal)

	task, err := db.FetchTask(ctx, authoredJob.Tasks[0].UUID)
	assert.NoError(t, err)
	assert.Equal(t, "testing æctivity", task.Activity)

	// Empty status should be rejected.
	err = db.UpdateJobsTaskStatuses(ctx, job, "", "testing empty status")
	assert.Error(t, err)

	numEmpty, _, err := db.CountTasksOfJobInStatus(ctx, job, "")
	assert.NoError(t, err)
	assert.Equal(t, 0, numEmpty, "tasks should not have their status changed")

	numSoftFailed, _, err = db.CountTasksOfJobInStatus(ctx, job, api.TaskStatusSoftFailed)
	assert.NoError(t, err)
	assert.Equal(t, 3, numSoftFailed, "all tasks should still be soft-failed")
}

func TestUpdateJobsTaskStatusesConditional(t *testing.T) {
	ctx, db, job, authoredJob := jobTasksTestFixtures(t)

	getTask := func(taskIndex int) *Task {
		task, err := db.FetchTask(ctx, authoredJob.Tasks[taskIndex].UUID)
		if err != nil {
			t.Fatalf("Fetching task %d: %v", taskIndex, err)
		}
		return task
	}

	setTaskStatus := func(taskIndex int, taskStatus api.TaskStatus) {
		task := getTask(taskIndex)
		task.Status = taskStatus
		if err := db.SaveTask(ctx, task); err != nil {
			t.Fatalf("Setting task %d to status %s: %v", taskIndex, taskStatus, err)
		}
	}

	setTaskStatus(0, api.TaskStatusFailed)
	setTaskStatus(1, api.TaskStatusCompleted)
	setTaskStatus(2, api.TaskStatusActive)

	err := db.UpdateJobsTaskStatusesConditional(ctx, job,
		[]api.TaskStatus{api.TaskStatusFailed, api.TaskStatusActive},
		api.TaskStatusCancelRequested, "some activity")
	assert.NoError(t, err)

	// Task statuses should have updated for tasks 0 and 2.
	assert.Equal(t, api.TaskStatusCancelRequested, getTask(0).Status)
	assert.Equal(t, api.TaskStatusCompleted, getTask(1).Status)
	assert.Equal(t, api.TaskStatusCancelRequested, getTask(2).Status)

	err = db.UpdateJobsTaskStatusesConditional(ctx, job,
		[]api.TaskStatus{api.TaskStatusFailed, api.TaskStatusActive},
		"", "empty task status should be disallowed")
	assert.Error(t, err)

	// Task statuses should remain unchanged.
	assert.Equal(t, api.TaskStatusCancelRequested, getTask(0).Status)
	assert.Equal(t, api.TaskStatusCompleted, getTask(1).Status)
	assert.Equal(t, api.TaskStatusCancelRequested, getTask(2).Status)

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

	job := job_compilers.AuthoredJob{
		JobID:    "263fd47e-b9f8-4637-b726-fd7e47ecfdae",
		Name:     "Test job",
		Status:   api.JobStatusUnderConstruction,
		Priority: 50,
		Settings: job_compilers.JobSettings{
			"frames":     "1-6",
			"chunk_size": 3.0, // The roundtrip to JSON in PostgreSQL can make this a float.
		},
		Metadata: job_compilers.JobMetadata{
			"author":  "Sybren",
			"project": "Sprite Fright",
		},
		Tasks: []job_compilers.AuthoredTask{task1, task2, task3},
	}

	return job
}

func jobTasksTestFixtures(t *testing.T) (context.Context, *DB, *Job, job_compilers.AuthoredJob) {
	db := CreateTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	authoredJob := createTestAuthoredJobWithTasks()
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

	return ctx, db, dbJob, authoredJob
}
