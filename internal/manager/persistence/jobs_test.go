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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	"golang.org/x/net/context"
)

func TestStoreAuthoredJob(t *testing.T) {
	db := createTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	task1 := job_compilers.AuthoredTask{
		Name: "render-1-3",
		Type: "blender",
		Commands: []job_compilers.AuthoredCommand{
			{
				Type: "blender-render",
				Parameters: job_compilers.AuthoredCommandParameters{
					"exe":       "{blender}",
					"blendfile": "/path/to/file.blend",
					"args": []interface{}{
						"--render-output", "/path/to/output/######.png",
						"--render-format", "PNG",
						"--render-frame", "1-3",
					},
				}},
		},
	}

	task2 := task1
	task2.Name = "render-4-6"
	task2.Commands[0].Parameters["frames"] = "4-6"

	task3 := job_compilers.AuthoredTask{
		Name: "preview-video",
		Type: "ffmpeg",
		Commands: []job_compilers.AuthoredCommand{
			{
				Type: "merge-frames-to-video",
				Parameters: job_compilers.AuthoredCommandParameters{
					"images":       "/path/to/output/######.png",
					"output":       "/path/to/output/preview.mkv",
					"ffmpegParams": "-c:v hevc -crf 31",
				}},
		},
		Dependencies: []*job_compilers.AuthoredTask{&task1, &task2},
	}

	job := job_compilers.AuthoredJob{
		JobID:    "263fd47e-b9f8-4637-b726-fd7e47ecfdae",
		Name:     "Test job",
		Priority: 50,
		Settings: job_compilers.JobSettings{
			"frames":     "1-6",
			"chunk_size": 3.0, // The roundtrip to JSON in PostgreSQL can make this a float.
		},
		Metadata: job_compilers.JobMetadata{
			"author":  "Sybren",
			"project": "Sprite Fright",
		},
		Tasks: []job_compilers.AuthoredTask{task1, task2, task3},
	}

	err := db.StoreJob(ctx, job)
	assert.NoError(t, err)

	fetchedJob, err := db.FetchJob(ctx, job.JobID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedJob)

	// Test contents of fetched job
	assert.Equal(t, job.JobID, fetchedJob.Id)
	assert.Equal(t, job.Name, fetchedJob.Name)
	assert.Equal(t, job.JobType, fetchedJob.Type)
	assert.Equal(t, job.Priority, fetchedJob.Priority)
	assert.EqualValues(t, map[string]interface{}(job.Settings), fetchedJob.Settings.AdditionalProperties)
	assert.EqualValues(t, map[string]string(job.Metadata), fetchedJob.Metadata.AdditionalProperties)

	// Fetch tasks of job.
	var dbJob Job
	tx := db.gormDB.Where(&Job{UUID: job.JobID}).Find(&dbJob)
	assert.NoError(t, tx.Error)
	var tasks []Task
	tx = db.gormDB.Where("job_id = ?", dbJob.ID).Find(&tasks)
	assert.NoError(t, tx.Error)

	assert.Len(t, tasks, 3)

	// TODO: test task contents.
}