package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"time"

	"git.blender.org/flamenco/pkg/api"
)

// This file contains functions for dealing with task/worker timeouts. Not database timeouts.

// workerStatusNoTimeout contains the worker statuses that are exempt from
// timeout checking. A worker in any other status will be subject to the timeout
// check.
var workerStatusNoTimeout = []api.WorkerStatus{
	api.WorkerStatusError,
	api.WorkerStatusOffline,
}

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

func (db *DB) FetchTimedOutWorkers(ctx context.Context, lastSeenBefore time.Time) ([]*Worker, error) {
	result := []*Worker{}
	tx := db.gormDB.WithContext(ctx).
		Model(&Worker{}).
		Where("workers.status not in ?", workerStatusNoTimeout).
		Where("workers.last_seen_at <= ?", lastSeenBefore).
		Scan(&result)
	if tx.Error != nil {
		return nil, workerError(tx.Error, "finding timed out workers (last seen before %s)", lastSeenBefore.String())
	}
	return result, nil
}
