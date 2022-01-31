package worker

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

const durationSleepCheck = 3 * time.Second

func (w *Worker) gotoStateAsleep(ctx context.Context) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	w.state = api.WorkerStatusAsleep
	w.doneWg.Add(2)
	go w.ackStateChange(ctx, w.state)
	go w.runStateAsleep(ctx)
}

func (w *Worker) runStateAsleep(ctx context.Context) {
	defer w.doneWg.Done()
	logger := log.With().Str("status", string(w.state)).Logger()
	logger.Info().Msg("sleeping")

	for {
		select {
		case <-ctx.Done():
			logger.Debug().Msg("state fetching interrupted by context cancellation")
			return
		case <-w.doneChan:
			logger.Debug().Msg("state fetching interrupted by shutdown")
			return
		case <-time.After(durationSleepCheck):
		}
		if !w.isState(api.WorkerStatusAwake) {
			logger.Debug().Msg("state fetching interrupted by state change")
			return
		}

		resp, err := w.client.WorkerStateWithResponse(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error checking upstream state changes")
		}
		switch {
		case resp.JSON200 != nil:
			log.Info().
				Str("requestedStatus", string(resp.JSON200.StatusRequested)).
				Msg("Manager requests status change")
			w.changeState(ctx, resp.JSON200.StatusRequested)
			return
		case resp.StatusCode() == http.StatusNoContent:
			log.Debug().Msg("we can keep sleeping")
			continue
		default:
			log.Warn().
				Int("code", resp.StatusCode()).
				Str("error", string(resp.Body)).
				Msg("unable to obtain requested state for unknown reason")
			continue
		}
	}
}
