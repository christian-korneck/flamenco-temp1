package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"git.blender.org/flamenco/internal/manager/job_compilers"
	"git.blender.org/flamenco/pkg/api"
)

type Job struct {
	Model
	UUID string `gorm:"type:char(36);default:'';unique;index"`

	Name     string        `gorm:"type:varchar(64);default:''"`
	JobType  string        `gorm:"type:varchar(32);default:''"`
	Priority int           `gorm:"type:smallint;default:0"`
	Status   api.JobStatus `gorm:"type:varchar(32);default:''"`
	Activity string        `gorm:"type:varchar(255);default:''"`

	Settings StringInterfaceMap `gorm:"type:jsonb"`
	Metadata StringStringMap    `gorm:"type:jsonb"`
}

type StringInterfaceMap map[string]interface{}
type StringStringMap map[string]string

type Task struct {
	Model
	UUID string `gorm:"type:char(36);default:'';unique;index"`

	Name     string         `gorm:"type:varchar(64);default:''"`
	Type     string         `gorm:"type:varchar(32);default:''"`
	JobID    uint           `gorm:"default:0"`
	Job      *Job           `gorm:"foreignkey:JobID;references:ID;constraint:OnDelete:CASCADE"`
	Priority int            `gorm:"type:smallint;default:50"`
	Status   api.TaskStatus `gorm:"type:varchar(16);default:''"`

	// Which worker is/was working on this.
	WorkerID      *uint
	Worker        *Worker   `gorm:"foreignkey:WorkerID;references:ID;constraint:OnDelete:CASCADE"`
	LastTouchedAt time.Time `gorm:"index"` // Should contain UTC timestamps.

	// Dependencies are tasks that need to be completed before this one can run.
	Dependencies []*Task `gorm:"many2many:task_dependencies;constraint:OnDelete:CASCADE"`

	Commands Commands `gorm:"type:jsonb"`
	Activity string   `gorm:"type:varchar(255);default:''"`
}

type Commands []Command

type Command struct {
	Name       string             `json:"name"`
	Parameters StringInterfaceMap `json:"parameters"`
}

func (c Commands) Value() (driver.Value, error) {
	return json.Marshal(c)
}
func (c *Commands) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &c)
}

func (js StringInterfaceMap) Value() (driver.Value, error) {
	return json.Marshal(js)
}
func (js *StringInterfaceMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &js)
}

func (js StringStringMap) Value() (driver.Value, error) {
	return json.Marshal(js)
}
func (js *StringStringMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &js)
}

// StoreJob stores an AuthoredJob and its tasks, and saves it to the database.
// The job will be in 'under construction' status. It is up to the caller to transition it to its desired initial status.
func (db *DB) StoreAuthoredJob(ctx context.Context, authoredJob job_compilers.AuthoredJob) error {
	return db.gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// TODO: separate conversion of struct types from storing things in the database.
		dbJob := Job{
			UUID:     authoredJob.JobID,
			Name:     authoredJob.Name,
			JobType:  authoredJob.JobType,
			Status:   authoredJob.Status,
			Priority: authoredJob.Priority,
			Settings: StringInterfaceMap(authoredJob.Settings),
			Metadata: StringStringMap(authoredJob.Metadata),
		}

		if err := tx.Create(&dbJob).Error; err != nil {
			return jobError(err, "storing job")
		}

		uuidToTask := make(map[string]*Task)
		for _, authoredTask := range authoredJob.Tasks {
			var commands []Command
			for _, authoredCommand := range authoredTask.Commands {
				commands = append(commands, Command{
					Name:       authoredCommand.Name,
					Parameters: StringInterfaceMap(authoredCommand.Parameters),
				})
			}

			dbTask := Task{
				Name:     authoredTask.Name,
				Type:     authoredTask.Type,
				UUID:     authoredTask.UUID,
				Job:      &dbJob,
				Priority: authoredTask.Priority,
				Status:   api.TaskStatusQueued,
				Commands: commands,
				// dependencies are stored below.
			}
			if err := tx.Create(&dbTask).Error; err != nil {
				return taskError(err, "storing task: %v", err)
			}

			uuidToTask[authoredTask.UUID] = &dbTask
		}

		// Store the dependencies between tasks.
		for _, authoredTask := range authoredJob.Tasks {
			if len(authoredTask.Dependencies) == 0 {
				continue
			}

			dbTask, ok := uuidToTask[authoredTask.UUID]
			if !ok {
				return taskError(nil, "unable to find task %q in the database, even though it was just authored", authoredTask.UUID)
			}

			deps := make([]*Task, len(authoredTask.Dependencies))
			for i, t := range authoredTask.Dependencies {
				depTask, ok := uuidToTask[t.UUID]
				if !ok {
					return taskError(nil, "finding task with UUID %q; a task depends on a task that is not part of this job", t.UUID)
				}
				deps[i] = depTask
			}

			dbTask.Dependencies = deps
			subQuery := tx.Model(dbTask).Updates(Task{Dependencies: deps})
			if subQuery.Error != nil {
				return taskError(subQuery.Error, "unable to store dependencies of task %q", authoredTask.UUID)
			}
		}

		return nil
	})
}

