// SPDX-License-Identifier: GPL-3.0-or-later
package api_impl

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
)

func (f *Flamenco) FetchJob(e echo.Context, jobID string) error {
	logger := requestLogger(e).With().
		Str("job", jobID).
		Logger()

	if !uuid.IsValid(jobID) {
		logger.Debug().Msg("invalid job ID received")
		return sendAPIError(e, http.StatusBadRequest, "job ID not valid")
	}

	logger.Debug().Msg("fetching job")

	ctx := e.Request().Context()
	dbJob, err := f.persist.FetchJob(ctx, jobID)
	if err != nil {
		logger.Warn().Err(err).Msg("cannot fetch job")
		return sendAPIError(e, http.StatusNotFound, fmt.Sprintf("job %+v not found", jobID))
	}

	apiJob := jobDBtoAPI(dbJob)
	return e.JSON(http.StatusOK, apiJob)
}

func (f *Flamenco) QueryJobs(e echo.Context) error {
	logger := requestLogger(e)

	var jobsQuery api.QueryJobsJSONRequestBody
	if err := e.Bind(&jobsQuery); err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	ctx := e.Request().Context()
	dbJobs, err := f.persist.QueryJobs(ctx, api.JobsQuery(jobsQuery))
	if err != nil {
		logger.Warn().Err(err).Msg("error querying for jobs")
		return sendAPIError(e, http.StatusInternalServerError, "error querying for jobs")
	}

	apiJobs := make([]api.Job, len(dbJobs))
	for i, dbJob := range dbJobs {
		apiJobs[i] = jobDBtoAPI(dbJob)
	}
	result := api.JobsQueryResult{
		Jobs: apiJobs,
	}
	return e.JSON(http.StatusOK, result)
}

func (f *Flamenco) FetchJobTasks(e echo.Context, jobID string) error {
	logger := requestLogger(e).With().
		Str("job", jobID).
		Logger()
	ctx := e.Request().Context()

	if !uuid.IsValid(jobID) {
		logger.Debug().Msg("invalid job ID received")
		return sendAPIError(e, http.StatusBadRequest, "job ID not valid")
	}

	tasks, err := f.persist.QueryJobTaskSummaries(ctx, jobID)
	if err != nil {
		logger.Warn().Err(err).Msg("error querying for jobs")
		return sendAPIError(e, http.StatusInternalServerError, "error querying for jobs")
	}

	summaries := make([]api.TaskSummary, len(tasks))
	for i, task := range tasks {
		summaries[i] = taskDBtoSummary(task)
	}
	result := api.JobTasksSummary{
		Tasks: &summaries,
	}
	return e.JSON(http.StatusOK, result)
}

func (f *Flamenco) FetchTask(e echo.Context, taskID string) error {
	logger := requestLogger(e).With().
		Str("task", taskID).
		Logger()
	ctx := e.Request().Context()

	if !uuid.IsValid(taskID) {
		logger.Debug().Msg("invalid job ID received")
		return sendAPIError(e, http.StatusBadRequest, "job ID not valid")
	}

	// Fetch & convert the task.
	task, err := f.persist.FetchTask(ctx, taskID)
	if errors.Is(err, persistence.ErrTaskNotFound) {
		logger.Debug().Msg("non-existent task requested")
		return sendAPIError(e, http.StatusNotFound, "no such task")
	}
	if err != nil {
		logger.Warn().Err(err).Msg("error fetching task")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task")
	}
	apiTask := taskDBtoAPI(task)

	// Fetch & convert the failure list.
	failedWorkers, err := f.persist.FetchTaskFailureList(ctx, task)
	if err != nil {
		logger.Warn().Err(err).Msg("error fetching task failure list")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching task failure list")
	}
	failedTaskWorkers := make([]api.TaskWorker, len(failedWorkers))
	for idx, worker := range failedWorkers {
		failedTaskWorkers[idx] = *workerToTaskWorker(worker)
	}
	apiTask.FailedByWorkers = &failedTaskWorkers

	return e.JSON(http.StatusOK, apiTask)
}

func taskDBtoSummary(task *persistence.Task) api.TaskSummary {
	return api.TaskSummary{
		Id:       task.UUID,
		Name:     task.Name,
		Priority: task.Priority,
		Status:   task.Status,
		TaskType: task.Type,
		Updated:  task.UpdatedAt,
	}
}
