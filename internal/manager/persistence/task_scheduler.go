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
	// Note that active tasks are not schedulable, because they're already dunning on some worker.
	schedulableTaskStatuses = []api.TaskStatus{api.TaskStatusQueued, api.TaskStatusSoftFailed}
	schedulableJobStatuses  = []api.JobStatus{api.JobStatusActive, api.JobStatusQueued}
	// completedTaskStatuses   = []api.TaskStatus{api.TaskStatusCompleted}
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
			if isDatabaseBusyError(err) {
				logger.Trace().Err(err).Msg("database busy while finding task for worker")
				return errDatabaseBusy
			}
			logger.Error().Err(err).Msg("finding task for worker")
			return fmt.Errorf("finding task for worker: %w", err)
		}
		if task == nil {
			// No task found, which is fine.
			return nil
		}

		// Found a task, now assign it to the requesting worker.
		if err := assignTaskToWorker(tx, w, task); err != nil {
			if isDatabaseBusyError(err) {
				logger.Trace().Err(err).Msg("database busy while assigning task to worker")
				return errDatabaseBusy
			}

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

	// If a task is alreay active & assigned to this worker, return just that.
	// Note that this task type could be blocklisted or no longer supported by the
	// Worker, but since it's active that is unlikely.
	assignedTaskResult := taskAssignedAndRunnableQuery(tx.Model(&task), w).
		Preload("Job").
		Find(&task)
	if assignedTaskResult.Error != nil {
		return nil, assignedTaskResult.Error
	}
	if assignedTaskResult.RowsAffected > 0 {
		return &task, nil
	}

	// Produce the 'current task ID' by selecting all its incomplete dependencies.
	// This can then be used in a subquery to filter out such tasks.
	// `tasks.id` is the task ID from the outer query.
	incompleteDepsQuery := tx.Table("tasks as tasks2").
		Select("tasks2.id").
		Joins("left join task_dependencies td on tasks2.id = td.task_id").
		Joins("left join tasks dep on dep.id = td.dependency_id").
		Where("tasks2.id = tasks.id").
		Where("dep.status is not NULL and dep.status != ?", api.TaskStatusCompleted)

	blockedTaskTypesQuery := tx.Model(&JobBlock{}).
		Select("job_blocks.task_type").
		Where("job_blocks.worker_id = ?", w.ID).
		Where("job_blocks.job_id = jobs.id")

	// Note that this query doesn't check for the assigned worker. Tasks that have
	// a 'schedulable' status might have been assigned to a worker, representing
	// the last worker to touch it -- it's not meant to indicate "ownership" of
	// the task.
	findTaskResult := tx.
		Model(&task).
		Joins("left join jobs on tasks.job_id = jobs.id").
		Joins("left join task_failures TF on tasks.id = TF.task_id and TF.worker_id=?", w.ID).
		Where("tasks.status in ?", schedulableTaskStatuses).   // Schedulable task statuses
		Where("jobs.status in ?", schedulableJobStatuses).     // Schedulable job statuses
		Where("tasks.type in ?", w.TaskTypes()).               // Supported task types
		Where("tasks.id not in (?)", incompleteDepsQuery).     // Dependencies completed
		Where("TF.worker_id is NULL").                         // Not failed before
		Where("tasks.type not in (?)", blockedTaskTypesQuery). // Non-blocklisted
		Order("jobs.priority desc").                           // Highest job priority
		Order("tasks.priority desc").                          // Highest task priority
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
	return tx.Model(t).
		Select("WorkerID", "LastTouchedAt").
		Updates(Task{WorkerID: &w.ID, LastTouchedAt: tx.NowFunc()}).Error
}

// taskAssignedAndRunnableQuery appends some GORM clauses to query for a task
// that's already assigned to this worker, and is in a runnable state.
func taskAssignedAndRunnableQuery(tx *gorm.DB, w *Worker) *gorm.DB {
	return tx.
		Joins("left join jobs on tasks.job_id = jobs.id").
		Where("tasks.status = ?", api.TaskStatusActive).
		Where("jobs.status in ?", schedulableJobStatuses).
		Where("tasks.worker_id = ?", w.ID). // assigned to this worker
		Limit(1)
}
