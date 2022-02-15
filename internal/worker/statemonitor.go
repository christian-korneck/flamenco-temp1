package worker

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
	"net/http"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
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
