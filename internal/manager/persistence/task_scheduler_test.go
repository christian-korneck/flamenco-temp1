package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
)

const schedulerTestTimeout = 100 * time.Millisecond

func TestNoTasks(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	w := linuxWorker(t, db)

	task, err := db.ScheduleTask(ctx, &w)
	assert.Nil(t, task)
	assert.NoError(t, err)
}

func TestOneJobOneTask(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	w := linuxWorker(t, db)

	authTask := authorTestTask("the task", "blender")
	atj := authorTestJob("b6a1d859-122f-4791-8b78-b943329a9989", "simple-blender-render", authTask)
	job := constructTestJob(ctx, t, db, atj)

	task, err := db.ScheduleTask(ctx, &w)
	assert.NoError(t, err)

	// Check the returned task.
	if task == nil {
		t.Fatal("task is nil")
	}
	assert.Equal(t, job.ID, task.JobID)
	if task.WorkerID == nil {
		t.Fatal("no worker assigned to task")
	}
	assert.Equal(t, w.ID, *task.WorkerID, "task must be assigned to the requesting worker")

	// Check the task in the database.
	now := db.gormDB.NowFunc()
	dbTask, err := db.FetchTask(context.Background(), authTask.UUID)
	assert.NoError(t, err)
	if dbTask == nil {
		t.Fatal("task cannot be fetched from database")
	}
	if dbTask.WorkerID == nil {
		t.Fatal("no worker assigned to task")
	}
	assert.Equal(t, w.ID, *dbTask.WorkerID, "task must be assigned to the requesting worker")
	assert.WithinDuration(t, now, dbTask.LastTouchedAt, time.Second, "task must be 'touched' by the worker after scheduling")
}

func TestOneJobThreeTasksByPrio(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	w := linuxWorker(t, db)

	att1 := authorTestTask("1 low-prio task", "blender")
	att2 := authorTestTask("2 high-prio task", "ffmpeg")
	att2.Priority = 100
	att3 := authorTestTask("3 low-prio task", "blender")
	atj := authorTestJob(
		"1295757b-e668-4c49-8b89-f73db8270e42",
		"simple-blender-render",
		att1, att2, att3)

	job := constructTestJob(ctx, t, db, atj)

	task, err := db.ScheduleTask(ctx, &w)
	assert.NoError(t, err)
	if task == nil {
		t.Fatal("task is nil")
	}

	assert.Equal(t, job.ID, task.JobID)
	if task.Job == nil {
		t.Fatal("task.Job is nil")
	}

	assert.Equal(t, att2.Name, task.Name, "the high-prio task should have been chosen")
}

func TestOneJobThreeTasksByDependencies(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	w := linuxWorker(t, db)

	att1 := authorTestTask("1 low-prio task", "blender")
	att2 := authorTestTask("2 high-prio task", "ffmpeg")
	att2.Priority = 100
	att2.Dependencies = []*job_compilers.AuthoredTask{&att1}
	att3 := authorTestTask("3 low-prio task", "blender")
	atj := authorTestJob(
		"1295757b-e668-4c49-8b89-f73db8270e42",
		"simple-blender-render",
		att1, att2, att3)
	job := constructTestJob(ctx, t, db, atj)

	task, err := db.ScheduleTask(ctx, &w)
	assert.NoError(t, err)
	if task == nil {
		t.Fatal("task is nil")
	}
	assert.Equal(t, job.ID, task.JobID)
	assert.Equal(t, att1.Name, task.Name, "the first task should have been chosen")
}

