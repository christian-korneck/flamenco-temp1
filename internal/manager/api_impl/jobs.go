package api_impl

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
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

func (f *Flamenco) GetJobTypes(e echo.Context) error {
	if f.jobCompiler == nil {
		log.Error().Msg("Flamenco is running without job compiler")
		return sendAPIError(e, http.StatusInternalServerError, "no job types available")
	}

	jobTypes := f.jobCompiler.ListJobTypes()
	return e.JSON(http.StatusOK, &jobTypes)
}

func (f *Flamenco) SubmitJob(e echo.Context) error {
	// TODO: move this into some middleware.
	logger := log.With().
		Str("ip", e.RealIP()).
		Logger()

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
	return e.JSON(http.StatusOK, dbJob)
}

func (f *Flamenco) FetchJob(e echo.Context, jobId string) error {
	// TODO: move this into some middleware.
	logger := requestLogger(e).With().
		Str("job_id", jobId).
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

	apiJob := api.Job{
		SubmittedJob: api.SubmittedJob{
			Name:     dbJob.Name,
			Priority: dbJob.Priority,
			Type:     dbJob.JobType,
		},

		Id:      dbJob.UUID,
		Created: dbJob.CreatedAt,
		Updated: dbJob.UpdatedAt,
		Status:  api.JobStatus(dbJob.Status),
	}

	apiJob.Settings = &api.JobSettings{AdditionalProperties: dbJob.Settings}
	apiJob.Metadata = &api.JobMetadata{AdditionalProperties: dbJob.Metadata}

	return e.JSON(http.StatusOK, apiJob)
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
		return sendAPIError(e, http.StatusNotFound, "task %+v not found", taskID)
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

	return e.String(http.StatusNoContent, "")
}

func (f *Flamenco) doTaskUpdate(
	ctx context.Context,
	logger zerolog.Logger,
	w *persistence.Worker,
	dbTask *persistence.Task,
	update api.TaskUpdateJSONRequestBody,
) error {
	if update.TaskStatus != nil {
		// TODO: check that this status transition is valid.
		// TODO: process this status transition.
		newStatus := string(*update.TaskStatus)
		logger.Info().
			Str("oldStatus", dbTask.Status).
			Str("newStatus", newStatus).
			Msg("task changing status")
		dbTask.Status = newStatus
	}

	if update.Activity != nil {
		dbTask.Activity = *update.Activity
	}

	if update.Log != nil {
		// TODO: write log to disk.
		logger.Warn().Msg("task logs are not yet handled")
	}

	return f.persist.SaveTask(ctx, dbTask)
}
