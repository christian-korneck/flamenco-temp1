package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/pkg/api"
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

// SetJobStatus is used by the web interface to change a job's status.
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

// SetTaskStatus is used by the web interface to change a task's status.
func (f *Flamenco) SetTaskStatus(e echo.Context, taskID string) error {
	logger := requestLogger(e)
	ctx := e.Request().Context()

	logger = logger.With().Str("task", taskID).Logger()

	var statusChange api.SetTaskStatusJSONRequestBody
	if err := e.Bind(&statusChange); err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	dbTask, err := f.persist.FetchTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, persistence.ErrTaskNotFound) {
			return sendAPIError(e, http.StatusNotFound, "no such task")
		}
		logger.Error().Err(err).Msg("error fetching task")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task")
	}

	logger = logger.With().
		Str("currentstatus", string(dbTask.Status)).
		Str("requestedStatus", string(statusChange.Status)).
		Str("reason", statusChange.Reason).
		Logger()
	logger.Info().Msg("task status change requested")

	// Store the reason for the status change in the task's Activity.
	dbTask.Activity = statusChange.Reason
	err = f.persist.SaveTaskActivity(ctx, dbTask)
	if err != nil {
		logger.Error().Err(err).Msg("error saving reason of task status change to its activity field")
		return sendAPIError(e, http.StatusInternalServerError, "unexpected error changing task status")
	}

	// Perform the actual status change.
	err = f.stateMachine.TaskStatusChange(ctx, dbTask, statusChange.Status)
	if err != nil {
		logger.Error().Err(err).Msg("error changing task status")
		return sendAPIError(e, http.StatusInternalServerError, "unexpected error changing task status")
	}
	return e.NoContent(http.StatusNoContent)
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

func taskDBtoAPI(dbTask *persistence.Task) api.Task {
	apiTask := api.Task{
		Id:       dbTask.UUID,
		Name:     dbTask.Name,
		Priority: dbTask.Priority,
		TaskType: dbTask.Type,
		Created:  dbTask.CreatedAt,
		Updated:  dbTask.UpdatedAt,
		Status:   dbTask.Status,
		Activity: dbTask.Activity,
		Commands: make([]api.Command, len(dbTask.Commands)),
	}

	if dbTask.Job != nil {
		apiTask.JobId = dbTask.Job.UUID
	}

	if dbTask.Worker != nil {
		apiTask.Worker = &api.TaskWorker{
			Id:      dbTask.Worker.UUID,
			Name:    dbTask.Worker.Name,
			Address: dbTask.Worker.Address,
		}
	}

	for i := range dbTask.Commands {
		apiTask.Commands[i] = commandDBtoAPI(dbTask.Commands[i])
	}

	return apiTask
}

func commandDBtoAPI(dbCommand persistence.Command) api.Command {
	return api.Command{
		Name:       dbCommand.Name,
		Parameters: dbCommand.Parameters,
	}
}
