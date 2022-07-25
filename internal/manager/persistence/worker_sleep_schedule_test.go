package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

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
		StartTime:  TimeOfDay{18, 0},
		EndTime:    TimeOfDay{9, 0},
	}
	tx := db.gormDB.Create(&created)
	if !assert.NoError(t, tx.Error) {
		t.FailNow()
	}

	fetched, err = db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	assert.NoError(t, err)
	assertEqualSleepSchedule(t, linuxWorker.ID, created, *fetched)
}

func TestFetchSleepScheduleWorker(t *testing.T) {
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

	// Create a sleep schedule.
	created := SleepSchedule{
		WorkerID: linuxWorker.ID,
		Worker:   &linuxWorker,

		IsActive:   true,
		DaysOfWeek: "mo,tu,th,fr",
		StartTime:  TimeOfDay{18, 0},
		EndTime:    TimeOfDay{9, 0},
	}
	tx := db.gormDB.Create(&created)
	if !assert.NoError(t, tx.Error) {
		t.FailNow()
	}

	dbSchedule, err := db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	assert.NoError(t, err)
	assert.Nil(t, dbSchedule.Worker, "worker should be nil when fetching schedule")

	err = db.FetchSleepScheduleWorker(ctx, dbSchedule)
	assert.NoError(t, err)
	if assert.NotNil(t, dbSchedule.Worker) {
		// Compare a few fields. If these are good, the correct worker has been fetched.
		assert.Equal(t, linuxWorker.ID, dbSchedule.Worker.ID)
		assert.Equal(t, linuxWorker.UUID, dbSchedule.Worker.UUID)
	}
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
		StartTime:  TimeOfDay{18, 0},
		EndTime:    TimeOfDay{9, 0},
	}

	// Not an existing Worker.
	err = db.SetWorkerSleepSchedule(ctx, "2cf6153a-3d4e-49f4-a5c0-1c9fc176e155", &schedule)
	assert.ErrorIs(t, err, ErrWorkerNotFound)

	// Create the sleep schedule.
	err = db.SetWorkerSleepSchedule(ctx, linuxWorker.UUID, &schedule)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	fetched, err := db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assertEqualSleepSchedule(t, linuxWorker.ID, schedule, *fetched)

	// Overwrite the schedule with one that already has a database ID.
	newSchedule := schedule
	newSchedule.IsActive = false
	newSchedule.DaysOfWeek = "mo,tu,we,th,fr"
	newSchedule.StartTime = TimeOfDay{2, 0}
	newSchedule.EndTime = TimeOfDay{6, 0}
	err = db.SetWorkerSleepSchedule(ctx, linuxWorker.UUID, &newSchedule)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	fetched, err = db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assertEqualSleepSchedule(t, linuxWorker.ID, newSchedule, *fetched)

	// Overwrite the schedule with a freshly constructed one.
	newerSchedule := SleepSchedule{
		WorkerID: linuxWorker.ID,
		Worker:   &linuxWorker,

		IsActive:   true,
		DaysOfWeek: "mo",
		StartTime:  TimeOfDay{3, 0},
		EndTime:    TimeOfDay{15, 0},
	}
	err = db.SetWorkerSleepSchedule(ctx, linuxWorker.UUID, &newerSchedule)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	fetched, err = db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assertEqualSleepSchedule(t, linuxWorker.ID, newerSchedule, *fetched)

	// Clear the sleep schedule.
	emptySchedule := SleepSchedule{
		WorkerID: linuxWorker.ID,
		Worker:   &linuxWorker,

		IsActive:   false,
		DaysOfWeek: "",
		StartTime:  emptyToD,
		EndTime:    emptyToD,
	}
	err = db.SetWorkerSleepSchedule(ctx, linuxWorker.UUID, &emptySchedule)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	fetched, err = db.FetchWorkerSleepSchedule(ctx, linuxWorker.UUID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assertEqualSleepSchedule(t, linuxWorker.ID, emptySchedule, *fetched)

}

func TestSetWorkerSleepScheduleNextCheck(t *testing.T) {
	ctx, finish, db := persistenceTestFixtures(t, 1*time.Second)
	defer finish()

	schedule := SleepSchedule{
		Worker: &Worker{
			UUID:   "2b1f857a-fd64-484b-9c17-cf89bbe47be7",
			Name:   "дрон 1",
			Status: api.WorkerStatusAwake,
		},
		IsActive:   true,
		DaysOfWeek: "mo,tu,th,fr",
		StartTime:  TimeOfDay{18, 0},
		EndTime:    TimeOfDay{9, 0},
	}
	// Use GORM to create the worker and sleep schedule in one go.
	if tx := db.gormDB.Create(&schedule); tx.Error != nil {
		panic(tx.Error)
	}

	future := db.gormDB.NowFunc().Add(5 * time.Hour)
	schedule.NextCheck = future

	err := db.SetWorkerSleepScheduleNextCheck(ctx, &schedule)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	fetched, err := db.FetchWorkerSleepSchedule(ctx, schedule.Worker.UUID)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assertEqualSleepSchedule(t, schedule.Worker.ID, schedule, *fetched)
}

