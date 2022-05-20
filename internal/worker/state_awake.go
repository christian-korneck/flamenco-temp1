package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

const (
	// How long to wait to fetch another task...
	durationNoTask       = 2 * time.Second  // ... if there is no task now.
	durationFetchFailed  = 10 * time.Second // ... if fetching failed somehow.
	durationTaskComplete = 2 * time.Second  // ... when a task was completed.

	mayKeepRunningPeriod = 10 * time.Second
)

// Implement error interface for `api.MayKeepRunning` to indicate a task run was
// aborted due to the Manager saying "NO".
type taskRunAborted api.MayKeepRunning

func (tra taskRunAborted) Error() string {
	switch {
	case tra.MayKeepRunning:
		return "task could have been kept running"
	case tra.StatusChangeRequested:
		return "worker status change requested"
	case tra.Reason == "":
		return "manager said NO"
	}
	return tra.Reason
}

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
			logger := w.loggerWithStatus()
			logger.Panic().
				Interface("panic", err).
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
		err := w.runTask(ctx, *task)
		if err != nil {
			var abortError taskRunAborted
			if errors.As(err, &abortError) {
				log.Warn().
					Str("task", task.Uuid).
					Str("reason", err.Error()).
					Msg("task aborted by request of Manager")
			} else if errors.Is(err, context.Canceled) {
				log.Warn().Interface("task", *task).Msg("task aborted due to context being closed")
			} else {
				log.Warn().Err(err).Interface("task", *task).Msg("error executing task")
			}
		}

		// Do some rate limiting. This is mostly useful while developing.
		time.Sleep(durationTaskComplete)
	}
}

// fetchTasks periodically tries to fetch a task from the Manager, returning it when obtained.
// Returns nil when a task could not be obtained and the period loop was cancelled.
func (w *Worker) fetchTask(ctx context.Context) *api.AssignedTask {
	logger := w.loggerWithStatus()

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

		logger.Debug().Msg("fetching tasks")
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

// runTask runs the given task.
func (w *Worker) runTask(ctx context.Context, task api.AssignedTask) error {
	// Create a sub-context to manage the life-span of both the running of the
	// task and the loop to check whether we're still allowed to run it.
	taskCtx, taskCancel := context.WithCancel(ctx)
	defer taskCancel()

	var taskRunnerErr, abortReason error

	// Run the actual task in a separate goroutine.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer taskCancel()
		taskRunnerErr = w.taskRunner.Run(taskCtx, task)
	}()

	// Do a periodic check to see if we're actually allowed to run this task.
checkloop:
	for {
		select {
		case <-taskCtx.Done():
			// The task is done, no more need to check.
			break checkloop
		case <-time.After(mayKeepRunningPeriod):
			// Time to do another check.
			break
		}

		mkr := w.mayIKeepRunning(taskCtx, task.Uuid)
		if mkr.MayKeepRunning {
			continue
		}

		abortReason = taskRunAborted(mkr)
		taskCancel()
		break checkloop
	}

	// Wait for the task runner to either complete or abort.
	wg.Wait()

	if abortReason != nil {
		return abortReason
	}

	return taskRunnerErr
}
