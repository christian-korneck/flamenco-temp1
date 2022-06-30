package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
)

// JobFilesURLPrefix is the URL prefix that the Flamenco API expects to serve
// the job-specific local files, i.e. the ones that are managed by
// `local_storage.StorageInfo`.
const JobFilesURLPrefix = "/job-files"

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

	// Only in this function, i.e. only when changing the job from the web
	// interface, does requeueing the job mean it should clear the failure list.
	// This is why this is implemented here, and not in the Task State Machine.
	switch statusChange.Status {
	case api.JobStatusRequeueing:
		if err := f.persist.ClearFailureListOfJob(ctx, dbJob); err != nil {
			logger.Error().Err(err).Msg("error clearing failure list")
			return sendAPIError(e, http.StatusInternalServerError, "unexpected error clearing the job's tasks' failure list")
		}
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

	// Only in this function, i.e. only when changing the task from the web
	// interface, does requeueing the task mean it should clear the failure list.
	// This is why this is implemented here, and not in the Task State Machine.
	switch statusChange.Status {
	case api.TaskStatusQueued:
		if err := f.persist.ClearFailureListOfTask(ctx, dbTask); err != nil {
			logger.Error().Err(err).Msg("error clearing failure list")
			return sendAPIError(e, http.StatusInternalServerError, "unexpected error clearing the task's failure list")
		}
	}

	return e.NoContent(http.StatusNoContent)
}

func (f *Flamenco) FetchTaskLogTail(e echo.Context, taskID string) error {
	logger := requestLogger(e)
	ctx := e.Request().Context()

	logger = logger.With().Str("task", taskID).Logger()
	if !uuid.IsValid(taskID) {
		logger.Warn().Msg("fetchTaskLogTail: bad task ID ")
		return sendAPIError(e, http.StatusBadRequest, "bad task ID")
	}

	dbTask, err := f.persist.FetchTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, persistence.ErrTaskNotFound) {
			return sendAPIError(e, http.StatusNotFound, "no such task")
		}
		logger.Error().Err(err).Msg("error fetching task")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task: %v", err)
	}
	logger = logger.With().Str("job", dbTask.Job.UUID).Logger()

	tail, err := f.logStorage.Tail(dbTask.Job.UUID, taskID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debug().Msg("task tail unavailable, task has no log on disk")
			return e.NoContent(http.StatusNoContent)
		}
		logger.Error().Err(err).Msg("unable to fetch task log tail")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task log tail: %v", err)
	}

	if tail == "" {
		logger.Debug().Msg("task tail unavailable, on-disk task log is empty")
		return e.NoContent(http.StatusNoContent)
	}

	logger.Debug().Msg("fetched task tail")
	return e.String(http.StatusOK, tail)
}

func (f *Flamenco) FetchJobBlocklist(e echo.Context, jobID string) error {
	if !uuid.IsValid(jobID) {
		return sendAPIError(e, http.StatusBadRequest, "job ID should be a UUID")
	}

	logger := requestLogger(e).With().Str("job", jobID).Logger()
	ctx := e.Request().Context()

	list, err := f.persist.FetchJobBlocklist(ctx, jobID)
	if err != nil {
		logger.Error().Err(err).Msg("error fetching job blocklist")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching job blocklist: %v", err)
	}

	apiList := api.JobBlocklist{}
	for _, item := range list {
		apiList = append(apiList, api.JobBlocklistEntry{
			TaskType: item.TaskType,
			WorkerId: item.Worker.UUID,
		})
	}

	return e.JSON(http.StatusOK, apiList)
}

func (f *Flamenco) RemoveJobBlocklist(e echo.Context, jobID string) error {
	if !uuid.IsValid(jobID) {
		return sendAPIError(e, http.StatusBadRequest, "job ID should be a UUID")
	}

	logger := requestLogger(e).With().Str("job", jobID).Logger()
	ctx := e.Request().Context()

	var job api.RemoveJobBlocklistJSONRequestBody
	if err := e.Bind(&job); err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	var lastErr error
	for _, entry := range job {
		sublogger := logger.With().
			Str("worker", entry.WorkerId).
			Str("taskType", entry.TaskType).
			Logger()
		err := f.persist.RemoveFromJobBlocklist(ctx, jobID, entry.WorkerId, entry.TaskType)
		if err != nil {
			sublogger.Error().Err(err).Msg("error removing entry from job blocklist")
			lastErr = err
			continue
		}
		sublogger.Info().Msg("removed entry from job blocklist")
	}

	if lastErr != nil {
		return sendAPIError(e, http.StatusInternalServerError,
			"error removing at least one entry from the blocklist: %v", lastErr)
	}

	return e.NoContent(http.StatusNoContent)
}

func (f *Flamenco) FetchJobLastRenderedInfo(e echo.Context, jobID string) error {
	if !uuid.IsValid(jobID) {
		return sendAPIError(e, http.StatusBadRequest, "job ID should be a UUID")
	}

	logger := requestLogger(e)
	info, err := f.lastRenderedInfoForJob(logger, jobID)
	if err != nil {
		logger.Error().
			Str("job", jobID).
			Err(err).
			Msg("error getting last-rendered info")
		return sendAPIError(e, http.StatusInternalServerError, "error finding last-rendered info: %v", err)
	}

	return e.JSON(http.StatusOK, info)
}

func (f *Flamenco) lastRenderedInfoForJob(logger zerolog.Logger, jobUUID string) (*api.JobLastRenderedImageInfo, error) {
	basePath := f.lastRender.PathForJob(jobUUID)
	relPath, err := f.localStorage.RelPath(basePath)
	if err != nil {
		return nil, fmt.Errorf(
			"last-rendered path for job %s is %q, which is outside local storage root: %w",
			jobUUID, basePath, err)
	}

	suffixes := []string{}
	for _, spec := range f.lastRender.ThumbSpecs() {
		suffixes = append(suffixes, spec.Filename)
	}

	info := api.JobLastRenderedImageInfo{
		Base:     path.Join(JobFilesURLPrefix, relPath),
		Suffixes: suffixes,
	}
	return &info, nil
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
		Worker:   workerToTaskWorker(dbTask.Worker),
	}

	if dbTask.Job != nil {
		apiTask.JobId = dbTask.Job.UUID
	}

	if !dbTask.LastTouchedAt.IsZero() {
		apiTask.LastTouched = &dbTask.LastTouchedAt
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

// workerToTaskWorker is nil-safe.
func workerToTaskWorker(worker *persistence.Worker) *api.TaskWorker {
	if worker == nil {
		return nil
	}
	return &api.TaskWorker{
		Id:      worker.UUID,
		Name:    worker.Name,
		Address: worker.Address,
	}
}
