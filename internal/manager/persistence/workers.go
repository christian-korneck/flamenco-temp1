package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"git.blender.org/flamenco/pkg/api"
)

type Worker struct {
	gorm.Model
	UUID   string `gorm:"type:char(36);default:'';unique;index;default:''"`
	Secret string `gorm:"type:varchar(255);default:''"`
	Name   string `gorm:"type:varchar(64);default:''"`

	Address  string           `gorm:"type:varchar(39);default:'';index"` // 39 = max length of IPv6 address.
	Platform string           `gorm:"type:varchar(16);default:''"`
	Software string           `gorm:"type:varchar(32);default:''"`
	Status   api.WorkerStatus `gorm:"type:varchar(16);default:''"`

	StatusRequested   api.WorkerStatus `gorm:"type:varchar(16);default:''"`
	LazyStatusRequest bool             `gorm:"type:smallint;default:0"`

	SupportedTaskTypes string `gorm:"type:varchar(255);default:''"` // comma-separated list of task types.
}

func (w *Worker) Identifier() string {
	return fmt.Sprintf("%s (%s)", w.Name, w.UUID)
}

// TaskTypes returns the worker's supported task types as list of strings.
func (w *Worker) TaskTypes() []string {
	return strings.Split(w.SupportedTaskTypes, ",")
}

// StatusChangeRequest stores a requested status change on the Worker.
// This just updates the Worker instance, but doesn't store the change in the
// database.
func (w *Worker) StatusChangeRequest(status api.WorkerStatus, isLazyRequest bool) {
	w.StatusRequested = status
	w.LazyStatusRequest = isLazyRequest
}

// StatusChangeClear clears the requested status change of the Worker.
// This just updates the Worker instance, but doesn't store the change in the
// database.
func (w *Worker) StatusChangeClear() {
	w.StatusRequested = ""
	w.LazyStatusRequest = false
}

func (db *DB) CreateWorker(ctx context.Context, w *Worker) error {
	if err := db.gormDB.WithContext(ctx).Create(w).Error; err != nil {
		return fmt.Errorf("creating new worker: %w", err)
	}
	return nil
}

func (db *DB) FetchWorker(ctx context.Context, uuid string) (*Worker, error) {
	w := Worker{}
	tx := db.gormDB.WithContext(ctx).
		First(&w, "uuid = ?", uuid)
	if tx.Error != nil {
		return nil, workerError(tx.Error, "fetching worker")
	}
	return &w, nil
}

func (db *DB) FetchWorkers(ctx context.Context) ([]*Worker, error) {
	workers := make([]*Worker, 0)
	tx := db.gormDB.WithContext(ctx).Model(&Worker{}).Scan(&workers)
	if tx.Error != nil {
		return nil, workerError(tx.Error, "fetching all workers")
	}
	return workers, nil
}

func (db *DB) SaveWorkerStatus(ctx context.Context, w *Worker) error {
	err := db.gormDB.WithContext(ctx).
		Model(w).
		Select("status", "status_requested").
		Updates(Worker{
			Status:          w.Status,
			StatusRequested: w.StatusRequested,
		}).Error
	if err != nil {
		return fmt.Errorf("saving worker: %w", err)
	}
	return nil
}

func (db *DB) SaveWorker(ctx context.Context, w *Worker) error {
	if err := db.gormDB.WithContext(ctx).Save(w).Error; err != nil {
		return fmt.Errorf("saving worker: %w", err)
	}
	return nil
}
