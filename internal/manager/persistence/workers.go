package persistence

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"git.blender.org/flamenco/pkg/api"
)

type Worker struct {
	gorm.Model
	UUID   string `gorm:"type:char(36);not null;unique;index"`
	Secret string `gorm:"type:varchar(255);not null"`
	Name   string `gorm:"type:varchar(64);not null"`

	Address         string           `gorm:"type:varchar(39);not null;index"` // 39 = max length of IPv6 address.
	LastActivity    string           `gorm:"type:varchar(255);not null"`
	Platform        string           `gorm:"type:varchar(16);not null"`
	Software        string           `gorm:"type:varchar(32);not null"`
	Status          api.WorkerStatus `gorm:"type:varchar(16);not null"`
	StatusRequested api.WorkerStatus `gorm:"type:varchar(16);not null;default:''"`

	SupportedTaskTypes string `gorm:"type:varchar(255);not null"` // comma-separated list of task types.
}

// TaskTypes returns the worker's supported task types as list of strings.
func (w *Worker) TaskTypes() []string {
	return strings.Split(w.SupportedTaskTypes, ",")
}

func (db *DB) CreateWorker(ctx context.Context, w *Worker) error {
	if err := db.gormDB.WithContext(ctx).Create(w).Error; err != nil {
		return fmt.Errorf("error creating new worker: %w", err)
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
		return fmt.Errorf("error saving worker: %w", err)
	}
	return nil
}

func (db *DB) SaveWorker(ctx context.Context, w *Worker) error {
	if err := db.gormDB.WithContext(ctx).Save(w).Error; err != nil {
		return fmt.Errorf("error saving worker: %w", err)
	}
	return nil
}
