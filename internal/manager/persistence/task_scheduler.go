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

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
	"gorm.io/gorm"
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
	logger.Debug().Msg("finding task for worker")

	// Run two queries in one transaction:
	// 1. find task, and
	// 2. assign the task to the worker.
	var task *Task
	txErr := db.gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		task, err = findTaskForWorker(tx, w)
		if err != nil {
			return fmt.Errorf("error finding task for worker: %w", err)
		}
		if task == nil {
			// No task found, which is fine.
			return nil
		}

		// Found a task, now assign it to the requesting worker.
		// Without the Select() call, Gorm will try and also store task.Job in the jobs database, which is not what we want.
		if err := assignTaskToWorker(tx, w, task); err != nil {
			logger.Warn().
				Str("taskID", task.UUID).
				Err(err).
				Msg("error assigning task to worker")
			return fmt.Errorf("error assigning task to worker: %w", err)
		}

		return nil
	})

	if txErr != nil {
		logger.Error().Err(txErr).Msg("error finding task for worker")
		return nil, fmt.Errorf("error finding task for worker: %w", txErr)
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
		Order("jobs.priority desc"). // Highest job priority
		Order("priority desc").      // Highest task priority
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
