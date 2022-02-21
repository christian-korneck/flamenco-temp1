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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	UUID string `gorm:"type:char(36);not null;unique;index"`

	Name     string `gorm:"type:varchar(64);not null"`
	JobType  string `gorm:"type:varchar(32);not null"`
	Priority int    `gorm:"type:smallint;not null"`
	Status   string `gorm:"type:varchar(32);not null"` // See JobStatusXxxx consts in openapi_types.gen.go

	Settings StringInterfaceMap `gorm:"type:jsonb"`
	Metadata StringStringMap    `gorm:"type:jsonb"`
}

type StringInterfaceMap map[string]interface{}
type StringStringMap map[string]string

type Task struct {
	gorm.Model
	UUID string `gorm:"type:char(36);not null;unique;index"`

	Name     string `gorm:"type:varchar(64);not null"`
	Type     string `gorm:"type:varchar(32);not null"`
	JobID    uint   `gorm:"not null"`
	Job      *Job   `gorm:"foreignkey:JobID;references:ID;constraint:OnDelete:CASCADE;not null"`
	Priority int    `gorm:"type:smallint;not null"`
	Status   string `gorm:"type:varchar(16);not null"`

	// Which worker is/was working on this.
	WorkerID *uint
	Worker   *Worker `gorm:"foreignkey:WorkerID;references:ID;constraint:OnDelete:CASCADE"`

	// Dependencies are tasks that need to be completed before this one can run.
	Dependencies []*Task `gorm:"many2many:task_dependencies;constraint:OnDelete:CASCADE"`

	Commands Commands `gorm:"type:jsonb"`
	Activity string   `gorm:"type:varchar(255);not null;default:\"\""`
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
	return db.gormDB.Transaction(func(tx *gorm.DB) error {
		// TODO: separate conversion of struct types from storing things in the database.
		dbJob := Job{
			UUID:     authoredJob.JobID,
			Name:     authoredJob.Name,
			JobType:  authoredJob.JobType,
			Status:   string(authoredJob.Status),
			Priority: authoredJob.Priority,
			Settings: StringInterfaceMap(authoredJob.Settings),
			Metadata: StringStringMap(authoredJob.Metadata),
		}

		if err := db.gormDB.Create(&dbJob).Error; err != nil {
			return fmt.Errorf("error storing job: %v", err)
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
				Status:   string(api.TaskStatusQueued),
				Commands: commands,
				// dependencies are stored below.
			}
			if err := db.gormDB.Create(&dbTask).Error; err != nil {
				return fmt.Errorf("error storing task: %v", err)
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
				return fmt.Errorf("unable to find task %q in the database, even though it was just authored", authoredTask.UUID)
			}

			deps := make([]*Task, len(authoredTask.Dependencies))
			for i, t := range authoredTask.Dependencies {
				depTask, ok := uuidToTask[t.UUID]
				if !ok {
					return fmt.Errorf("error finding task with UUID %q; a task depends on a task that is not part of this job", t.UUID)
				}
				deps[i] = depTask
			}

			dbTask.Dependencies = deps
			if err := db.gormDB.Save(dbTask).Error; err != nil {
				return fmt.Errorf("unable to store dependencies of task %q: %w", authoredTask.UUID, err)
			}
		}

		return nil
	})
}

func (db *DB) FetchJob(ctx context.Context, jobUUID string) (*Job, error) {
	dbJob := Job{}
	findResult := db.gormDB.First(&dbJob, "uuid = ?", jobUUID)
	if findResult.Error != nil {
		return nil, findResult.Error
	}

	return &dbJob, nil
}

func (db *DB) SaveJobStatus(ctx context.Context, j *Job) error {
	if err := db.gormDB.Model(j).Updates(Job{Status: j.Status}).Error; err != nil {
		return fmt.Errorf("error saving job status: %w", err)
	}
	return nil
}

func (db *DB) FetchTask(ctx context.Context, taskUUID string) (*Task, error) {
	dbTask := Task{}
	findResult := db.gormDB.First(&dbTask, "uuid = ?", taskUUID)
	if findResult.Error != nil {
		return nil, findResult.Error
	}

	return &dbTask, nil
}

func (db *DB) SaveTask(ctx context.Context, t *Task) error {
	if err := db.gormDB.Save(t).Error; err != nil {
		return fmt.Errorf("error saving task: %w", err)
	}
	return nil
}
