// SPDX-License-Identifier: GPL-3.0-or-later
package api_impl

import (
	"net/http"

	"git.blender.org/flamenco/pkg/api"
	"github.com/labstack/echo/v4"
)

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
		sendAPIError(e, http.StatusInternalServerError, "error querying for jobs")
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
