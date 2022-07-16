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
	DaysOfWeek string `gorm:"default:''"`
	StartTime  string `gorm:"default:''"`
	EndTime    string `gorm:"default:''"`

	NextCheck *time.Time
}

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

func (db *DB) SetWorkerSleepSchedule(ctx context.Context, workerUUID string, schedule SleepSchedule) error {
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
