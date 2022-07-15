package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"git.blender.org/flamenco/pkg/api"
)

// SPDX-License-Identifier: GPL-3.0-or-later

// TaskUpdate is a queued task update.
type TaskUpdate struct {
	Model

	TaskID  string `gorm:"type:varchar(36);default:''"`
	Payload []byte `gorm:"type:BLOB"`
}

func (t *TaskUpdate) Unmarshal() (*api.TaskUpdateJSONRequestBody, error) {
	var apiTaskUpdate api.TaskUpdateJSONRequestBody
	if err := json.Unmarshal(t.Payload, &apiTaskUpdate); err != nil {
		return nil, err
	}
	return &apiTaskUpdate, nil
}

// UpstreamBufferQueueSize returns how many task updates are queued in the upstream buffer.
func (db *DB) UpstreamBufferQueueSize(ctx context.Context) (int, error) {
	var queueSize int64
	tx := db.gormDB.WithContext(ctx).
		Model(&TaskUpdate{}).
		Count(&queueSize)
	if tx.Error != nil {
		return 0, fmt.Errorf("counting queued task updates: %w", tx.Error)
	}
	return int(queueSize), nil
}

// UpstreamBufferQueue queues a task update in the upstrema buffer.
func (db *DB) UpstreamBufferQueue(ctx context.Context, taskID string, apiTaskUpdate api.TaskUpdateJSONRequestBody) error {
	blob, err := json.Marshal(apiTaskUpdate)
	if err != nil {
		return fmt.Errorf("converting task update to JSON: %w", err)
	}

	taskUpdate := TaskUpdate{
		TaskID:  taskID,
		Payload: blob,
	}

	tx := db.gormDB.WithContext(ctx).Create(&taskUpdate)
	return tx.Error
}

// UpstreamBufferFrontItem returns the first-queued item. The item remains queued.
func (db *DB) UpstreamBufferFrontItem(ctx context.Context) (*TaskUpdate, error) {
	taskUpdate := TaskUpdate{}

	findResult := db.gormDB.WithContext(ctx).
		Order("ID").
		Limit(1).
		Find(&taskUpdate)
	if findResult.Error != nil {
		return nil, findResult.Error
	}
	if taskUpdate.ID == 0 {
		// No update fetched, which doesn't result in an error with Limt(1).Find(&task).
		return nil, nil
	}

	return &taskUpdate, nil
}

// UpstreamBufferDiscard discards the queued task update with the given row ID.
func (db *DB) UpstreamBufferDiscard(ctx context.Context, queuedTaskUpdate *TaskUpdate) error {
	tx := db.gormDB.WithContext(ctx).Delete(queuedTaskUpdate)
	return tx.Error
}
