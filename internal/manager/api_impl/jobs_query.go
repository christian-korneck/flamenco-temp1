// SPDX-License-Identifier: GPL-3.0-or-later
package api_impl

import (
	"fmt"
	"net/http"

	"git.blender.org/flamenco/pkg/api"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (f *Flamenco) FetchJob(e echo.Context, jobID string) error {
	logger := requestLogger(e).With().
		Str("job", jobID).
		Logger()

	if _, err := uuid.Parse(jobID); err != nil {
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
