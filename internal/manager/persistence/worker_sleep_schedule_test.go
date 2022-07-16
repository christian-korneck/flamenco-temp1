package persistence

import (
	"testing"
	"time"

	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestFetchWorkerSleepSchedule(t *testing.T) {
	ctx, finish, db := persistenceTestFixtures(t, 1*time.Second)
	defer finish()

	linuxWorker := Worker{
		UUID:               uuid.New(),
		Name:               "дрон",
		Address:            "fe80::5054:ff:fede:2ad7",
		Platform:           "linux",
		Software:           "3.0",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management",
	}
	err := db.CreateWorker(ctx, &linuxWorker)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Not an existing Worker.
	fetched, err := db.FetchWorkerSleepSchedule(ctx, "2cf6153a-3d4e-49f4-a5c0-1c9fc176e155")
	assert.NoError(t, err, "non-existent worker should not cause an error")
	assert.Nil(t, fetched)

	// No sleep schedule.
	fetched, err = db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	assert.NoError(t, err, "non-existent schedule should not cause an error")
	assert.Nil(t, fetched)

	// Create a sleep schedule.
	created := SleepSchedule{
		WorkerID: linuxWorker.ID,
		Worker:   &linuxWorker,

		IsActive:   true,
		DaysOfWeek: "mo,tu,th,fr",
		StartTime:  "18:00",
		EndTime:    "09:00",
	}
	tx := db.gormDB.Create(&created)
	if !assert.NoError(t, tx.Error) {
		t.FailNow()
	}

	fetched, err = db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	assert.NoError(t, err)
	assertEqualSleepSchedule(t, linuxWorker.ID, created, *fetched)
}

func TestSetWorkerSleepSchedule(t *testing.T) {
	ctx, finish, db := persistenceTestFixtures(t, 1*time.Second)
	defer finish()

	linuxWorker := Worker{
		UUID:               uuid.New(),
		Name:               "дрон",
		Address:            "fe80::5054:ff:fede:2ad7",
		Platform:           "linux",
		Software:           "3.0",
		Status:             api.WorkerStatusAwake,
		SupportedTaskTypes: "blender,ffmpeg,file-management",
	}
	err := db.CreateWorker(ctx, &linuxWorker)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	schedule := SleepSchedule{
		WorkerID: linuxWorker.ID,
		Worker:   &linuxWorker,

		IsActive:   true,
		DaysOfWeek: "mo,tu,th,fr",
		StartTime:  "18:00",
		EndTime:    "09:00",
	}

	// Not an existing Worker.
	err = db.SetWorkerSleepSchedule(ctx, "2cf6153a-3d4e-49f4-a5c0-1c9fc176e155", schedule)
	assert.ErrorIs(t, err, ErrWorkerNotFound)

	// Create the sleep schedule.
	err = db.SetWorkerSleepSchedule(ctx, linuxWorker.UUID, schedule)
	assert.NoError(t, err)
	fetched, err := db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	assert.NoError(t, err)
	assertEqualSleepSchedule(t, linuxWorker.ID, schedule, *fetched)

	// Overwrite the schedule with one that already has a database ID.
	newSchedule := schedule
	newSchedule.IsActive = false
	newSchedule.DaysOfWeek = "mo,tu,we,th,fr"
	newSchedule.StartTime = "02:00"
	newSchedule.EndTime = "06:00"
	err = db.SetWorkerSleepSchedule(ctx, linuxWorker.UUID, newSchedule)
	assert.NoError(t, err)
	fetched, err = db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	assert.NoError(t, err)
	assertEqualSleepSchedule(t, linuxWorker.ID, newSchedule, *fetched)

	// Overwrite the schedule with a freshly constructed one.
	newerSchedule := SleepSchedule{
		WorkerID: linuxWorker.ID,
		Worker:   &linuxWorker,

		IsActive:   true,
		DaysOfWeek: "mo",
		StartTime:  "03:27",
		EndTime:    "15:47",
	}
	err = db.SetWorkerSleepSchedule(ctx, linuxWorker.UUID, newerSchedule)
	assert.NoError(t, err)
	fetched, err = db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	assert.NoError(t, err)
	assertEqualSleepSchedule(t, linuxWorker.ID, newerSchedule, *fetched)
}

func assertEqualSleepSchedule(t *testing.T, workerID uint, expect, actual SleepSchedule) {
	assert.Equal(t, workerID, actual.WorkerID)
	assert.Nil(t, actual.Worker, "the Worker itself should not be fetched")
	assert.Equal(t, expect.IsActive, actual.IsActive)
	assert.Equal(t, expect.DaysOfWeek, actual.DaysOfWeek)
	assert.Equal(t, expect.StartTime, actual.StartTime)
	assert.Equal(t, expect.EndTime, actual.EndTime)
}
