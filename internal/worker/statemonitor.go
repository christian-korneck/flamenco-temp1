package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

// queryManagerForStateChange asks the Manager whether we should go to another state or not.
// Any error communicating with the Manager is logged but otherwise ignored.
// Returns nil when no state change is requested.
func (w *Worker) queryManagerForStateChange(ctx context.Context) *api.WorkerStatus {
	resp, err := w.client.WorkerStateWithResponse(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("error checking upstream state changes")
		return nil
	}
	switch {
	case resp.JSON200 != nil:
		log.Info().
			Str("requestedStatus", string(resp.JSON200.StatusRequested)).
			Msg("Manager requests status change")
		return &resp.JSON200.StatusRequested
	case resp.StatusCode() == http.StatusNoContent:
		log.Debug().Msg("we can stay in the current state")
	default:
		log.Warn().
			Int("code", resp.StatusCode()).
			Str("error", string(resp.Body)).
			Msg("unable to obtain requested state for unknown reason")
	}

	return nil
}

// mayIKeepRunning asks the Manager whether we can keep running a certain task.
// Any error communicating with the Manager is logged but otherwise ignored.
func (w *Worker) mayIKeepRunning(ctx context.Context, taskID string) api.MayKeepRunning {
	resp, err := w.client.MayWorkerRunWithResponse(ctx, taskID)
	if err != nil {
		log.Warn().
			Err(err).
			Str("task", taskID).
			Msg("error asking Manager may-I-keep-running task")
		return api.MayKeepRunning{MayKeepRunning: true}
	}

	switch {
	case resp.JSON200 != nil:
		mkr := *resp.JSON200
		logCtx := log.With().
			Str("task", taskID).
			Bool("mayKeepRunning", mkr.MayKeepRunning).
			Bool("statusChangeRequested", mkr.StatusChangeRequested)
		if mkr.Reason != "" {
			logCtx = logCtx.Str("reason", mkr.Reason)
		}
		logger := logCtx.Logger()
		logger.Debug().Msg("may-i-keep-running response")
		return mkr
	default:
		log.Warn().
			Str("task", taskID).
			Int("code", resp.StatusCode()).
			Str("error", string(resp.Body)).
			Msg("unable to check may-i-keep-running for unknown reason")
		return api.MayKeepRunning{MayKeepRunning: true}
	}
}
