package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"net/http"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
	"github.com/labstack/echo/v4"
)

func (f *Flamenco) FetchWorkers(e echo.Context) error {
	logger := requestLogger(e)
	dbWorkers, err := f.persist.FetchWorkers(e.Request().Context())
	if err != nil {
		logger.Error().Err(err).Msg("error fetching all workers")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching workers: %v", err)
	}

	apiWorkers := make([]api.WorkerSummary, len(dbWorkers))
	for i := range dbWorkers {
		apiWorkers[i] = workerSummary(*dbWorkers[i])
	}

	logger.Debug().Msg("fetched all workers")
	return e.JSON(http.StatusOK, api.WorkerList{
		Workers: apiWorkers,
	})
}

func (f *Flamenco) FetchWorker(e echo.Context, workerUUID string) error {
	logger := requestLogger(e)
	logger = logger.With().Str("worker", workerUUID).Logger()

	if !uuid.IsValid(workerUUID) {
		return sendAPIError(e, http.StatusBadRequest, "not a valid UUID")
	}

	dbWorker, err := f.persist.FetchWorker(e.Request().Context(), workerUUID)
	if errors.Is(err, persistence.ErrWorkerNotFound) {
		logger.Debug().Msg("non-existent worker requested")
		return sendAPIError(e, http.StatusNotFound, "worker %q not found", workerUUID)
	}
	if err != nil {
		logger.Error().Err(err).Msg("error fetching worker")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching worker: %v", err)
	}

	logger.Debug().Msg("fetched worker")
	apiWorker := workerDBtoAPI(dbWorker)
	return e.JSON(http.StatusOK, apiWorker)
}

func workerSummary(w persistence.Worker) api.WorkerSummary {
	summary := api.WorkerSummary{
		Id:       w.UUID,
		Nickname: w.Name,
		Status:   w.Status,
		Version:  w.Software,
	}
	if w.StatusRequested != "" {
		summary.StatusRequested = &w.StatusRequested
	}
	return summary
}

func workerDBtoAPI(dbWorker *persistence.Worker) api.Worker {
	apiWorker := api.Worker{
		WorkerSummary: api.WorkerSummary{
			Id:       dbWorker.UUID,
			Nickname: dbWorker.Name,
			Status:   dbWorker.Status,
			Version:  dbWorker.Software,
		},
		IpAddress:          dbWorker.Address,
		Platform:           dbWorker.Platform,
		SupportedTaskTypes: dbWorker.TaskTypes(),
	}

	if dbWorker.StatusRequested != "" {
		apiWorker.StatusRequested = &dbWorker.StatusRequested
	}

	return apiWorker
}
