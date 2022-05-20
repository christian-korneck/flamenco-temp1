// Package persistence provides the database interface for Flamenco Manager.
package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
)

func TestSimpleQuery(t *testing.T) {
	ctx, close, db, job, _ := jobTasksTestFixtures(t)
	defer close()

	// Sanity check.
	if !assert.Equal(t, api.JobStatusUnderConstruction, job.Status, "check job status is as expected") {
		t.FailNow()
	}

	// Check empty result when querying for other status.
	result, err := db.QueryJobs(ctx, api.JobsQuery{
		StatusIn: &[]api.JobStatus{api.JobStatusActive, api.JobStatusCanceled},
	})
	assert.NoError(t, err)
	assert.Len(t, result, 0)

	// Check job was returned properly on correct status.
	result, err = db.QueryJobs(ctx, api.JobsQuery{
		StatusIn: &[]api.JobStatus{api.JobStatusUnderConstruction, api.JobStatusCanceled},
	})
	assert.NoError(t, err)
	if !assert.Len(t, result, 1) {
		t.FailNow()
	}
	assert.Equal(t, job.ID, result[0].ID)

}

func TestQueryMetadata(t *testing.T) {
	ctx, close, db := persistenceTestFixtures(t, 0)
	defer close()

	testJob := persistAuthoredJob(t, ctx, db, createTestAuthoredJobWithTasks())

	otherAuthoredJob := createTestAuthoredJobWithTasks()
	otherAuthoredJob.Status = api.JobStatusActive
	otherAuthoredJob.Tasks = []job_compilers.AuthoredTask{}
	otherAuthoredJob.JobID = "138678c8-efd0-452b-ac05-397ff4c02b26"
	otherAuthoredJob.Metadata["project"] = "Other Project"
	otherJob := persistAuthoredJob(t, ctx, db, otherAuthoredJob)

	var (
		result []*Job
		err    error
	)

	// Check empty result when querying for specific metadata:
	result, err = db.QueryJobs(ctx, api.JobsQuery{
		Metadata: &api.JobsQuery_Metadata{
			AdditionalProperties: map[string]string{
				"project": "Secret Future Project",
			}}})
	assert.NoError(t, err)
	assert.Len(t, result, 0)

	// Check job was returned properly when querying for the right project.
	result, err = db.QueryJobs(ctx, api.JobsQuery{
		Metadata: &api.JobsQuery_Metadata{
			AdditionalProperties: map[string]string{
				"project": testJob.Metadata["project"],
			}}})
	assert.NoError(t, err)
	if !assert.Len(t, result, 1) {
		t.FailNow()
	}
	assert.Equal(t, testJob.ID, result[0].ID)

	// Check for the other job
	result, err = db.QueryJobs(ctx, api.JobsQuery{
		Metadata: &api.JobsQuery_Metadata{
			AdditionalProperties: map[string]string{
				"project": otherJob.Metadata["project"],
			}}})
	assert.NoError(t, err)
	if !assert.Len(t, result, 1) {
		t.FailNow()
	}
	assert.Equal(t, otherJob.ID, result[0].ID)

	// Check job was returned properly when querying for empty metadata.
	result, err = db.QueryJobs(ctx, api.JobsQuery{
		OrderBy:  &[]string{"status"},
		Metadata: &api.JobsQuery_Metadata{AdditionalProperties: map[string]string{}},
	})
	assert.NoError(t, err)
	if !assert.Len(t, result, 2) {
		t.FailNow()
	}
	// 'active' should come before 'under-construction':
	assert.Equal(t, otherJob.ID, result[0].ID, "status is %s", result[0].Status)
	assert.Equal(t, testJob.ID, result[1].ID, "status is %s", result[1].Status)
}

func TestQueryJobTaskSummaries(t *testing.T) {
	ctx, close, db, job, authoredJob := jobTasksTestFixtures(t)
	defer close()

	expectTaskUUIDs := map[string]bool{}
	for _, task := range authoredJob.Tasks {
		expectTaskUUIDs[task.UUID] = true
	}

	// Create another test job, just to check we get the right tasks back.
	otherAuthoredJob := createTestAuthoredJobWithTasks()
	otherAuthoredJob.Status = api.JobStatusActive
	for i := range otherAuthoredJob.Tasks {
		otherAuthoredJob.Tasks[i].UUID = uuid.New()
		otherAuthoredJob.Tasks[i].Dependencies = []*job_compilers.AuthoredTask{}
	}
	otherAuthoredJob.JobID = "138678c8-efd0-452b-ac05-397ff4c02b26"
	otherAuthoredJob.Metadata["project"] = "Other Project"
	persistAuthoredJob(t, ctx, db, otherAuthoredJob)

	// Sanity check for the above code, there should be 6 tasks overall, 3 per job.
	var numTasks int64
	tx := db.gormDB.Model(&Task{}).Count(&numTasks)
	assert.NoError(t, tx.Error)
	assert.Equal(t, int64(6), numTasks)

	// Get the task summaries of a particular job.
	summaries, err := db.QueryJobTaskSummaries(ctx, job.UUID)
	assert.NoError(t, err)

	assert.Len(t, summaries, len(expectTaskUUIDs))
	for _, summary := range summaries {
		assert.True(t, expectTaskUUIDs[summary.UUID], "%q should be in %v", summary.UUID, expectTaskUUIDs)
	}
}
