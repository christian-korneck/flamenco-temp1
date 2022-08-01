package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"math"
	"time"

	"gorm.io/gorm/clause"
)

// JobBlock keeps track of which Worker is not allowed to run which task type on which job.
type JobBlock struct {
	// Don't include the standard Gorm UpdatedAt or DeletedAt fields, as they're useless here.
	// Entries will never be updated, and should never be soft-deleted but just purged from existence.
	ID        uint
	CreatedAt time.Time

	JobID uint `gorm:"default:0;uniqueIndex:job_worker_tasktype"`
	Job   *Job `gorm:"foreignkey:JobID;references:ID;constraint:OnDelete:CASCADE"`

	WorkerID uint    `gorm:"default:0;uniqueIndex:job_worker_tasktype"`
	Worker   *Worker `gorm:"foreignkey:WorkerID;references:ID;constraint:OnDelete:CASCADE"`

	TaskType string `gorm:"uniqueIndex:job_worker_tasktype"`
}

// AddWorkerToJobBlocklist prevents this Worker of getting any task, of this type, on this job, from the task scheduler.
func (db *DB) AddWorkerToJobBlocklist(ctx context.Context, job *Job, worker *Worker, taskType string) error {
	entry := JobBlock{
		Job:      job,
		Worker:   worker,
		TaskType: taskType,
	}
	tx := db.gormDB.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&entry)
	return tx.Error
}

func (db *DB) FetchJobBlocklist(ctx context.Context, jobUUID string) ([]JobBlock, error) {
	entries := []JobBlock{}

	tx := db.gormDB.WithContext(ctx).
		Model(JobBlock{}).
		Joins("inner join jobs on jobs.id = job_blocks.job_id").
		Joins("Worker").
		Where("jobs.uuid = ?", jobUUID).
		Order("Worker.name").
		Scan(&entries)
	return entries, tx.Error
}

// ClearJobBlocklist removes the entire blocklist of this job.
func (db *DB) ClearJobBlocklist(ctx context.Context, job *Job) error {
	tx := db.gormDB.WithContext(ctx).
		Where("job_id = ?", job.ID).
		Delete(JobBlock{})
	return tx.Error
}

func (db *DB) RemoveFromJobBlocklist(ctx context.Context, jobUUID, workerUUID, taskType string) error {
	// Find the job ID.
	job := Job{}
	tx := db.gormDB.WithContext(ctx).
		Select("id").
		Where("uuid = ?", jobUUID).
		Find(&job)
	if tx.Error != nil {
		return jobError(tx.Error, "fetching job with uuid=%q", jobUUID)
	}

	// Find the worker ID.
	worker := Worker{}
	tx = db.gormDB.WithContext(ctx).
		Select("id").
		Where("uuid = ?", workerUUID).
		Find(&worker)
	if tx.Error != nil {
		return workerError(tx.Error, "fetching worker with uuid=%q", workerUUID)
	}

	// Remove the blocklist entry.
	tx = db.gormDB.WithContext(ctx).
		Where("job_id = ?", job.ID).
		Where("worker_id = ?", worker.ID).
		Where("task_type = ?", taskType).
		Delete(JobBlock{})
	return tx.Error
}

// WorkersLeftToRun returns a set of worker UUIDs that can run tasks of the given type on the given job.
//
// NOTE: this does NOT consider the task failure list, which blocks individual
// workers from individual tasks. This is ONLY concerning the job blocklist.
func (db *DB) WorkersLeftToRun(ctx context.Context, job *Job, taskType string) (map[string]bool, error) {
	// Find the IDs of the workers blocked on this job + tasktype combo.
	blockedWorkers := db.gormDB.
		Table("workers as blocked_workers").
		Select("blocked_workers.id").
		Joins("inner join job_blocks JB on blocked_workers.id = JB.worker_id").
		Where("JB.job_id = ?", job.ID).
		Where("JB.task_type = ?", taskType)

	// Find the workers NOT blocked.
	workers := []*Worker{}
	tx := db.gormDB.WithContext(ctx).
		Model(&Worker{}).
		Select("uuid").
		Where("id not in (?)", blockedWorkers).
		Scan(&workers)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// From the list of workers, construct the map of UUIDs.
	uuidMap := map[string]bool{}
	for _, worker := range workers {
		uuidMap[worker.UUID] = true
	}

	return uuidMap, nil
}

// CountTaskFailuresOfWorker returns the number of task failures of this worker, on this particular job and task type.
func (db *DB) CountTaskFailuresOfWorker(ctx context.Context, job *Job, worker *Worker, taskType string) (int, error) {
	var numFailures int64

	tx := db.gormDB.WithContext(ctx).
		Model(&TaskFailure{}).
		Joins("inner join tasks T on task_failures.task_id = T.id").
		Where("task_failures.worker_id = ?", worker.ID).
		Where("T.job_id = ?", job.ID).
		Where("T.type = ?", taskType).
		Count(&numFailures)

	if numFailures > math.MaxInt {
		panic("overflow error in number of failures")
	}

	return int(numFailures), tx.Error
}
