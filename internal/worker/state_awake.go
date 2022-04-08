package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

const (
	// How long to wait to fetch another task...
	durationNoTask      = 2 * time.Second  // ... if there is no task now.
	durationFetchFailed = 10 * time.Second // ... if fetching failed somehow.
)

func (w *Worker) gotoStateAwake(ctx context.Context) {
	w.stateMutex.Lock()
	w.state = api.WorkerStatusAwake
	w.stateMutex.Unlock()

	w.doneWg.Add(2)
	w.ackStateChange(ctx, w.state)

	go w.runStateAwake(ctx)
}

// runStateAwake fetches a task and executes it, in an endless loop.
func (w *Worker) runStateAwake(ctx context.Context) {
	defer func() {
		err := recover()
		if err != nil {
			w.SignOff(ctx)
			log.Panic().
				Interface("panic", err).
				Str("workerStatus", string(w.state)).
				Msg("panic, so signed off and going to stop")
		}
	}()

	defer w.doneWg.Done()
	defer log.Debug().Msg("stopping state 'awake'")

	for {
		task := w.fetchTask(ctx)
		if task == nil {
			return
		}

		// The task runner's listener will be responsible for sending results back
		// to the Manager. This code only needs to fetch a task and run it.
		err := w.taskRunner.Run(ctx, *task)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Warn().Interface("task", *task).Msg("task aborted due to context being closed")
			} else {
				log.Warn().Err(err).Interface("task", *task).Msg("error executing task")
			}
		}

		// Do some rate limiting. This is mostly useful while developing.
		time.Sleep(2 * time.Second)
	}
}

// fetchTasks periodically tries to fetch a task from the Manager, returning it when obtained.
// Returns nil when a task could not be obtained and the period loop was cancelled.
func (w *Worker) fetchTask(ctx context.Context) *api.AssignedTask {
	logger := log.With().Str("status", string(w.state)).Logger()

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

		logger.Info().Msg("fetching tasks")
		resp, err := w.client.ScheduleTaskWithResponse(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error obtaining task")
			wait = durationFetchFailed
			continue
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
		case resp.StatusCode() == http.StatusNoContent:
			log.Debug().Msg("no task available")
			// TODO: implement gradual back-off, to avoid too frequent checks when the
			// farm is idle.
			wait = durationNoTask
		default:
			log.Warn().
				Int("code", resp.StatusCode()).
				Str("error", string(resp.Body)).
				Msg("unable to obtain task for unknown reason")
			wait = durationFetchFailed
		}

	}
}
