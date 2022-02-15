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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
	"golang.org/x/net/context"
)

func TestCreateFetchWorker(t *testing.T) {
	db := CreateTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	w := Worker{
		UUID:               uuid.New().String(),
		Name:               "дрон",
		Address:            "fe80::5054:ff:fede:2ad7",
		LastActivity:       "",
		Platform:           "linux",
		Software:           "3.0",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management",
	}

	err := db.CreateWorker(ctx, &w)
	assert.NoError(t, err)

	fetchedWorker, err := db.FetchWorker(ctx, w.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedWorker)

	// Test contents of fetched job
	assert.Equal(t, w.UUID, fetchedWorker.UUID)
	assert.Equal(t, w.Name, fetchedWorker.Name)
	assert.Equal(t, w.Address, fetchedWorker.Address)
	assert.Equal(t, w.LastActivity, fetchedWorker.LastActivity)
	assert.Equal(t, w.Platform, fetchedWorker.Platform)
	assert.Equal(t, w.Software, fetchedWorker.Software)
	assert.Equal(t, w.Status, fetchedWorker.Status)

	assert.EqualValues(t, w.SupportedTaskTypes, fetchedWorker.SupportedTaskTypes)
}

func TestSaveWorker(t *testing.T) {
	db := CreateTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	w := Worker{
		UUID:               uuid.New().String(),
		Name:               "дрон",
		Address:            "fe80::5054:ff:fede:2ad7",
		LastActivity:       "",
		Platform:           "linux",
		Software:           "3.0",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management",
	}

	err := db.CreateWorker(ctx, &w)
	assert.NoError(t, err)

	fetchedWorker, err := db.FetchWorker(ctx, w.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedWorker)

	// Update all updatable fields of the Worker
	updatedWorker := *fetchedWorker
	updatedWorker.Name = "7 မှ 9"
	updatedWorker.Address = "fe80::cafe:f00d"
	updatedWorker.LastActivity = "Rendering"
	updatedWorker.Platform = "windows"
	updatedWorker.Software = "3.1"
	updatedWorker.Status = api.WorkerStatusAsleep
	updatedWorker.SupportedTaskTypes = "blender,ffmpeg,file-management,misc"

	// Saving only the status should just do that.
	err = db.SaveWorkerStatus(ctx, &updatedWorker)
	assert.NoError(t, err)
	assert.Equal(t, "7 မှ 9", updatedWorker.Name, "Saving status should not touch the name")

	// Check saved worker
	fetchedWorker, err = db.FetchWorker(ctx, w.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedWorker)
	assert.Equal(t, updatedWorker.Status, fetchedWorker.Status, "new status should have been saved")
	assert.NotEqual(t, updatedWorker.Name, fetchedWorker.Name, "non-status fields should not have been updated")

	// Saving the entire worker should save everything.
	err = db.SaveWorker(ctx, &updatedWorker)
	assert.NoError(t, err)

	// Check saved worker
	fetchedWorker, err = db.FetchWorker(ctx, w.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedWorker)
	assert.Equal(t, updatedWorker.Status, fetchedWorker.Status, "new status should have been saved")
	assert.Equal(t, updatedWorker.Name, fetchedWorker.Name, "non-status fields should also have been updated")
	assert.Equal(t, updatedWorker.Software, fetchedWorker.Software, "non-status fields should also have been updated")
}