// FetchJob fetches a single job, without fetching its tasks.
func (db *DB) FetchJob(ctx context.Context, jobUUID string) (*Job, error) {
	dbJob := Job{}
	findResult := db.gormDB.WithContext(ctx).First(&dbJob, "uuid = ?", jobUUID)
	if findResult.Error != nil {
		return nil, jobError(findResult.Error, "fetching job")
	}

	return &dbJob, nil
}

// DeleteJob deletes a job from the database.
// The deletion cascades to its tasks and other job-related tables.
func (db *DB) DeleteJob(ctx context.Context, jobUUID string) error {
	tx := db.gormDB.WithContext(ctx).
		Where("uuid = ?", jobUUID).
		Delete(&Job{})
	if tx.Error != nil {
		return jobError(tx.Error, "deleting job")
	}
	return nil
}

func (db *DB) FetchJobsInStatus(ctx context.Context, jobStatuses ...api.JobStatus) ([]*Job, error) {
	var jobs []*Job

	tx := db.gormDB.WithContext(ctx).
		Model(&Job{}).
		Where("status in ?", jobStatuses).
		Scan(&jobs)

	if tx.Error != nil {
		return nil, jobError(tx.Error, "fetching jobs in status %q", jobStatuses)
	}
	return jobs, nil
}

// SaveJobStatus saves the job's Status and Activity fields.
func (db *DB) SaveJobStatus(ctx context.Context, j *Job) error {
	tx := db.gormDB.WithContext(ctx).
		Model(j).
		Updates(Job{Status: j.Status, Activity: j.Activity})
	if tx.Error != nil {
		return jobError(tx.Error, "saving job status")
	}
	return nil
}

func (db *DB) FetchTask(ctx context.Context, taskUUID string) (*Task, error) {
	dbTask := Task{}
	tx := db.gormDB.WithContext(ctx).
		Joins("Job").
		First(&dbTask, "tasks.uuid = ?", taskUUID)
	if tx.Error != nil {
		return nil, taskError(tx.Error, "fetching task")
	}
	return &dbTask, nil
}

func (db *DB) SaveTask(ctx context.Context, t *Task) error {
	if err := db.gormDB.WithContext(ctx).Save(t).Error; err != nil {
		return taskError(err, "saving task")
	}
	return nil
}

func (db *DB) SaveTaskActivity(ctx context.Context, t *Task) error {
	if err := db.gormDB.Model(t).Updates(Task{Activity: t.Activity}).Error; err != nil {
		return taskError(err, "saving task activity")
	}
	return nil
}

func (db *DB) TaskAssignToWorker(ctx context.Context, t *Task, w *Worker) error {
	tx := db.gormDB.WithContext(ctx).
		Model(t).Updates(Task{WorkerID: &w.ID})
	if tx.Error != nil {
		return taskError(tx.Error, "assigning task %s to worker %s", t.UUID, w.UUID)
	}

	// Gorm updates t.WorkerID itself, but not t.Worker (even when it's added to
	// the Updates() call above).
	t.Worker = w

	return nil
}

func (db *DB) FetchTasksOfWorkerInStatus(ctx context.Context, worker *Worker, taskStatus api.TaskStatus) ([]*Task, error) {
	result := []*Task{}
	tx := db.gormDB.WithContext(ctx).
		Model(&Task{}).
		Joins("Job").
		Where("tasks.worker_id = ?", worker.ID).
		Where("tasks.status = ?", taskStatus).
		Scan(&result)
	if tx.Error != nil {
		return nil, taskError(tx.Error, "finding tasks of worker %s in status %q", worker.UUID, taskStatus)
	}
	return result, nil
}

