package stresser

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/worker"
	"git.blender.org/flamenco/pkg/api"
)

const (
	// For fetching the task to stress test.
	durationNoTask      = 1 * time.Second // ... if there is no task now.
	durationFetchFailed = 2 * time.Second // ... if fetching failed somehow.
)

var (
	ErrTaskReassigned     = worker.ErrTaskReassigned
	ErrTaskUpdateRejected = errors.New("task update was rejected")
)

func GetFlamencoClient(
	ctx context.Context,
	config worker.WorkerConfigWithCredentials,
) worker.FlamencoClient {
	startupCtx, startupCtxCancel := context.WithTimeout(ctx, 10*time.Second)
	defer startupCtxCancel()

	client, startupState := worker.RegisterOrSignOn(startupCtx, config)
	if startupState != api.WorkerStatusAwake {
		log.Fatal().Str("requestedStartupState", string(startupState)).Msg("stresser should always be awake")
	}

	ackStateChange(ctx, client, startupState)

	return client
}

func fetchTask(ctx context.Context, client worker.FlamencoClient) *api.AssignedTask {
	// Initially don't wait at all.
	var wait time.Duration

	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("task fetching interrupted by context cancellation")
			return nil
		case <-time.After(wait):
		}

		log.Debug().Msg("fetching tasks")
		resp, err := client.ScheduleTaskWithResponse(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error obtaining task")
			wait = durationFetchFailed
			continue
		}
		switch {
		case resp.JSON200 != nil:
			return resp.JSON200
		case resp.JSON423 != nil:
			log.Fatal().Str("requestedStatus", string(resp.JSON423.StatusRequested)).
				Msg("Manager requests status change, stresser does not support this")
			return nil
		case resp.JSON403 != nil:
			log.Error().
				Int("code", resp.StatusCode()).
				Str("error", string(resp.JSON403.Message)).
				Msg("access denied")
			wait = durationFetchFailed
		case resp.StatusCode() == http.StatusNoContent:
			log.Debug().Msg("no task available")
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

func ackStateChange(ctx context.Context, client worker.FlamencoClient, state api.WorkerStatus) {
	req := api.WorkerStateChangedJSONRequestBody{Status: state}

	logger := log.With().Str("state", string(state)).Logger()
	logger.Debug().Msg("notifying Manager of our state")

	resp, err := client.WorkerStateChangedWithResponse(ctx, req)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to notify Manager of status change")
		return
	}

	// The 'default' response is for error cases.
	if resp.JSONDefault != nil {
		logger.Fatal().
			Str("httpCode", resp.HTTPResponse.Status).
			Interface("error", resp.JSONDefault).
			Msg("error sending status change to Manager")
		return
	}
}

func sendTaskUpdate(ctx context.Context, client worker.FlamencoClient, taskID string, update api.TaskUpdate) error {
	resp, err := client.TaskUpdateWithResponse(ctx, taskID, api.TaskUpdateJSONRequestBody(update))
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusConflict:
		return worker.ErrTaskReassigned
	default:
		return fmt.Errorf("%w: task=%s", ErrTaskUpdateRejected, taskID)
	}
}
