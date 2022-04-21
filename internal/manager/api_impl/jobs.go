package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/pkg/api"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func (f *Flamenco) GetJobTypes(e echo.Context) error {
	logger := requestLogger(e)

	if f.jobCompiler == nil {
		logger.Error().Msg("Flamenco is running without job compiler")
		return sendAPIError(e, http.StatusInternalServerError, "no job types available")
	}

	logger.Debug().Msg("listing job types")
	jobTypes := f.jobCompiler.ListJobTypes()
	return e.JSON(http.StatusOK, &jobTypes)
}

func (f *Flamenco) GetJobType(e echo.Context, typeName string) error {
	logger := requestLogger(e)

	if f.jobCompiler == nil {
		logger.Error().Msg("Flamenco is running without job compiler")
		return sendAPIError(e, http.StatusInternalServerError, "no job types available")
	}

	logger.Debug().Str("typeName", typeName).Msg("getting job type")
	jobType, err := f.jobCompiler.GetJobType(typeName)
	if err != nil {
		if err == job_compilers.ErrJobTypeUnknown {
			return sendAPIError(e, http.StatusNotFound, "no such job type known")
		}
		return sendAPIError(e, http.StatusInternalServerError, "error getting job type")
	}

	return e.JSON(http.StatusOK, jobType)
}

func (f *Flamenco) SubmitJob(e echo.Context) error {
	logger := requestLogger(e)

	var job api.SubmitJobJSONRequestBody
	if err := e.Bind(&job); err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	logger = logger.With().
		Str("type", job.Type).
		Str("name", job.Name).
		Logger()
	logger.Info().Msg("new Flamenco job received")

	ctx := e.Request().Context()
	submittedJob := api.SubmittedJob(job)
	authoredJob, err := f.jobCompiler.Compile(ctx, submittedJob)
	if err != nil {
		logger.Warn().Err(err).Msg("error compiling job")
		// TODO: make this a more specific error object for this API call.
		return sendAPIError(e, http.StatusBadRequest, fmt.Sprintf("error compiling job: %v", err))
	}

	logger = logger.With().Str("job_id", authoredJob.JobID).Logger()

	// TODO: check whether this job should be queued immediately or start paused.
	authoredJob.Status = api.JobStatusQueued

	if err := f.persist.StoreAuthoredJob(ctx, *authoredJob); err != nil {
		logger.Error().Err(err).Msg("error persisting job in database")
		return sendAPIError(e, http.StatusInternalServerError, "error persisting job in database")
	}

	dbJob, err := f.persist.FetchJob(ctx, authoredJob.JobID)
	if err != nil {
		logger.Error().Err(err).Msg("unable to retrieve just-stored job from database")
		return sendAPIError(e, http.StatusInternalServerError, "error retrieving job from database")
	}

	jobUpdate := webupdates.NewJobUpdate(dbJob)
	f.broadcaster.BroadcastNewJob(jobUpdate)

	apiJob := jobDBtoAPI(dbJob)
	return e.JSON(http.StatusOK, apiJob)
}

func (f *Flamenco) FetchJob(e echo.Context, jobId string) error {
	logger := requestLogger(e).With().
		Str("job", jobId).
		Logger()

	if _, err := uuid.Parse(jobId); err != nil {
		logger.Debug().Msg("invalid job ID received")
		return sendAPIError(e, http.StatusBadRequest, "job ID not valid")
	}

	logger.Debug().Msg("fetching job")

	ctx := e.Request().Context()
	dbJob, err := f.persist.FetchJob(ctx, jobId)
	if err != nil {
		logger.Warn().Err(err).Msg("cannot fetch job")
		return sendAPIError(e, http.StatusNotFound, fmt.Sprintf("job %+v not found", jobId))
	}

	apiJob := jobDBtoAPI(dbJob)
	return e.JSON(http.StatusOK, apiJob)
}

func (f *Flamenco) SetJobStatus(e echo.Context, jobID string) error {
	logger := requestLogger(e)
	ctx := e.Request().Context()

	logger = logger.With().Str("job", jobID).Logger()

	var statusChange api.SetJobStatusJSONRequestBody
	if err := e.Bind(&statusChange); err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	dbJob, err := f.persist.FetchJob(ctx, jobID)
	if err != nil {
		if errors.Is(err, persistence.ErrJobNotFound) {
			return sendAPIError(e, http.StatusNotFound, "no such job")
		}
		logger.Error().Err(err).Msg("error fetching job")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching job")
	}

	logger = logger.With().
		Str("currentstatus", string(dbJob.Status)).
		Str("requestedStatus", string(statusChange.Status)).
		Str("reason", statusChange.Reason).
		Logger()
	logger.Info().Msg("job status change requested")

	err = f.stateMachine.JobStatusChange(ctx, dbJob, statusChange.Status, statusChange.Reason)
	if err != nil {
		logger.Error().Err(err).Msg("error changing job status")
		return sendAPIError(e, http.StatusInternalServerError, "unexpected error changing job status")
	}
	return e.NoContent(http.StatusNoContent)
}