func (db *DB) JobHasTasksInStatus(ctx context.Context, job *Job, taskStatus api.TaskStatus) (bool, error) {
	var numTasksInStatus int64
	tx := db.gormDB.WithContext(ctx).
		Model(&Task{}).
		Where("job_id", job.ID).
		Where("status", taskStatus).
		Count(&numTasksInStatus)
	if tx.Error != nil {
		return false, taskError(tx.Error, "counting tasks of job %s in status %q", job.UUID, taskStatus)
	}
	return numTasksInStatus > 0, nil
}

func (db *DB) CountTasksOfJobInStatus(
	ctx context.Context,
	job *Job,
	taskStatuses ...api.TaskStatus,
) (numInStatus, numTotal int, err error) {
	type Result struct {
		Status   api.TaskStatus
		NumTasks int
	}
	var results []Result

	tx := db.gormDB.WithContext(ctx).
		Model(&Task{}).
		Select("status, count(*) as num_tasks").
		Where("job_id", job.ID).
		Group("status").
		Scan(&results)

	if tx.Error != nil {
		return 0, 0, jobError(tx.Error, "count tasks of job %s in status %q", job.UUID, taskStatuses)
	}

	// Create lookup table for which statuses to count.
	countStatus := map[api.TaskStatus]bool{}
	for _, status := range taskStatuses {
		countStatus[status] = true
	}

	// Count the number of tasks per status.
	for _, result := range results {
		if countStatus[result.Status] {
			numInStatus += result.NumTasks
		}
		numTotal += result.NumTasks
	}

	return
}

// FetchTaskIDsOfJob returns all tasks of the given job.
func (db *DB) FetchTasksOfJob(ctx context.Context, job *Job) ([]*Task, error) {
	var tasks []*Task
	tx := db.gormDB.WithContext(ctx).
		Model(&Task{}).
		Where("job_id", job.ID).
		Scan(&tasks)
	if tx.Error != nil {
		return nil, taskError(tx.Error, "fetching tasks of job %s", job.UUID)
	}

	for i := range tasks {
		tasks[i].Job = job
	}

	return tasks, nil
}

// FetchTasksOfJobInStatus returns those tasks of the given job that have any of the given statuses.
func (db *DB) FetchTasksOfJobInStatus(ctx context.Context, job *Job, taskStatuses ...api.TaskStatus) ([]*Task, error) {
	var tasks []*Task
	tx := db.gormDB.WithContext(ctx).
		Model(&Task{}).
		Where("job_id", job.ID).
		Where("status in ?", taskStatuses).
		Scan(&tasks)
	if tx.Error != nil {
		return nil, taskError(tx.Error, "fetching tasks of job %s in status %q", job.UUID, taskStatuses)
	}

	for i := range tasks {
		tasks[i].Job = job
	}

	return tasks, nil
}

// UpdateJobsTaskStatuses updates the status & activity of all tasks of `job`.
func (db *DB) UpdateJobsTaskStatuses(ctx context.Context, job *Job,
	taskStatus api.TaskStatus, activity string) error {

	if taskStatus == "" {
		return taskError(nil, "empty status not allowed")
	}

	tx := db.gormDB.WithContext(ctx).
		Model(Task{}).
		Where("job_Id = ?", job.ID).
		Updates(Task{Status: taskStatus, Activity: activity})

	if tx.Error != nil {
		return taskError(tx.Error, "updating status of all tasks of job %s", job.UUID)
	}
	return nil
}

// UpdateJobsTaskStatusesConditional updates the status & activity of the tasks of `job`,
// limited to those tasks with status in `statusesToUpdate`.
func (db *DB) UpdateJobsTaskStatusesConditional(ctx context.Context, job *Job,
	statusesToUpdate []api.TaskStatus, taskStatus api.TaskStatus, activity string) error {

	if taskStatus == "" {
		return taskError(nil, "empty status not allowed")
	}

	tx := db.gormDB.WithContext(ctx).
		Model(Task{}).
		Where("job_Id = ?", job.ID).
		Where("status in ?", statusesToUpdate).
		Updates(Task{Status: taskStatus, Activity: activity})
	if tx.Error != nil {
		return taskError(tx.Error, "updating status of all tasks in status %v of job %s", statusesToUpdate, job.UUID)
	}
	return nil
}

// TaskTouchedByWorker marks the task as 'touched' by a worker. This is used for timeout detection.
func (db *DB) TaskTouchedByWorker(ctx context.Context, t *Task) error {
	tx := db.gormDB.WithContext(ctx).
		Model(t).
		Updates(Task{LastTouchedAt: db.gormDB.NowFunc()})
	if err := tx.Error; err != nil {
		return taskError(err, "saving task 'last touched at'")
	}
	return nil
}
