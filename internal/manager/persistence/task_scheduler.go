package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"git.blender.org/flamenco/pkg/api"
)

var (
	schedulableTaskStatuses = []api.TaskStatus{api.TaskStatusQueued, api.TaskStatusSoftFailed, api.TaskStatusActive}
	completedTaskStatuses   = []api.TaskStatus{api.TaskStatusCompleted}
	schedulableJobStatuses  = []api.JobStatus{api.JobStatusActive, api.JobStatusQueued, api.JobStatusRequeued}
)

// ScheduleTask finds a task to execute by the given worker.
// If no task is available, (nil, nil) is returned, as this is not an error situation.
// NOTE: this does not also fetch returnedTask.Worker, but returnedTask.WorkerID is set.
func (db *DB) ScheduleTask(ctx context.Context, w *Worker) (*Task, error) {
	logger := log.With().Str("worker", w.UUID).Logger()
	logger.Trace().Msg("finding task for worker")

	// Run two queries in one transaction:
	// 1. find task, and
	// 2. assign the task to the worker.
	var task *Task
	txErr := db.gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		task, err = findTaskForWorker(tx, w)
		if err != nil {
			logger.Error().Err(err).Msg("finding task for worker")
			return fmt.Errorf("finding task for worker: %w", err)
		}
		if task == nil {
			// No task found, which is fine.
			return nil
		}

		// Found a task, now assign it to the requesting worker.
		if err := assignTaskToWorker(tx, w, task); err != nil {
			logger.Warn().
				Str("taskID", task.UUID).
				Err(err).
				Msg("assigning task to worker")
			return fmt.Errorf("assigning task to worker: %w", err)
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	if task == nil {
		logger.Debug().Msg("no task for worker")
		return nil, nil
	}

	logger.Info().
		Str("taskID", task.UUID).
		Msg("assigned task to worker")

	return task, nil
}

func findTaskForWorker(tx *gorm.DB, w *Worker) (*Task, error) {
	task := Task{}
	findTaskResult := tx.
		Model(&task).
		Joins("left join jobs on tasks.job_id = jobs.id").
		Joins("left join task_dependencies on tasks.id = task_dependencies.task_id").
		Joins("left join tasks as tdeps on tdeps.id = task_dependencies.dependency_id").
		Where("tasks.status in ?", schedulableTaskStatuses).                       // Schedulable task statuses
		Where("tdeps.status in ? or tdeps.status is NULL", completedTaskStatuses). // Dependencies completed
		Where("jobs.status in ?", schedulableJobStatuses).                         // Schedulable job statuses
		Where("tasks.type in ?", w.TaskTypes()).                                   // Supported task types
		Where("tasks.worker_id = ? or tasks.worker_id is NULL", w.ID).             // assigned to this worker or not assigned at all
		// TODO: Non-blacklisted
		Order("jobs.priority desc").  // Highest job priority
		Order("tasks.priority desc"). // Highest task priority
		Limit(1).
		Preload("Job").
		Find(&task)

	if findTaskResult.Error != nil {
		return nil, findTaskResult.Error
	}
	if task.ID == 0 {
		// No task fetched, which doesn't result in an error with Limt(1).Find(&task).
		return nil, nil
	}

	return &task, nil
}

func assignTaskToWorker(tx *gorm.DB, w *Worker, t *Task) error {
	// Without the Select() call, Gorm will try and also store task.Job in the
	// jobs database, which is not what we want.
	return tx.Model(t).Select("worker_id").Updates(Task{WorkerID: &w.ID}).Error
}
