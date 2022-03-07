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

	Address         string           `gorm:"type:varchar(39);default:'';index"` // 39 = max length of IPv6 address.
	LastActivity    string           `gorm:"type:varchar(255);default:''"`
	Platform        string           `gorm:"type:varchar(16);default:''"`
	Software        string           `gorm:"type:varchar(32);default:''"`
	Status          api.WorkerStatus `gorm:"type:varchar(16);default:''"`
	StatusRequested api.WorkerStatus `gorm:"type:varchar(16);default:''"`

	SupportedTaskTypes string `gorm:"type:varchar(255);default:''"` // comma-separated list of task types.
}

// TaskTypes returns the worker's supported task types as list of strings.
func (w *Worker) TaskTypes() []string {
	return strings.Split(w.SupportedTaskTypes, ",")
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
		return nil, tx.Error
	}
	return &w, nil
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
