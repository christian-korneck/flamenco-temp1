package timeout_checker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

const taskTimeout = 20 * time.Minute

func TestTimeoutCheckerTiming(t *testing.T) {
	canaryTest(t)

	ttc, finish, mocks := timeoutCheckerTestFixtures(t)
	defer finish()

	// Determine the deadlines relative to the initial clock value.
	initialTime := mocks.clock.Now().UTC()
	deadlines := []time.Time{
		initialTime.Add(-taskTimeout + timeoutInitialSleep + 0*timeoutCheckInterval),
		initialTime.Add(-taskTimeout + timeoutInitialSleep + 1*timeoutCheckInterval),
		initialTime.Add(-taskTimeout + timeoutInitialSleep + 2*timeoutCheckInterval),
	}

	mocks.run(ttc)

	// Wait for the timeout checker to actually be sleeping, otherwise it could
	// have a different sleep-start time than we expect.
	time.Sleep(1 * time.Millisecond)

	// No workers are timing out in this test.
	mocks.persist.EXPECT().FetchTimedOutWorkers(mocks.ctx, gomock.Any()).AnyTimes().Return(nil, nil)

	// Expect three fetches, one after the initial sleep time, and two a regular interval later.
	fetchTimes := make([]time.Time, 0)
	firstCall := mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, deadlines[0]).
		DoAndReturn(func(ctx context.Context, deadline time.Time) ([]*persistence.Task, error) {
			fetchTimes = append(fetchTimes, mocks.clock.Now().UTC())
			return []*persistence.Task{}, nil
		})

	secondCall := mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, deadlines[1]).
		DoAndReturn(func(ctx context.Context, deadline time.Time) ([]*persistence.Task, error) {
			fetchTimes = append(fetchTimes, mocks.clock.Now().UTC())
			// Return a database error. This shouldn't break the check loop.
			return []*persistence.Task{}, errors.New("testing what errors do")
		}).
		After(firstCall)

	thirdCall := mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, deadlines[2]).
		DoAndReturn(func(ctx context.Context, deadline time.Time) ([]*persistence.Task, error) {
			fetchTimes = append(fetchTimes, mocks.clock.Now().UTC())
			return []*persistence.Task{}, nil
		}).
		After(secondCall)

	// Having an AnyTimes() expectation here makes it possible to produce some
	// more sensible error messages than the mocking framework would give (which
	// would just abort the test saying the call doesn't match the above three
	// expectations).
	mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, deadline time.Time) ([]*persistence.Task, error) {
			fetchTimes = append(fetchTimes, mocks.clock.Now().UTC())
			assert.Failf(t, "extra call to FetchTimedOutTasks", "deadline=%s", deadline.String())
			return []*persistence.Task{}, nil
		}).
		After(thirdCall).
		AnyTimes()

	mocks.clock.Add(2 * time.Minute) // Should still be sleeping.
	mocks.clock.Add(2 * time.Minute) // Should still be sleeping.
	mocks.clock.Add(time.Minute)     // Should trigger the first fetch.
	mocks.clock.Add(time.Minute)     // Should trigger the second fetch.
	mocks.clock.Add(time.Minute)     // Should trigger the third fetch.

	// Wait for the timeout checker to actually run & hit the expected calls.
	time.Sleep(1 * time.Millisecond)

	for idx, fetchTime := range fetchTimes {
		// Check for zero values first, because they can be a bit confusing in the assert.Equal() logs.
		if !assert.Falsef(t, fetchTime.IsZero(), "fetchTime[%d] should not be zero", idx) {
			continue
		}
		expect := initialTime.Add(timeoutInitialSleep + time.Duration(idx)*timeoutCheckInterval)
		assert.Equalf(t, expect, fetchTime, "fetchTime[%d] not as expected", idx)
	}
}

func TestTaskTimeout(t *testing.T) {
	canaryTest(t)

	ttc, finish, mocks := timeoutCheckerTestFixtures(t)
	defer finish()

	mocks.run(ttc)

	// Wait for the timeout checker to actually be sleeping, otherwise it could
	// have a different sleep-start time than we expect.
	time.Sleep(1 * time.Millisecond)

	lastTime := mocks.clock.Now().UTC().Add(-1 * time.Hour)

	job := persistence.Job{UUID: "JOB-UUID"}
	worker := persistence.Worker{
		UUID:  "WORKER-UUID",
		Name:  "Tester",
		Model: persistence.Model{ID: 47},
	}
	taskUnassigned := persistence.Task{
		UUID:          "TASK-UUID-UNASSIGNED",
		Job:           &job,
		LastTouchedAt: lastTime,
	}
	taskUnknownWorker := persistence.Task{
		UUID:          "TASK-UUID-UNKNOWN",
		Job:           &job,
		LastTouchedAt: lastTime,
		WorkerID:      &worker.ID,
	}
	taskAssigned := persistence.Task{
		UUID:          "TASK-UUID-ASSIGNED",
		Job:           &job,
		LastTouchedAt: lastTime,
		WorkerID:      &worker.ID,
		Worker:        &worker,
	}

	mocks.persist.EXPECT().FetchTimedOutWorkers(mocks.ctx, gomock.Any()).AnyTimes().Return(nil, nil)

	mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, gomock.Any()).
		Return([]*persistence.Task{&taskUnassigned, &taskUnknownWorker, &taskAssigned}, nil)

	mocks.taskStateMachine.EXPECT().TaskStatusChange(mocks.ctx, &taskUnassigned, api.TaskStatusFailed)
	mocks.taskStateMachine.EXPECT().TaskStatusChange(mocks.ctx, &taskUnknownWorker, api.TaskStatusFailed)
	mocks.taskStateMachine.EXPECT().TaskStatusChange(mocks.ctx, &taskAssigned, api.TaskStatusFailed)

	mocks.logStorage.EXPECT().WriteTimestamped(gomock.Any(), job.UUID, taskUnassigned.UUID,
		"Task timed out. It was assigned to worker -unassigned-, but untouched since 2022-06-09T11:00:00Z")
	mocks.logStorage.EXPECT().WriteTimestamped(gomock.Any(), job.UUID, taskUnknownWorker.UUID,
		"Task timed out. It was assigned to worker -unknown-, but untouched since 2022-06-09T11:00:00Z")
	mocks.logStorage.EXPECT().WriteTimestamped(gomock.Any(), job.UUID, taskAssigned.UUID,
		"Task timed out. It was assigned to worker Tester (WORKER-UUID), but untouched since 2022-06-09T11:00:00Z")

	// All the timeouts should be handled after the initial sleep.
	mocks.clock.Add(timeoutInitialSleep)
}