func TestTwoJobsThreeTasks(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	w := linuxWorker(t, db)

	att1_1 := authorTestTask("1.1 low-prio task", "blender")
	att1_2 := authorTestTask("1.2 high-prio task", "ffmpeg")
	att1_2.Priority = 100
	att1_2.Dependencies = []*job_compilers.AuthoredTask{&att1_1}
	att1_3 := authorTestTask("1.3 low-prio task", "blender")
	atj1 := authorTestJob(
		"1295757b-e668-4c49-8b89-f73db8270e42",
		"simple-blender-render",
		att1_1, att1_2, att1_3)

	att2_1 := authorTestTask("2.1 low-prio task", "blender")
	att2_2 := authorTestTask("2.2 high-prio task", "ffmpeg")
	att2_2.Priority = 100
	att2_2.Dependencies = []*job_compilers.AuthoredTask{&att2_1}
	att2_3 := authorTestTask("2.3 highest-prio task", "blender")
	att2_3.Priority = 150
	atj2 := authorTestJob(
		"7180617b-da70-411c-8b38-b972ab2bae8d",
		"simple-blender-render",
		att2_1, att2_2, att2_3)
	atj2.Priority = 100 // Increase priority over job 1.

	constructTestJob(ctx, t, db, atj1)
	job2 := constructTestJob(ctx, t, db, atj2)

	task, err := db.ScheduleTask(ctx, &w)
	assert.NoError(t, err)
	if task == nil {
		t.Fatal("task is nil")
	}
	assert.Equal(t, job2.ID, task.JobID)
	assert.Equal(t, att2_3.Name, task.Name, "the 3rd task of the 2nd job should have been chosen")
}

func TestSomeButNotAllDependenciesCompleted(t *testing.T) {
	// There was a bug in the task scheduler query, where it would schedule a task
	// if any of its dependencies was completed (instead of all dependencies).
	// This test reproduces that problematic scenario.
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	att1 := authorTestTask("1.1 completed task", "blender")
	att2 := authorTestTask("1.2 queued task of unsupported type", "unsupported")
	att3 := authorTestTask("1.3 queued task with queued dependency", "ffmpeg")
	att3.Dependencies = []*job_compilers.AuthoredTask{&att1, &att2}

	atj := authorTestJob("1295757b-e668-4c49-8b89-f73db8270e42", "simple-blender-render", att1, att2, att3)
	constructTestJob(ctx, t, db, atj)

	// Complete the first task. The other two are still `queued`.
	setTaskStatus(t, db, att1.UUID, api.TaskStatusCompleted)

	w := linuxWorker(t, db)
	task, err := db.ScheduleTask(ctx, &w)
	assert.NoError(t, err)
	if task != nil {
		t.Fatalf("there should not be any task assigned, but received %q", task.Name)
	}
}

func TestAlreadyAssigned(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	w := linuxWorker(t, db)

	att1 := authorTestTask("1 low-prio task", "blender")
	att2 := authorTestTask("2 high-prio task", "ffmpeg")
	att2.Priority = 100
	att3 := authorTestTask("3 low-prio task", "blender")
	atj := authorTestJob(
		"1295757b-e668-4c49-8b89-f73db8270e42",
		"simple-blender-render",
		att1, att2, att3)

	constructTestJob(ctx, t, db, atj)

	// Assign the task to the worker, and mark it as Active.
	// This should make it get returned by the scheduler, even when there is
	// another, higher-prio task to be done.
	dbTask3, err := db.FetchTask(ctx, att3.UUID)
	assert.NoError(t, err)
	dbTask3.WorkerID = &w.ID
	dbTask3.Status = api.TaskStatusActive
	err = db.SaveTask(ctx, dbTask3)
	assert.NoError(t, err)

	task, err := db.ScheduleTask(ctx, &w)
	assert.NoError(t, err)
	if task == nil {
		t.Fatal("task is nil")
	}

	assert.Equal(t, att3.Name, task.Name, "the already-assigned task should have been chosen")
}

func TestAssignedToOtherWorker(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	w := linuxWorker(t, db)
	w2 := windowsWorker(t, db)

	att1 := authorTestTask("1 low-prio task", "blender")
	att2 := authorTestTask("2 high-prio task", "ffmpeg")
	att2.Priority = 100
	atj := authorTestJob(
		"1295757b-e668-4c49-8b89-f73db8270e42",
		"simple-blender-render",
		att1, att2)

	constructTestJob(ctx, t, db, atj)

	// Assign the high-prio task to the other worker. Because the task is queued,
	// it shouldn't matter which worker it's assigned to.
	dbTask2, err := db.FetchTask(ctx, att2.UUID)
	assert.NoError(t, err)
	dbTask2.WorkerID = &w2.ID
	dbTask2.Status = api.TaskStatusQueued
	err = db.SaveTask(ctx, dbTask2)
	assert.NoError(t, err)

	task, err := db.ScheduleTask(ctx, &w)
	assert.NoError(t, err)
	if task == nil {
		t.Fatal("task is nil")
	}

	assert.Equal(t, att2.Name, task.Name, "the high-prio task should have been chosen")
	assert.Equal(t, *task.WorkerID, w.ID, "the task should now be assigned to the worker it was scheduled for")
}

