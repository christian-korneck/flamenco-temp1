// Package persistence provides the database interface for Flamenco Manager.
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	"gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

// TODO : have this configurable from the CLI.
const dbDSN = "host=localhost user=flamenco password=flamenco dbname=flamenco TimeZone=Europe/Amsterdam"

// DB provides the database interface.
type DB struct {
	gormDB *gorm.DB
}

func OpenDB(ctx context.Context) (*DB, error) {
	db, err := openDB(ctx, dbDSN)
	if err != nil {
		return nil, err
	}

	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

func openDB(ctx context.Context, uri string) (*DB, error) {
	// TODO: don't log the password.
	log.Info().Str("dsn", uri).Msg("opening database")

	gormDB, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		log.Panic().Err(err).Msg("failed to connect database")
	}

	db := DB{
		gormDB: gormDB,
	}
	return &db, nil
}

func (db *DB) StoreJob(ctx context.Context, authoredJob job_compilers.AuthoredJob) error {
	return db.gormDB.Transaction(func(tx *gorm.DB) error {
		// TODO: separate conversion of struct types from storing things in the database.
		dbJob := Job{
			UUID:     authoredJob.JobID,
			Name:     authoredJob.Name,
			JobType:  authoredJob.JobType,
			Priority: int8(authoredJob.Priority),
			Settings: StringInterfaceMap(authoredJob.Settings),
			Metadata: StringStringMap(authoredJob.Metadata),
		}

		if err := db.gormDB.Create(&dbJob).Error; err != nil {
			return fmt.Errorf("error storing job: %v", err)
		}

		for _, authoredTask := range authoredJob.Tasks {
			var commands []Command
			for _, authoredCommand := range authoredTask.Commands {
				commands = append(commands, Command{
					Type:       authoredCommand.Type,
					Parameters: StringInterfaceMap(authoredCommand.Parameters),
				})
			}

			dbTask := Task{
				Name:     authoredTask.Name,
				Type:     authoredTask.Type,
				Job:      &dbJob,
				Priority: authoredTask.Priority,
				Status:   string(api.TaskStatusProcessing), // TODO: is this the right place to set the default status?
				// TODO: store dependencies
				Commands: commands,
			}
			if err := db.gormDB.Create(&dbTask).Error; err != nil {
				return fmt.Errorf("error storing task: %v", err)
			}
		}

		return nil
	})
}

func (db *DB) FetchJob(ctx context.Context, jobID string) (*api.Job, error) {
	dbJob := Job{}
	findResult := db.gormDB.First(&dbJob, "uuid = ?", jobID)
	if findResult.Error != nil {
		return nil, findResult.Error
	}

	apiJob := api.Job{
		SubmittedJob: api.SubmittedJob{
			Name:     dbJob.Name,
			Priority: int(dbJob.Priority),
			Type:     dbJob.JobType,
		},

		Id:      dbJob.UUID,
		Created: dbJob.CreatedAt,
		Updated: dbJob.UpdatedAt,
		Status:  api.JobStatus(dbJob.Status),
	}

	apiJob.Settings = &api.JobSettings{AdditionalProperties: dbJob.Settings}
	apiJob.Metadata = &api.JobMetadata{AdditionalProperties: dbJob.Metadata}

	return &apiJob, nil
}
