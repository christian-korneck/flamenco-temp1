package stresser

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"git.blender.org/flamenco/internal/worker"
	"git.blender.org/flamenco/pkg/api"
	"github.com/rs/zerolog/log"
)

const (
	// For the actual stress test.
	durationWaitStress = 500 * time.Millisecond
	reportPeriod       = 2 * time.Second
)

var (
	numRequests = 0
	numFailed   = 0
	startTime   time.Time

	mutex = sync.RWMutex{}
)

func Run(ctx context.Context, client worker.FlamencoClient) {
	// Get a task.
	task := fetchTask(ctx, client)
	if task == nil {
		log.Fatal().Msg("error obtaining task, shutting down stresser")
		return
	}
	logger := log.With().Str("task", task.Uuid).Logger()

	// Mark the task as active.
	err := sendTaskUpdate(ctx, client, task.Uuid, api.TaskUpdate{
		Activity:   ptr("Stress testing"),
		TaskStatus: ptr(api.TaskStatusActive),
	})
	if err != nil {
		logger.Warn().Err(err).Msg("Manager rejected task becoming active. Going to stress it anyway.")
	}

	startTime = time.Now()

	go reportStatisticsLoop(ctx)

	// Do the stress test.
	var wait time.Duration
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("stresser interrupted by context cancellation")
			return
		case <-time.After(wait):
		}

		increaseNumRequests()
		err := stress(ctx, client, task)
		if err != nil {
			log.Info().Err(err).Str("task", task.Uuid).Msg("Manager rejected task update")
			increaseNumFailed()
		}

		wait = durationWaitStress
	}
}

func stress(ctx context.Context, client worker.FlamencoClient, task *api.AssignedTask) error {
	logline := "This is a log-line for stress testing. It will be repeated more than once.\n"
	bigLog := strings.Repeat(logline, 1000)

	mutex.RLock()
	update := api.TaskUpdate{
		Activity: ptr(fmt.Sprintf("stress test update %v", numRequests)),
		Log:      &bigLog,
	}
	mutex.RUnlock()

	return sendTaskUpdate(ctx, client, task.Uuid, update)
}

func ptr[T any](value T) *T {
	return &value
}

func increaseNumRequests() {
	mutex.Lock()
	defer mutex.Unlock()
	numRequests++
}

func increaseNumFailed() {
	mutex.Lock()
	defer mutex.Unlock()
	numFailed++
}

func reportStatistics() {
	mutex.RLock()
	defer mutex.RUnlock()

	duration := time.Since(startTime)
	durationInSeconds := float64(duration) / float64(time.Second)
	reqPerSecond := float64(numRequests) / durationInSeconds

	log.Info().
		Int("numRequests", numRequests).
		Int("numFailed", numFailed).
		Str("duration", duration.String()).
		Float64("requestsPerSecond", reqPerSecond).
		Msg("stress progress")
}

func reportStatisticsLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(reportPeriod):
			reportStatistics()
		}
	}
}
