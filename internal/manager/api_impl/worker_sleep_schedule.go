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

func (f *Flamenco) FetchWorkerSleepSchedule(e echo.Context, workerUUID string) error {
	if !uuid.IsValid(workerUUID) {
		return sendAPIError(e, http.StatusBadRequest, "not a valid UUID")
	}

	ctx := e.Request().Context()
	logger := requestLogger(e)
	logger = logger.With().Str("worker", workerUUID).Logger()
	schedule, err := f.sleepScheduler.FetchSchedule(ctx, workerUUID)

	switch {
	case errors.Is(err, persistence.ErrWorkerNotFound):
		logger.Warn().Msg("FetchWorkerSleepSchedule: worker does not exist")
		return sendAPIError(e, http.StatusNotFound, "worker %q does not exist", workerUUID)
	case err != nil:
		logger.Error().Err(err).Msg("FetchWorkerSleepSchedule: error fetching sleep schedule")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching sleep schedule: %v", err)
	case schedule == nil:
		return e.NoContent(http.StatusNoContent)
	}

	apiSchedule := api.WorkerSleepSchedule{
		DaysOfWeek: schedule.DaysOfWeek,
		EndTime:    schedule.EndTime.String(),
		IsActive:   schedule.IsActive,
		StartTime:  schedule.StartTime.String(),
	}
	return e.JSON(http.StatusOK, apiSchedule)
}

func (f *Flamenco) SetWorkerSleepSchedule(e echo.Context, workerUUID string) error {
	if !uuid.IsValid(workerUUID) {
		return sendAPIError(e, http.StatusBadRequest, "not a valid UUID")
	}

	ctx := e.Request().Context()
	logger := requestLogger(e)
	logger = logger.With().Str("worker", workerUUID).Logger()

	var req api.SetWorkerSleepScheduleJSONRequestBody
	err := e.Bind(&req)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}
	schedule := api.WorkerSleepSchedule(req)

	// Create a sleep schedule that can be persisted.
	dbSchedule := persistence.SleepSchedule{
		IsActive:   schedule.IsActive,
		DaysOfWeek: schedule.DaysOfWeek,
	}
	if err := dbSchedule.StartTime.Scan(schedule.StartTime); err != nil {
		logger.Warn().Err(err).Msg("bad request received, cannot parse schedule start time")
		return sendAPIError(e, http.StatusBadRequest, "invalid format for schedule start time")
	}
	if err := dbSchedule.EndTime.Scan(schedule.EndTime); err != nil {
		logger.Warn().Err(err).Msg("bad request received, cannot parse schedule end time")
		return sendAPIError(e, http.StatusBadRequest, "invalid format for schedule end time")
	}

	// Send the sleep schedule to the scheduler.
	err = f.sleepScheduler.SetSchedule(ctx, workerUUID, &dbSchedule)
	switch {
	case errors.Is(err, persistence.ErrWorkerNotFound):
		logger.Warn().Msg("SetWorkerSleepSchedule: worker does not exist")
		return sendAPIError(e, http.StatusNotFound, "worker %q does not exist", workerUUID)
	case err != nil:
		logger.Error().Err(err).Msg("SetWorkerSleepSchedule: error fetching sleep schedule")
		return sendAPIError(e, http.StatusInternalServerError, "error fetching sleep schedule: %v", err)
	}

	logger.Info().Interface("schedule", schedule).Msg("worker sleep schedule updated")
	return e.NoContent(http.StatusNoContent)
}
