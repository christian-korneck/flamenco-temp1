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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/job_compilers"
	"golang.org/x/net/context"
	_ "modernc.org/sqlite"
)

const testURI = "testing.sqlite"

func createTestDB(t *testing.T) (*DB, func()) {
	// Creating a new database should be fast.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	db, err := openDB(ctx, testURI)
	assert.Nil(t, err)

	return db, func() {
		os.Remove(testURI)
	}
}

func TestStoreAuthoredJob(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := db.StoreJob(ctx, job_compilers.AuthoredJob{
		JobID:    "263fd47e-b9f8-4637-b726-fd7e47ecfdae",
		Name:     "Test job",
		Priority: 50,
		Settings: job_compilers.JobSettings{
			"frames":     "1-20",
			"chunk_size": 3,
		},
		Metadata: job_compilers.JobMetadata{
			"author":  "Sybren",
			"project": "Sprite Fright",
		},
	})

	assert.Nil(t, err)

	// TODO: fetch the job to see it was stored well.
}