func TestFetchSleepSchedulesToCheck(t *testing.T) {
	ctx, finish, db := persistenceTestFixtures(t, 1*time.Second)
	defer finish()

	mockedNow := mustParseTime("2022-06-07T11:14:47+02:00").UTC()
	mockedPast := mockedNow.Add(-10 * time.Second)
	mockedFuture := mockedNow.Add(10 * time.Second)

	db.gormDB.NowFunc = func() time.Time { return mockedNow }

	schedule0 := SleepSchedule{ // Next check in the past -> should be checked.
		Worker: &Worker{
			UUID:   "2b1f857a-fd64-484b-9c17-cf89bbe47be7",
			Name:   "дрон 1",
			Status: api.WorkerStatusAwake,
		},
		IsActive:   true,
		DaysOfWeek: "mo,tu,th,fr",
		StartTime:  TimeOfDay{18, 0},
		EndTime:    TimeOfDay{9, 0},

		NextCheck: mockedPast,
	}

	schedule1 := SleepSchedule{ // Next check in future -> should not be checked.
		Worker: &Worker{
			UUID:   "4475738e-41eb-47b2-8bca-2bbcabab69bb",
			Name:   "дрон 2",
			Status: api.WorkerStatusAwake,
		},
		IsActive:   true,
		DaysOfWeek: "mo,tu,th,fr",
		StartTime:  TimeOfDay{18, 0},
		EndTime:    TimeOfDay{9, 0},

		NextCheck: mockedFuture,
	}

	schedule2 := SleepSchedule{ // Next check is zero value -> should be checked.
		Worker: &Worker{
			UUID:   "dc251817-6a11-4548-a36a-07b0d50b4c21",
			Name:   "дрон 3",
			Status: api.WorkerStatusAwake,
		},
		IsActive:   true,
		DaysOfWeek: "mo,tu,th,fr",
		StartTime:  TimeOfDay{18, 0},
		EndTime:    TimeOfDay{9, 0},

		NextCheck: time.Time{}, // zero value for time.
	}

	schedule3 := SleepSchedule{ // Schedule inactive -> should not be checked.
		Worker: &Worker{
			UUID:   "874d5fc6-5784-4d43-8c20-6e7e73fc1b8d",
			Name:   "дрон 4",
			Status: api.WorkerStatusAwake,
		},
		IsActive:   false,
		DaysOfWeek: "mo,tu,th,fr",
		StartTime:  TimeOfDay{18, 0},
		EndTime:    TimeOfDay{9, 0},

		NextCheck: mockedPast, // next check in the past, so if active it would be checked.
	}

	// Use GORM to create the workers and sleep schedules in one go.
	scheds := []*SleepSchedule{&schedule0, &schedule1, &schedule2, &schedule3}
	for idx := range scheds {
		if tx := db.gormDB.Create(scheds[idx]); tx.Error != nil {
			panic(tx.Error)
		}
	}

	toCheck, err := db.FetchSleepSchedulesToCheck(ctx)
	if assert.NoError(t, err) && assert.Len(t, toCheck, 2) {
		assertEqualSleepSchedule(t, schedule0.Worker.ID, schedule0, *toCheck[0])
		assert.Nil(t, toCheck[0].Worker, "the Worker should NOT be fetched")
		assertEqualSleepSchedule(t, schedule2.Worker.ID, schedule1, *toCheck[1])
		assert.Nil(t, toCheck[1].Worker, "the Worker should NOT be fetched")
	}
}

func assertEqualSleepSchedule(t *testing.T, workerID uint, expect, actual SleepSchedule) {
	assert.Equal(t, workerID, actual.WorkerID, "sleep schedule is assigned to different worker")
	assert.Nil(t, actual.Worker, "the Worker itself should not be fetched")
	assert.Equal(t, expect.IsActive, actual.IsActive, "IsActive does not match")
	assert.Equal(t, expect.DaysOfWeek, actual.DaysOfWeek, "DaysOfWeek does not match")
	assert.Equal(t, expect.StartTime, actual.StartTime, "StartTime does not match")
	assert.Equal(t, expect.EndTime, actual.EndTime, "EndTime does not match")
}

func mustParseTime(timeString string) time.Time {
	parsed, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		panic(err)
	}
	return parsed
}
