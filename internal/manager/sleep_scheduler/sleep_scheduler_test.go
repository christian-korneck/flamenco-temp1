package sleep_scheduler

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/sleep_scheduler/mocks"
	"git.blender.org/flamenco/pkg/api"
)

func TestFetchSchedule(t *testing.T) {
	ss, mocks, ctx := testFixtures(t)

	workerUUID := "aeb49d8a-6903-41b3-b545-77b7a1c0ca19"
	dbSched := persistence.SleepSchedule{}
	mocks.persist.EXPECT().FetchWorkerSleepSchedule(ctx, workerUUID).Return(&dbSched, nil)

	sched, err := ss.FetchSchedule(ctx, workerUUID)
	if assert.NoError(t, err) {
		assert.Equal(t, &dbSched, sched)
	}
}

func TestSetSchedule(t *testing.T) {
	ss, mocks, ctx := testFixtures(t)

	workerUUID := "aeb49d8a-6903-41b3-b545-77b7a1c0ca19"

	sched := persistence.SleepSchedule{
		IsActive:   true,
		DaysOfWeek: " mo  tu  we",
		StartTime:  mkToD(9, 0),
		EndTime:    mkToD(18, 0),

		Worker: &persistence.Worker{
			UUID:   workerUUID,
			Status: api.WorkerStatusAwake,
		},
	}
	expectSavedSchedule := sched
	expectSavedSchedule.DaysOfWeek = "mo tu we" // Expect a cleanup
	expectNextCheck := mocks.todayAt(18, 0)     // "now" is at 11:14:47, expect a check at the end time.
	expectSavedSchedule.NextCheck = expectNextCheck

	// Expect the new schedule to be saved.
	mocks.persist.EXPECT().SetWorkerSleepSchedule(ctx, workerUUID, &expectSavedSchedule)

	// Expect the new schedule to be immediately applied to the Worker.
	// `TestApplySleepSchedule` checks those values, no need to do that here.
	mocks.persist.EXPECT().SaveWorkerStatus(ctx, gomock.Any())
	mocks.broadcaster.EXPECT().BroadcastWorkerUpdate(gomock.Any())

	err := ss.SetSchedule(ctx, workerUUID, &sched)
	assert.NoError(t, err)
}

func TestSetScheduleSwappedStartEnd(t *testing.T) {
	ss, mocks, ctx := testFixtures(t)

	workerUUID := "aeb49d8a-6903-41b3-b545-77b7a1c0ca19"

	sched := persistence.SleepSchedule{
		IsActive:   true,
		DaysOfWeek: "mo tu we",
		StartTime:  mkToD(18, 0),
		EndTime:    mkToD(9, 0),

		// Worker already in the right state, so no saving/broadcasting expected.
		Worker: &persistence.Worker{
			UUID:   workerUUID,
			Status: api.WorkerStatusAsleep,
		},
	}

	expectSavedSchedule := persistence.SleepSchedule{
		IsActive:   true,
		DaysOfWeek: "mo tu we",
		StartTime:  mkToD(9, 0), // Expect start and end time to be corrected.
		EndTime:    mkToD(18, 0),
		NextCheck:  mocks.todayAt(18, 0), // "now" is at 11:14:47, expect a check at the end time.
		Worker:     sched.Worker,
	}

	mocks.persist.EXPECT().SetWorkerSleepSchedule(ctx, workerUUID, &expectSavedSchedule)

	err := ss.SetSchedule(ctx, workerUUID, &sched)
	assert.NoError(t, err)
}

