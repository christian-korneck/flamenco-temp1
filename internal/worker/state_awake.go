package worker

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

const (
	// How long to wait to fetch another task...
	durationNoTask      = 5 * time.Second  // ... if there is no task now.
	durationFetchFailed = 10 * time.Second // ... if fetching failed somehow.
)

var (
	errUnknownTaskRequestStatus = errors.New("unknown task request status")
	errReregistrationRequired   = errors.New("re-registration is required")
)

func (w *Worker) gotoStateAwake(ctx context.Context) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	w.state = api.WorkerStatusAwake

	w.doneWg.Add(2)
	go w.ackStateChange(ctx, w.state)
	go w.runStateAwake(ctx)
}

func (w *Worker) runStateAwake(ctx context.Context) {
	defer w.doneWg.Done()
	task := w.fetchTask(ctx)
	if task == nil {
		return
	}

	// TODO: actually execute the task
	log.Error().Interface("task", *task).Msg("task execution not implemented yet")
}

// fetchTasks periodically tries to fetch a task from the Manager, returning it when obtained.
// Returns nil when a task could not be obtained and the period loop was cancelled.
func (w *Worker) fetchTask(ctx context.Context) *api.AssignedTask {
	logger := log.With().Str("status", string(w.state)).Logger()
	logger.Info().Msg("fetching tasks")

	// Initially don't wait at all.
	var wait time.Duration

	for {
		select {
		case <-ctx.Done():
			logger.Debug().Msg("task fetching interrupted by context cancellation")
			return nil
		case <-w.doneChan:
			logger.Debug().Msg("task fetching interrupted by shutdown")
			return nil
		case <-time.After(wait):
		}
		if !w.isState(api.WorkerStatusAwake) {
			logger.Debug().Msg("task fetching interrupted by state change")
			return nil
		}

		resp, err := w.client.ScheduleTaskWithResponse(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error obtaining task")
		}
		switch {
		case resp.JSON200 != nil:
			log.Info().
				Interface("task", resp.JSON200).
				Msg("obtained task")
			return resp.JSON200
		case resp.JSON423 != nil:
			log.Info().
				Str("requestedStatus", string(resp.JSON423.StatusRequested)).
				Msg("Manager requests status change")
			w.changeState(ctx, resp.JSON423.StatusRequested)
			return nil
		case resp.JSON403 != nil:
			log.Error().
				Int("code", resp.StatusCode()).
				Str("error", string(resp.JSON403.Message)).
				Msg("access denied")
			wait = durationFetchFailed
			continue
		case resp.StatusCode() == http.StatusNoContent:
			log.Info().Msg("no task available")
			wait = durationNoTask
			continue
		default:
			log.Warn().
				Int("code", resp.StatusCode()).
				Str("error", string(resp.Body)).
				Msg("unable to obtain task for unknown reason")
			wait = durationFetchFailed
			continue
		}
	}
}
