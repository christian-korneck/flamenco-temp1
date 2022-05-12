package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"time"

	"git.blender.org/flamenco/pkg/api"
)

const durationSleepCheck = 3 * time.Second

func (w *Worker) gotoStateAsleep(ctx context.Context) {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()

	w.state = api.WorkerStatusAsleep
	w.doneWg.Add(2)
	w.ackStateChange(ctx, w.state)
	go w.runStateAsleep(ctx)
}

func (w *Worker) runStateAsleep(ctx context.Context) {
	defer w.doneWg.Done()
	logger := w.loggerWithStatus()
	logger.Info().Msg("sleeping")

	for {
		select {
		case <-ctx.Done():
			logger.Debug().Msg("asleep state interrupted by context cancellation")
			return
		case <-w.doneChan:
			logger.Debug().Msg("asleep state interrupted by shutdown")
			return
		case <-time.After(durationSleepCheck):
			if w.changeStateIfRequested(ctx) {
				return
			}
		}
	}
}
