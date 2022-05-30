package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"net/http"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
	"github.com/labstack/echo/v4"
)

func (f *Flamenco) FetchWorkers(e echo.Context) error {
	dbWorkers, err := f.persist.FetchWorkers(e.Request().Context())
	if err != nil {
		return sendAPIError(e, http.StatusInternalServerError, "error fetching workers: %v", err)
	}

	apiWorkers := make([]api.WorkerSummary, len(dbWorkers))
	for i := range dbWorkers {
		apiWorkers[i] = workerSummary(*dbWorkers[i])
	}

	return e.JSON(http.StatusOK, api.WorkerList{
		Workers: apiWorkers,
	})
}

func workerSummary(w persistence.Worker) api.WorkerSummary {
	summary := api.WorkerSummary{
		Id:       w.UUID,
		Nickname: w.Name,
		Status:   w.Status,
	}
	if w.StatusRequested != "" {
		summary.StatusRequested = &w.StatusRequested
	}
	return summary
}
