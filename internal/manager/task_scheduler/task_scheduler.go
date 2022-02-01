// Package task_scheduler can choose which task to assign to a worker.
package task_scheduler

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
	"errors"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
	"gorm.io/gorm"
)

var (
	schedulableTaskStatuses = []api.TaskStatus{api.TaskStatusQueued, api.TaskStatusSoftFailed}
	completedTaskStatuses   = []api.TaskStatus{api.TaskStatusCompleted}
	schedulableJobStatuses  = []api.JobStatus{api.JobStatusActive, api.JobStatusQueued, api.JobStatusRequeued}
)

type TaskScheduler struct {
	db PersistenceService
}

type PersistenceService interface {
	GormDB() *gorm.DB
}

func NewTaskScheduler(db PersistenceService) *TaskScheduler {
	return &TaskScheduler{db}
}

// ScheduleTask finds a task to execute by the given worker.
// If no task is available, (nil, nil) is returned, as this is not an error situation.
func (ts *TaskScheduler) ScheduleTask(w *persistence.Worker) (*persistence.Task, error) {
	task, err := ts.findTaskForWorker(w)

	// TODO: Mark the task as Active, and push the status change to whatever I think up to handle those changes.
	// TODO: Store in the database that this task is assigned to this worker.

	return task, err
}

func (ts *TaskScheduler) findTaskForWorker(w *persistence.Worker) (*persistence.Task, error) {

	logger := log.With().Str("worker", w.UUID).Logger()
	logger.Debug().Msg("finding task for worker")

	task := persistence.Task{}
	db := ts.db.GormDB()
	tx := db.Debug().
		Model(&task).
		Joins("left join jobs on tasks.job_id = jobs.id").
		Joins("left join task_dependencies on tasks.id = task_dependencies.task_id").
		Joins("left join tasks as tdeps on tdeps.id = task_dependencies.dependency_id").
		Where("tasks.status in ?", schedulableTaskStatuses).                       // Schedulable task statuses
		Where("tdeps.status in ? or tdeps.status is NULL", completedTaskStatuses). // Dependencies completed
		Where("jobs.status in ?", schedulableJobStatuses).                         // Schedulable job statuses
		// TODO: Supported task types
		// TODO: Non-blacklisted
		Order("jobs.priority desc"). // Highest job priority
		Order("priority desc").      // Highest task priority
		Limit(1).
		Preload("Job").
		First(&task)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			logger.Debug().Msg("no task for worker")
			return nil, nil
		}
		logger.Error().Err(tx.Error).Msg("error finding task for worker")
		return nil, tx.Error
	}

	return &task, nil
}
