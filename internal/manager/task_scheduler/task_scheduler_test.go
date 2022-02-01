// Package task_scheduler can choose which task to assign to a worker.
package task_scheduler

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
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

func TestNoTasks(t *testing.T) {
	db := persistence.CreateTestDB(t)
	ts := NewTaskScheduler(db)
	w := linuxWorker()

	task, err := ts.ScheduleTask(&w)
	assert.Nil(t, task)
	assert.NoError(t, err)
}

func TestOneJobOneTask(t *testing.T) {
	db := persistence.CreateTestDB(t)
	ts := NewTaskScheduler(db)
	w := linuxWorker()

	authTask := authorTestTask("the task", "blender-render")
	atj := authorTestJob("b6a1d859-122f-4791-8b78-b943329a9989", "simple-blender-render", authTask)
	job := constructTestJob(t, db, atj)

	task, err := ts.ScheduleTask(&w)
	assert.NoError(t, err)
	if task == nil {
		t.Fatal("task is nil")
	}
	assert.Equal(t, job.ID, task.JobID)
}

func TestOneJobThreeTasksByPrio(t *testing.T) {
	db := persistence.CreateTestDB(t)
	ts := NewTaskScheduler(db)
	w := linuxWorker()

	att1 := authorTestTask("1 low-prio task", "blender-render")
	att2 := authorTestTask("2 high-prio task", "render-preview")
	att2.Priority = 100
	att3 := authorTestTask("3 low-prio task", "blender-render")
	atj := authorTestJob(
		"1295757b-e668-4c49-8b89-f73db8270e42",
		"simple-blender-render",
		att1, att2, att3)

	job := constructTestJob(t, db, atj)

	task, err := ts.ScheduleTask(&w)
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
	db := persistence.CreateTestDB(t)
	ts := NewTaskScheduler(db)
	w := linuxWorker()

	att1 := authorTestTask("1 low-prio task", "blender-render")
	att2 := authorTestTask("2 high-prio task", "render-preview")
	att2.Priority = 100
	att2.Dependencies = []*job_compilers.AuthoredTask{&att1}
	att3 := authorTestTask("3 low-prio task", "blender-render")
	atj := authorTestJob(
		"1295757b-e668-4c49-8b89-f73db8270e42",
		"simple-blender-render",
		att1, att2, att3)
	job := constructTestJob(t, db, atj)

	task, err := ts.ScheduleTask(&w)
	assert.NoError(t, err)
	if task == nil {
		t.Fatal("task is nil")
	}
	assert.Equal(t, job.ID, task.JobID)
	assert.Equal(t, att1.Name, task.Name, "the first task should have been chosen")
}

// To test: worker with non-active state.
// Unlike Flamenco v2, this Manager shouldn't change a worker's status
// simply because it requests a task. New tasks for non-awake workers
// should be rejected.

// To test: blacklists

// To test: variable replacement

func constructTestJob(
	t *testing.T, db *persistence.DB, authoredJob job_compilers.AuthoredJob,
) *persistence.Job {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := db.StoreAuthoredJob(ctx, authoredJob)
	assert.NoError(t, err)

	dbJob, err := db.FetchJob(ctx, authoredJob.JobID)
	assert.NoError(t, err)

	// Queue the job.
	dbJob.Status = string(api.JobStatusQueued)
	err = db.SaveJobStatus(ctx, dbJob)
	assert.NoError(t, err)

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
		UUID:         uuid.NewString(),
		Priority:     50,
		Commands:     make([]job_compilers.AuthoredCommand, 0),
		Dependencies: dependencies,
	}
	return task
}

func linuxWorker() persistence.Worker {
	w := persistence.Worker{
		UUID:               "b13b8322-3e96-41c3-940a-3d581008a5f8",
		Name:               "Linux",
		Platform:           "linux",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management",
	}
	return w
}