func (f *Flamenco) TaskUpdate(e echo.Context, taskID string) error {
	logger := requestLogger(e)
	worker := requestWorkerOrPanic(e)

	if _, err := uuid.Parse(taskID); err != nil {
		logger.Debug().Msg("invalid task ID received")
		return sendAPIError(e, http.StatusBadRequest, "task ID not valid")
	}
	logger = logger.With().Str("taskID", taskID).Logger()

	// Fetch the task, to see if this worker is even allowed to send us updates.
	ctx := e.Request().Context()
	dbTask, err := f.persist.FetchTask(ctx, taskID)
	if err != nil {
		logger.Warn().Err(err).Msg("cannot fetch task")
		if errors.Is(err, persistence.ErrTaskNotFound) {
			return sendAPIError(e, http.StatusNotFound, "task %+v not found", taskID)
		}
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task")
	}
	if dbTask == nil {
		panic("task could not be fetched, but database gave no error either")
	}

	// Decode the request body.
	var taskUpdate api.TaskUpdateJSONRequestBody
	if err := e.Bind(&taskUpdate); err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}
	if dbTask.WorkerID == nil {
		logger.Warn().
			Msg("worker trying to update task that's not assigned to any worker")
		return sendAPIError(e, http.StatusConflict, "task %+v is not assigned to any worker, so also not to you", taskID)
	}
	if *dbTask.WorkerID != worker.ID {
		logger.Warn().Msg("worker trying to update task that's assigned to another worker")
		return sendAPIError(e, http.StatusConflict, "task %+v is not assigned to you", taskID)
	}

	// TODO: check whether this task may undergo the requested status change.

	if err := f.doTaskUpdate(ctx, logger, worker, dbTask, taskUpdate); err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "unable to handle status update: %v", err)
	}

	return e.NoContent(http.StatusNoContent)
}

func (f *Flamenco) doTaskUpdate(
	ctx context.Context,
	logger zerolog.Logger,
	w *persistence.Worker,
	dbTask *persistence.Task,
	update api.TaskUpdateJSONRequestBody,
) error {
	if dbTask.Job == nil {
		logger.Panic().Msg("dbTask.Job is nil, unable to continue")
	}

	var dbErr error

	if update.TaskStatus != nil {
		oldTaskStatus := dbTask.Status
		err := f.stateMachine.TaskStatusChange(ctx, dbTask, *update.TaskStatus)
		if err != nil {
			logger.Error().Err(err).
				Str("newTaskStatus", string(*update.TaskStatus)).
				Str("oldTaskStatus", string(oldTaskStatus)).
				Msg("error changing task status")
			dbErr = fmt.Errorf("changing status of task %s to %q: %w",
				dbTask.UUID, *update.TaskStatus, err)
		}
	}

	if update.Activity != nil {
		dbTask.Activity = *update.Activity
		dbErr = f.persist.SaveTaskActivity(ctx, dbTask)
	}

	if update.Log != nil {
		// Errors writing the log to file should be logged in our own logging
		// system, but shouldn't abort the render. As such, `err` is not returned to
		// the caller.
		err := f.logStorage.Write(logger, dbTask.Job.UUID, dbTask.UUID, *update.Log)
		if err != nil {
			logger.Error().Err(err).Msg("error writing task log")
		}
	}

	return dbErr
}

func jobDBtoAPI(dbJob *persistence.Job) api.Job {
	apiJob := api.Job{
		SubmittedJob: api.SubmittedJob{
			Name:     dbJob.Name,
			Priority: dbJob.Priority,
			Type:     dbJob.JobType,
		},

		Id:       dbJob.UUID,
		Created:  dbJob.CreatedAt,
		Updated:  dbJob.UpdatedAt,
		Status:   api.JobStatus(dbJob.Status),
		Activity: dbJob.Activity,
	}

	apiJob.Settings = &api.JobSettings{AdditionalProperties: dbJob.Settings}
	apiJob.Metadata = &api.JobMetadata{AdditionalProperties: dbJob.Metadata}

	return apiJob
}
