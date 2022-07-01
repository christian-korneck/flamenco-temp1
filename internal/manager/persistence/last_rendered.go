package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"

	"gorm.io/gorm/clause"
)

// LastRendered only has one entry in its database table, to indicate the job
// that was the last to receive a "last rendered image" from a Worker.
// This is used to show the global last-rendered image in the web interface.
type LastRendered struct {
	Model
	JobID uint `gorm:"default:0"`
	Job   *Job `gorm:"foreignkey:JobID;references:ID;constraint:OnDelete:CASCADE"`
}

// SetLastRendered sets this job as the one with the most recent rendered image.
func (db *DB) SetLastRendered(ctx context.Context, j *Job) error {
	render := LastRendered{
		// Always use the same database ID to ensure a single entry.
		Model: Model{ID: uint(1)},

		JobID: j.ID,
		Job:   j,
	}

	tx := db.gormDB.
		WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&render)
	return tx.Error
}

// GetLastRendered returns the UUID of the job with the most recent rendered image.
func (db *DB) GetLastRenderedJobUUID(ctx context.Context) (string, error) {
	job := Job{}
	tx := db.gormDB.WithContext(ctx).
		Joins("inner join last_rendereds LR on jobs.id = LR.job_id").
		Select("uuid").
		Find(&job)
	if tx.Error != nil {
		return "", jobError(tx.Error, "finding job with most rencent render")
	}
	return job.UUID, nil
}
