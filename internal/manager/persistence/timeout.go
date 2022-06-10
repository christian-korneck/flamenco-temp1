package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"time"

	"git.blender.org/flamenco/pkg/api"
)

// This file contains functions for dealing with task/worker timeouts. Not database timeouts.

// FetchTimedOutTasks returns a slice of tasks that have timed out.
//
// In order to time out, a task must be in status `active` and not touched by a
// Worker since `untouchedSince`.
//
// The returned tasks also have their `Job` and `Worker` fields set.
func (db *DB) FetchTimedOutTasks(ctx context.Context, untouchedSince time.Time) ([]*Task, error) {
	result := []*Task{}
	tx := db.gormDB.WithContext(ctx).
		Model(&Task{}).
		Joins("Job").
		Joins("Worker").
		Where("tasks.status = ?", api.TaskStatusActive).
		Where("tasks.last_touched_at <= ?", untouchedSince).
		Scan(&result)
	if tx.Error != nil {
		return nil, taskError(tx.Error, "finding timed out tasks (untouched since %s)", untouchedSince.String())
	}
	return result, nil
}