func TestApplySleepSchedule(t *testing.T) {
	ss, mocks, ctx := testFixtures(t)

	worker := persistence.Worker{
		Model:  persistence.Model{ID: 5},
		UUID:   "74997de4-c530-4913-b89f-c489f14f7634",
		Status: api.WorkerStatusOffline,
	}

	sched := persistence.SleepSchedule{
		IsActive:   true,
		DaysOfWeek: "mo tu we",
		StartTime:  mkToD(9, 0),
		EndTime:    mkToD(18, 0),
	}

	testForExpectedStatus := func(expectedNewStatus api.WorkerStatus) {
		// Take a copy of the worker & schedule, for test isolation.
		testSchedule := sched
		testWorker := worker

		// Expect the Worker to be fetched.
		mocks.persist.EXPECT().FetchSleepScheduleWorker(ctx, &testSchedule).DoAndReturn(
			func(ctx context.Context, schedule *persistence.SleepSchedule) error {
				schedule.Worker = &testWorker
				return nil
			})

		// Construct the worker as we expect it to be saved to the database.
		savedWorker := testWorker
		savedWorker.LazyStatusRequest = false
		savedWorker.StatusRequested = expectedNewStatus
		mocks.persist.EXPECT().SaveWorkerStatus(ctx, &savedWorker)

		// Expect SocketIO broadcast.
		var sioUpdate api.SocketIOWorkerUpdate
		mocks.broadcaster.EXPECT().BroadcastWorkerUpdate(gomock.Any()).DoAndReturn(
			func(workerUpdate api.SocketIOWorkerUpdate) {
				sioUpdate = workerUpdate
			})

		// Actually apply the sleep schedule.
		err := ss.ApplySleepSchedule(ctx, &testSchedule)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		// Check the SocketIO broadcast.
		if sioUpdate.Id != "" {
			assert.Equal(t, testWorker.UUID, sioUpdate.Id)
			assert.False(t, sioUpdate.StatusChange.IsLazy)
			assert.Equal(t, expectedNewStatus, sioUpdate.StatusChange.Status)
		}
	}

	// Move the clock to the middle of the sleep schedule, so worker should sleep.
	mocks.clock.Set(mocks.todayAt(10, 47))
	testForExpectedStatus(api.WorkerStatusAsleep)

	// Move the clock to before the sleep schedule start.
	mocks.clock.Set(mocks.todayAt(0, 3))
	testForExpectedStatus(api.WorkerStatusAwake)

	// Move the clock to after the sleep schedule ends.
	mocks.clock.Set(mocks.todayAt(19, 59))
	testForExpectedStatus(api.WorkerStatusAwake)

	// Test that the worker should sleep, and has already been requested to sleep,
	// but lazily. This should trigger a non-lazy status change request.
	mocks.clock.Set(mocks.todayAt(10, 47))
	worker.Status = api.WorkerStatusAwake
	worker.StatusRequested = api.WorkerStatusAsleep
	worker.LazyStatusRequest = true
	testForExpectedStatus(api.WorkerStatusAsleep)
}

func TestApplySleepScheduleNoStatusChange(t *testing.T) {
	ss, mocks, ctx := testFixtures(t)

	worker := persistence.Worker{
		Model:  persistence.Model{ID: 5},
		UUID:   "74997de4-c530-4913-b89f-c489f14f7634",
		Status: api.WorkerStatusAsleep,
	}

	sched := persistence.SleepSchedule{
		IsActive:   true,
		DaysOfWeek: "mo tu we",
		StartTime:  mkToD(9, 0),
		EndTime:    mkToD(18, 0),
	}

	runTest := func() {
		// Take a copy of the worker & schedule, for test isolation.
		testSchedule := sched
		testWorker := worker

		// Expect the Worker to be fetched.
		mocks.persist.EXPECT().FetchSleepScheduleWorker(ctx, &testSchedule).DoAndReturn(
			func(ctx context.Context, schedule *persistence.SleepSchedule) error {
				schedule.Worker = &testWorker
				return nil
			})

		// Apply the sleep schedule. This should not trigger any persistence or broadcasts.
		err := ss.ApplySleepSchedule(ctx, &testSchedule)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	// Move the clock to the middle of the sleep schedule, so the schedule always
	// wants the worker to sleep.
	mocks.clock.Set(mocks.todayAt(10, 47))

	// Current status is already good.
	worker.Status = api.WorkerStatusAsleep
	runTest()

	// Current status is not the right one, but the requested status is already good.
	worker.Status = api.WorkerStatusAwake
	worker.StatusRequested = api.WorkerStatusAsleep
	worker.LazyStatusRequest = false
	runTest()

	// Current status is not the right one, but error state should not be overwrittne.
	worker.Status = api.WorkerStatusError
	worker.StatusRequested = ""
	worker.LazyStatusRequest = false
	runTest()
}

type TestMocks struct {
	clock       *clock.Mock
	persist     *mocks.MockPersistenceService
	broadcaster *mocks.MockChangeBroadcaster
}

// todayAt returns whatever the mocked clock's "now" is set to, with the time set
// to the given time. Seconds and sub-seconds are set to zero.
func (m *TestMocks) todayAt(hour, minute int) time.Time {
	now := m.clock.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
}

// endOfDay returns midnight of the day after whatever the mocked clock's "now" is set to.
func (m *TestMocks) endOfDay() time.Time {
	now := m.clock.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
}

func testFixtures(t *testing.T) (*SleepScheduler, TestMocks, context.Context) {
	ctx := context.Background()

	mockedClock := clock.NewMock()
	mockedNow, err := time.Parse(time.RFC3339, "2022-06-07T11:14:47+02:00")
	if err != nil {
		panic(err)
	}
	mockedClock.Set(mockedNow)
	if !assert.Equal(t, time.Tuesday.String(), mockedNow.Weekday().String()) {
		t.Fatal("tests assume 'now' is a Tuesday")
	}

	mockCtrl := gomock.NewController(t)
	mocks := TestMocks{
		clock:       mockedClock,
		persist:     mocks.NewMockPersistenceService(mockCtrl),
		broadcaster: mocks.NewMockChangeBroadcaster(mockCtrl),
	}
	ss := New(mocks.clock, mocks.persist, mocks.broadcaster)
	return ss, mocks, ctx
}

func mkToD(hour, minute int) persistence.TimeOfDay {
	return persistence.TimeOfDay{Hour: hour, Minute: minute}
}
