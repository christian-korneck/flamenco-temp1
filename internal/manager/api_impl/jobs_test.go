package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](value T) *T {
	return &value
}

func TestSubmitJob(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	submittedJob := api.SubmittedJob{
		Name:     "поднео посао",
		Type:     "test",
		Priority: 50,
	}

	// Expect the job compiler to be called.
	authoredJob := job_compilers.AuthoredJob{
		JobID:    "afc47568-bd9d-4368-8016-e91d945db36d",
		Name:     submittedJob.Name,
		JobType:  submittedJob.Type,
		Priority: submittedJob.Priority,
		Status:   api.JobStatusUnderConstruction,
		Created:  time.Now(),
	}
	mf.jobCompiler.EXPECT().Compile(gomock.Any(), submittedJob).Return(&authoredJob, nil)

	// Expect the job to be saved with 'queued' status:
	queuedJob := authoredJob
	queuedJob.Status = api.JobStatusQueued
	mf.persistence.EXPECT().StoreAuthoredJob(gomock.Any(), queuedJob).Return(nil)

	// Expect the job to be fetched from the database again:
	dbJob := persistence.Job{
		UUID:     queuedJob.JobID,
		Name:     queuedJob.Name,
		JobType:  queuedJob.JobType,
		Priority: queuedJob.Priority,
		Status:   queuedJob.Status,
		Settings: persistence.StringInterfaceMap{},
		Metadata: persistence.StringStringMap{},
	}
	mf.persistence.EXPECT().FetchJob(gomock.Any(), queuedJob.JobID).Return(&dbJob, nil)

	// Expect the new job to be broadcast.
	jobUpdate := api.SocketIOJobUpdate{
		Id:       dbJob.UUID,
		Name:     &dbJob.Name,
		Priority: dbJob.Priority,
		Status:   dbJob.Status,
		Type:     dbJob.JobType,
		Updated:  dbJob.UpdatedAt,
	}
	mf.broadcaster.EXPECT().BroadcastNewJob(jobUpdate)

	// Do the call.
	echoCtx := mf.prepareMockedJSONRequest(submittedJob)
	requestWorkerStore(echoCtx, &worker)
	err := mf.flamenco.SubmitJob(echoCtx)
	assert.NoError(t, err)

}

func TestGetJobTypeHappy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mf := newMockedFlamenco(mockCtrl)

	// Get an existing job type.
	jt := api.AvailableJobType{
		Name:  "test-job-type",
		Label: "Test Job Type",
		Settings: []api.AvailableJobSetting{
			{Key: "setting", Type: api.AvailableJobSettingTypeString},
		},
	}
	mf.jobCompiler.EXPECT().GetJobType("test-job-type").
		Return(jt, nil)

	echoCtx := mf.prepareMockedRequest(nil)
	err := mf.flamenco.GetJobType(echoCtx, "test-job-type")
	assert.NoError(t, err)

	assertResponseJSON(t, echoCtx, http.StatusOK, jt)
}

func TestGetJobTypeUnknown(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mf := newMockedFlamenco(mockCtrl)

	// Get a non-existing job type.
	mf.jobCompiler.EXPECT().GetJobType("nonexistent-type").
		Return(api.AvailableJobType{}, job_compilers.ErrJobTypeUnknown)

	echoCtx := mf.prepareMockedRequest(nil)
	err := mf.flamenco.GetJobType(echoCtx, "nonexistent-type")
	assert.NoError(t, err)
	assertResponseJSON(t, echoCtx, http.StatusNotFound, api.Error{
		Code:    http.StatusNotFound,
		Message: "no such job type known",
	})
}

func TestGetJobTypeError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mf := newMockedFlamenco(mockCtrl)

	// Get an error situation.
	mf.jobCompiler.EXPECT().GetJobType("error").
		Return(api.AvailableJobType{}, errors.New("didn't expect this"))
	echoCtx := mf.prepareMockedRequest(nil)
	err := mf.flamenco.GetJobType(echoCtx, "error")
	assert.NoError(t, err)
	assertResponseAPIError(t, echoCtx, http.StatusInternalServerError, "error getting job type")
}

func TestSetJobStatus_nonexistentJob(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)

	jobID := "18a9b096-d77e-438c-9be2-74397038298b"
	statusUpdate := api.JobStatusChange{
		Status: api.JobStatusCancelRequested,
		Reason: "someone pushed a button",
	}

	mf.persistence.EXPECT().FetchJob(gomock.Any(), jobID).Return(nil, persistence.ErrJobNotFound)

	// Do the call.
	echoCtx := mf.prepareMockedJSONRequest(statusUpdate)
	err := mf.flamenco.SetJobStatus(echoCtx, jobID)
	assert.NoError(t, err)

	assertResponseAPIError(t, echoCtx, http.StatusNotFound, "no such job")
}

func TestSetJobStatus_happy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)

	jobID := "18a9b096-d77e-438c-9be2-74397038298b"
	statusUpdate := api.JobStatusChange{
		Status: api.JobStatusCancelRequested,
		Reason: "someone pushed a button",
	}
	dbJob := persistence.Job{
		UUID:     jobID,
		Name:     "test job",
		Status:   api.JobStatusActive,
		Settings: persistence.StringInterfaceMap{},
		Metadata: persistence.StringStringMap{},
	}

	// Set up expectations.
	ctx := gomock.Any()
	mf.persistence.EXPECT().FetchJob(ctx, jobID).Return(&dbJob, nil)
	mf.stateMachine.EXPECT().JobStatusChange(ctx, &dbJob, statusUpdate.Status, "someone pushed a button")

	// Do the call.
	echoCtx := mf.prepareMockedJSONRequest(statusUpdate)
	err := mf.flamenco.SetJobStatus(echoCtx, jobID)
	assert.NoError(t, err)

	assertResponseEmpty(t, echoCtx)
}
