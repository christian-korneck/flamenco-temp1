package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm/clause"
)

// SleepSchedule belongs to a Worker, and determines when it's automatically
// sent to the 'asleep' and 'awake' states.
type SleepSchedule struct {
	Model

	WorkerID uint    `gorm:"default:0;unique;index"`
	Worker   *Worker `gorm:"foreignkey:WorkerID;references:ID;constraint:OnDelete:CASCADE"`

	IsActive bool `gorm:"default:false;index"`

	// Space-separated two-letter strings indicating days of week the schedule is
	// active ("mo", "tu", etc.). Empty means "every day".
	DaysOfWeek string    `gorm:"default:''"`
	StartTime  TimeOfDay `gorm:"default:''"`
	EndTime    TimeOfDay `gorm:"default:''"`

	NextCheck time.Time
}

// FetchWorkerSleepSchedule fetches the worker's sleep schedule.
// It does not fetch the worker itself. If you need that, call
// `FetchSleepScheduleWorker()` afterwards.
func (db *DB) FetchWorkerSleepSchedule(ctx context.Context, workerUUID string) (*SleepSchedule, error) {
	logger := log.With().Str("worker", workerUUID).Logger()
	logger.Trace().Msg("fetching worker sleep schedule")

	var sched SleepSchedule
	tx := db.gormDB.WithContext(ctx).
		Joins("inner join workers on workers.id = sleep_schedules.worker_id").
		Where("workers.uuid = ?", workerUUID).
		// This is the same as First(&sched), except it doesn't cause an error if it doesn't exist:
		Limit(1).Find(&sched)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if sched.ID == 0 {
		return nil, nil
	}
	return &sched, nil
}

func (db *DB) SetWorkerSleepSchedule(ctx context.Context, workerUUID string, schedule *SleepSchedule) error {
	logger := log.With().Str("worker", workerUUID).Logger()
	logger.Trace().Msg("setting worker sleep schedule")

	worker, err := db.FetchWorker(ctx, workerUUID)
	if err != nil {
		return fmt.Errorf("fetching worker %q: %w", workerUUID, err)
	}
	schedule.WorkerID = worker.ID
	schedule.Worker = worker

	tx := db.gormDB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "worker_id"}},
			UpdateAll: true,
		}).
		Create(&schedule)
	return tx.Error
}

func (db *DB) SetWorkerSleepScheduleNextCheck(ctx context.Context, schedule *SleepSchedule) error {
	tx := db.gormDB.WithContext(ctx).
		Select("next_check").
		Updates(schedule)
	return tx.Error
}

// FetchSleepScheduleWorker sets the given schedule's `Worker` pointer.
func (db *DB) FetchSleepScheduleWorker(ctx context.Context, schedule *SleepSchedule) error {
	var worker Worker
	tx := db.gormDB.WithContext(ctx).First(&worker, schedule.WorkerID)
	if tx.Error != nil {
		return workerError(tx.Error, "finding worker by their sleep schedule")
	}
	schedule.Worker = &worker
	return nil
}

// FetchSleepSchedulesToCheck returns the sleep schedules that are due for a check.
func (db *DB) FetchSleepSchedulesToCheck(ctx context.Context) ([]*SleepSchedule, error) {
	log.Trace().Msg("fetching sleep schedules that need checking")

	now := db.gormDB.NowFunc()

	schedules := []*SleepSchedule{}
	tx := db.gormDB.WithContext(ctx).
		Model(&SleepSchedule{}).
		Where("is_active = ?", true).
		Where("next_check <= ? or next_check is NULL or next_check = ''", now).
		Scan(&schedules)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return schedules, nil
}
