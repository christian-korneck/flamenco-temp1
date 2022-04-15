package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
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
	jobUpdate := api.JobUpdate{
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

func TestTaskUpdate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)
	worker := testWorker()

	// Construct the JSON request object.
	taskUpdate := api.TaskUpdateJSONRequestBody{
		Activity:   ptr("testing"),
		Log:        ptr("line1\nline2\n"),
		TaskStatus: ptr(api.TaskStatusFailed),
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
		Activity: "pre-update activity",
	}

	// Expect the task to be fetched.
	mf.persistence.EXPECT().FetchTask(gomock.Any(), taskID).Return(&mockTask, nil)

	// Expect the task status change to be handed to the state machine.
	var statusChangedtask persistence.Task
	mf.stateMachine.EXPECT().TaskStatusChange(gomock.Any(), gomock.AssignableToTypeOf(&persistence.Task{}), api.TaskStatusFailed).
		DoAndReturn(func(ctx context.Context, task *persistence.Task, newStatus api.TaskStatus) error {
			statusChangedtask = *task
			return nil
		})

	// Expect the activity to be updated.
	var actUpdatedTask persistence.Task
	mf.persistence.EXPECT().SaveTaskActivity(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, task *persistence.Task) error {
			actUpdatedTask = *task
			return nil
		})

	// Expect the log to be written.
	mf.logStorage.EXPECT().Write(gomock.Any(), jobID, taskID, "line1\nline2\n")

	// Do the call.
	echoCtx := mf.prepareMockedJSONRequest(taskUpdate)
	requestWorkerStore(echoCtx, &worker)
	err := mf.flamenco.TaskUpdate(echoCtx, taskID)

	// Check the saved task.
	assert.NoError(t, err)
	assert.Equal(t, mockTask.UUID, statusChangedtask.UUID)
	assert.Equal(t, mockTask.UUID, actUpdatedTask.UUID)
	assert.Equal(t, "pre-update activity", statusChangedtask.Activity) // the 'save' should come from the change in status.
	assert.Equal(t, "testing", actUpdatedTask.Activity)                // the activity should be saved separately.
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

	assertJSONResponse(t, echoCtx, http.StatusOK, jt)
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
	assertJSONResponse(t, echoCtx, http.StatusNotFound, api.Error{
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
	assertAPIErrorResponse(t, echoCtx, http.StatusInternalServerError, "error getting job type")
}
