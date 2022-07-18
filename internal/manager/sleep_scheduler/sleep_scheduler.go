package sleep_scheduler

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

// Time period for checking the schedule of every worker.
const checkInterval = 1 * time.Minute

// SleepScheduler manages wake/sleep cycles of Workers.
type SleepScheduler struct {
	clock       clock.Clock
	persist     PersistenceService
	broadcaster ChangeBroadcaster
}

// New creates a new SleepScheduler.
func New(clock clock.Clock, persist PersistenceService, broadcaster ChangeBroadcaster) *SleepScheduler {
	return &SleepScheduler{
		clock:       clock,
		persist:     persist,
		broadcaster: broadcaster,
	}
}

// Run occasionally checks the sleep schedule and updates workers.
// It stops running when the context closes.
func (ss *SleepScheduler) Run(ctx context.Context) {
	log.Info().
		Str("checkInterval", checkInterval.String()).
		Msg("sleep scheduler starting")
	defer log.Info().Msg("sleep scheduler shutting down")

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(checkInterval):
			ss.CheckSchedules(ctx)
		}
	}
}

func (ss *SleepScheduler) FetchSchedule(ctx context.Context, workerUUID string) (*persistence.SleepSchedule, error) {
	return ss.persist.FetchWorkerSleepSchedule(ctx, workerUUID)
}

// SetSleepSchedule stores the given schedule as the worker's new sleep schedule.
// The new schedule is immediately applied to the Worker.
func (ss *SleepScheduler) SetSchedule(ctx context.Context, workerUUID string, schedule *persistence.SleepSchedule) error {
	// Ensure 'start' actually preceeds 'end'.
	if schedule.StartTime.HasValue() &&
		schedule.EndTime.HasValue() &&
		schedule.EndTime.IsBefore(schedule.StartTime) {
		schedule.StartTime, schedule.EndTime = schedule.EndTime, schedule.StartTime
	}

	schedule.DaysOfWeek = cleanupDaysOfWeek(schedule.DaysOfWeek)
	schedule.NextCheck = ss.calculateNextCheck(schedule)

	if err := ss.persist.SetWorkerSleepSchedule(ctx, workerUUID, schedule); err != nil {
		return fmt.Errorf("persisting sleep schedule of worker %s: %w", workerUUID, err)
	}

	return ss.ApplySleepSchedule(ctx, schedule)
}

// WorkerStatus returns the status the worker should be in right now, according to its schedule.
// If the worker has no schedule active, returns 'awake'.
func (ss *SleepScheduler) WorkerStatus(ctx context.Context, workerUUID string) (api.WorkerStatus, error) {
	schedule, err := ss.persist.FetchWorkerSleepSchedule(ctx, workerUUID)
	if err != nil {
		return "", err
	}
	return ss.scheduledWorkerStatus(schedule), nil
}

// scheduledWorkerStatus returns the expected worker status for the current date/time.
func (ss *SleepScheduler) scheduledWorkerStatus(sched *persistence.SleepSchedule) api.WorkerStatus {
	now := ss.clock.Now()
	return scheduledWorkerStatus(now, sched)
}

// Return a timestamp when the next scheck for this schedule is due.
func (ss *SleepScheduler) calculateNextCheck(schedule *persistence.SleepSchedule) time.Time {
	now := ss.clock.Now()
	return calculateNextCheck(now, schedule)
}

// ApplySleepSchedule sets worker.StatusRequested if the scheduler demands a status change.
func (ss *SleepScheduler) ApplySleepSchedule(ctx context.Context, schedule *persistence.SleepSchedule) error {
	// Find the Worker managed by this schedule.
	worker := schedule.Worker
	if worker == nil {
		err := ss.persist.FetchSleepScheduleWorker(ctx, schedule)
		if err != nil {
			return err
		}
		worker = schedule.Worker
	}

	scheduled := ss.scheduledWorkerStatus(schedule)
	if scheduled == "" ||
		(worker.StatusRequested == scheduled && !worker.LazyStatusRequest) ||
		(worker.Status == scheduled && worker.StatusRequested == "") {
		// The worker is already in the right state, or is non-lazily requested to
		// go to the right state, so nothing else has to be done.
		return nil
	}

	logger := log.With().
		Str("worker", worker.Identifier()).
		Str("currentStatus", string(worker.Status)).
		Str("scheduledStatus", string(scheduled)).
		Logger()

	if worker.StatusRequested != "" {
		logger.Info().Str("oldStatusRequested", string(worker.StatusRequested)).
			Msg("sleep scheduler: overruling previously requested status with scheduled status")
	} else {
		logger.Info().Msg("sleep scheduler: requesting worker to switch to scheduled status")
	}

	if err := ss.updateWorkerStatus(ctx, worker, scheduled); err != nil {
		return err
	}
	return nil
}

func (ss *SleepScheduler) updateWorkerStatus(
	ctx context.Context,
	worker *persistence.Worker,
	newStatus api.WorkerStatus,
) error {
	// Sleep schedule should be adhered to immediately, no lazy requests.
	// A render task can run for hours, so better to not wait for it.
	worker.StatusChangeRequest(newStatus, false)

	err := ss.persist.SaveWorkerStatus(ctx, worker)
	if err != nil {
		return fmt.Errorf("error saving worker %s to database: %w", worker.Identifier(), err)
	}

	// Broadcast worker change via SocketIO
	ss.broadcaster.BroadcastWorkerUpdate(api.SocketIOWorkerUpdate{
		Id:     worker.UUID,
		Name:   worker.Name,
		Status: worker.Status,
		StatusChange: &api.WorkerStatusChangeRequest{
			IsLazy: false,
			Status: worker.StatusRequested,
		},
		Updated: worker.UpdatedAt,
		Version: worker.Software,
	})

	return nil
}

// CheckSchedules updates the status of all workers for which a schedule is active.
func (ss *SleepScheduler) CheckSchedules(ctx context.Context) {
	toCheck, err := ss.persist.FetchSleepSchedulesToCheck(ctx)
	if err != nil {
		log.Error().Err(err).Msg("sleep scheduler: unable to fetch sleep schedules")
		return
	}
	if len(toCheck) == 0 {
		log.Trace().Msg("sleep scheduler: no sleep schedules need checking")
		return
	}

	log.Debug().Int("numWorkers", len(toCheck)).Msg("sleep scheduler: checking worker sleep schedules")

	for _, schedule := range toCheck {
		ss.checkSchedule(ctx, schedule)
	}
}

func (ss *SleepScheduler) checkSchedule(ctx context.Context, schedule *persistence.SleepSchedule) {
	// Compute the next time to check.
	schedule.NextCheck = ss.calculateNextCheck(schedule)
	if err := ss.persist.SetWorkerSleepScheduleNextCheck(ctx, schedule); err != nil {
		log.Error().
			Err(err).
			Str("worker", schedule.Worker.Identifier()).
			Msg("sleep scheduler: error refreshing worker's sleep schedule")
		return
	}

	// Apply the schedule to the worker.
	if err := ss.ApplySleepSchedule(ctx, schedule); err != nil {
		log.Error().
			Err(err).
			Str("worker", schedule.Worker.Identifier()).
			Msg("sleep scheduler: error applying worker's sleep schedule")
		return
	}
}
