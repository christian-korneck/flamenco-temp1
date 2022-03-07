// Package persistence provides the database interface for Flamenco Manager.
package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/pkg/api"
)

func TestCreateFetchWorker(t *testing.T) {
	ctx, cancel, db := persistenceTestFixtures(t, 1*time.Second)
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
	ctx, cancel, db := persistenceTestFixtures(t, 1*time.Second)
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
