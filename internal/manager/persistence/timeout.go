package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"time"

	"git.blender.org/flamenco/pkg/api"
)

// This file contains functions for dealing with task/worker timeouts. Not database timeouts.

func (db *DB) FetchTimedOutTasks(ctx context.Context, untouchedSince time.Time) ([]*Task, error) {
	result := []*Task{}
	tx := db.gormDB.WithContext(ctx).
		Model(&Task{}).
		Joins("Job").
		Where("tasks.status = ?", api.TaskStatusActive).
		Where("tasks.last_touched_at <= ?", untouchedSince).
		Scan(&result)
	if tx.Error != nil {
		return nil, taskError(tx.Error, "finding timed out tasks (untouched since %s)", untouchedSince.String())
	}
	return result, nil
}
