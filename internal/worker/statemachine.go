package worker

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

func (w *Worker) setupStateMachine() {
	w.stateStarters[api.WorkerStatusAsleep] = w.gotoStateAsleep
	w.stateStarters[api.WorkerStatusAwake] = w.gotoStateAwake
	w.stateStarters[api.WorkerStatusShutdown] = w.gotoStateShutdown
}

// Called whenever the Flamenco Manager has a change in current status for us.
func (w *Worker) changeState(ctx context.Context, newState api.WorkerStatus) {
	w.stateMutex.Lock()
	logger := log.With().
		Str("newState", string(newState)).
		Str("curState", string(w.state)).
		Logger()
	w.stateMutex.Unlock()

	logger.Info().Msg("state change")

	starter, ok := w.stateStarters[newState]
	if !ok {
		logger.Warn().Interface("available", w.stateStarters).Msg("no state starter for this state, going to sleep instead")
		starter = w.gotoStateAsleep
	}
	starter(ctx)
}

// Confirm that we're now in a certain state.
//
// This ACK can be given without a request from the server, for example to support
// state changes originating from UNIX signals.
//
// The state is passed as string so that this function can run independently of
// the current w.state (for thread-safety)
func (w *Worker) ackStateChange(ctx context.Context, state api.WorkerStatus) {
	defer w.doneWg.Done()

	req := api.WorkerStateChangedJSONRequestBody{Status: state}

	logger := log.With().Str("state", string(state)).Logger()
	logger.Debug().Msg("notifying Manager of our state")

	resp, err := w.client.WorkerStateChangedWithResponse(ctx, req)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to notify Manager of status change")
		return
	}

	// The 'default' response is for error cases.
	if resp.JSONDefault != nil {
		logger.Warn().
			Str("httpCode", resp.HTTPResponse.Status).
			Interface("error", resp.JSONDefault).
			Msg("error sending status change to Manager")
		return
	}
}

func (w *Worker) isState(state api.WorkerStatus) bool {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	return w.state == state
}
