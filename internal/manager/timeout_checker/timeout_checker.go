package timeout_checker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/rs/zerolog/log"
)

// Interval for checking all active tasks for timeouts.
const timeoutCheckInterval = 1 * time.Minute

// Delay for the intial check. This gives workers a chance to reconnect to the Manager
// and send updates after the Manager has started.
const timeoutInitialSleep = 5 * time.Minute

// TimeoutChecker periodically times out tasks and workers if the worker hasn't sent any update recently.
type TimeoutChecker struct {
	taskTimeout   time.Duration
	workerTimeout time.Duration

	clock            clock.Clock
	persist          PersistenceService
	taskStateMachine TaskStateMachine
	logStorage       LogStorage
	broadcaster      ChangeBroadcaster
}

// New creates a new TimeoutChecker.
func New(
	taskTimeout,
	workerTimeout time.Duration,
	clock clock.Clock,
	persist PersistenceService,
	taskStateMachine TaskStateMachine,
	logStorage LogStorage,
	broadcaster ChangeBroadcaster,
) *TimeoutChecker {
	return &TimeoutChecker{
		taskTimeout:   taskTimeout,
		workerTimeout: workerTimeout,

		clock:            clock,
		persist:          persist,
		taskStateMachine: taskStateMachine,
		logStorage:       logStorage,
		broadcaster:      broadcaster,
	}
}

// Run runs the timeout checker until the context closes.
func (ttc *TimeoutChecker) Run(ctx context.Context) {
	defer log.Info().Msg("TimeoutChecker: shutting down")

	if ttc.taskTimeout == 0 {
		log.Warn().Msg("TimeoutChecker: no timeout duration configured, will not check for task timeouts")
		return
	}

	log.Info().
		Str("taskTimeout", ttc.taskTimeout.String()).
		Str("initialSleep", timeoutInitialSleep.String()).
		Str("checkInterval", timeoutCheckInterval.String()).
		Msg("TimeoutChecker: starting up")

	// Start with a delay, so that workers get a chance to push their updates
	// after the manager has started up.
	waitDur := timeoutInitialSleep

	for {
		select {
		case <-ctx.Done():
			return
		case <-ttc.clock.After(waitDur):
			waitDur = timeoutCheckInterval
		}
		ttc.checkTasks(ctx)
		ttc.checkWorkers(ctx)
	}
}