func TestPreviouslyFailed(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, schedulerTestTimeout)
	defer cancel()

	w := linuxWorker(t, db)

	att1 := authorTestTask("1 failed task", "blender")
	att2 := authorTestTask("2 expected task", "blender")
	atj := authorTestJob(
		"1295757b-e668-4c49-8b89-f73db8270e42",
		"simple-blender-render",
		att1, att2)
	job := constructTestJob(ctx, t, db, atj)

	// Mimick that this worker already failed the first task.
	tasks, err := db.FetchTasksOfJob(ctx, job)
	assert.NoError(t, err)
	numFailed, err := db.AddWorkerToTaskFailedList(ctx, tasks[0], &w)
	assert.NoError(t, err)
	assert.Equal(t, 1, numFailed)

	// This should assign the 2nd task.
	task, err := db.ScheduleTask(ctx, &w)
	assert.NoError(t, err)
	if task == nil {
		t.Fatal("task is nil")
	}
	assert.Equal(t, att2.Name, task.Name, "the second task should have been chosen")
}

// To test: blocklists

// To test: variable replacement

func constructTestJob(
	ctx context.Context, t *testing.T, db *DB, authoredJob job_compilers.AuthoredJob,
) *Job {
	err := db.StoreAuthoredJob(ctx, authoredJob)
	if err != nil {
		t.Fatalf("storing authored job: %v", err)
	}

	dbJob, err := db.FetchJob(ctx, authoredJob.JobID)
	if err != nil {
		t.Fatalf("fetching authored job: %v", err)
	}

	// Queue the job.
	dbJob.Status = api.JobStatusQueued
	err = db.SaveJobStatus(ctx, dbJob)
	if err != nil {
		t.Fatalf("queueing job: %v", err)
	}

	return dbJob
}

func authorTestJob(jobUUID, jobType string, tasks ...job_compilers.AuthoredTask) job_compilers.AuthoredJob {
	job := job_compilers.AuthoredJob{
		JobID:    jobUUID,
		Name:     "test job",
		JobType:  jobType,
		Priority: 50,
		Status:   api.JobStatusQueued,
		Created:  time.Now(),
		Tasks:    tasks,
	}
	return job
}

func authorTestTask(name, taskType string, dependencies ...*job_compilers.AuthoredTask) job_compilers.AuthoredTask {
	task := job_compilers.AuthoredTask{
		Name:         name,
		Type:         taskType,
		UUID:         uuid.New(),
		Priority:     50,
		Commands:     make([]job_compilers.AuthoredCommand, 0),
		Dependencies: dependencies,
	}
	return task
}

func setTaskStatus(t *testing.T, db *DB, taskUUID string, status api.TaskStatus) {
	ctx := context.Background()
	task, err := db.FetchTask(ctx, taskUUID)
	if err != nil {
		t.Fatal(err)
	}

	task.Status = status

	err = db.SaveTask(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
}

func linuxWorker(t *testing.T, db *DB) Worker {
	w := Worker{
		UUID:               "b13b8322-3e96-41c3-940a-3d581008a5f8",
		Name:               "Linux",
		Platform:           "linux",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management,misc",
	}

	err := db.gormDB.Save(&w).Error
	if err != nil {
		t.Logf("cannot save Linux worker: %v", err)
		t.FailNow()
	}

	return w
}

func windowsWorker(t *testing.T, db *DB) Worker {
	w := Worker{
		UUID:               "4f6ee45e-c8fc-4c31-bf5c-922f2415deb1",
		Name:               "Windows",
		Platform:           "windows",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management,misc",
	}

	err := db.gormDB.Save(&w).Error
	if err != nil {
		t.Logf("cannot save Windows worker: %v", err)
		t.FailNow()
	}

	return w
}
