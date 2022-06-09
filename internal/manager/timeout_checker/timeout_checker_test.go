package timeout_checker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/internal/manager/timeout_checker/mocks"
	"git.blender.org/flamenco/pkg/api"
)

const taskTimeout = 20 * time.Minute

func TestTimeoutCheckerTiming(t *testing.T) {
	ttc, finish, mocks := timeoutCheckerTestFixtures(t)
	defer finish()

	mocks.run(ttc)

	// Wait for the timeout checker to actually be sleeping, otherwise it could
	// have a different sleep-start time than we expect.
	time.Sleep(1 * time.Millisecond)

	// Determine the deadlines relative to the initial clock value.
	initialTime := mocks.clock.Now().UTC()
	deadlines := []time.Time{
		initialTime.Add(timeoutInitialSleep - taskTimeout),
		initialTime.Add(timeoutInitialSleep - taskTimeout + 1*timeoutCheckInterval),
		initialTime.Add(timeoutInitialSleep - taskTimeout + 2*timeoutCheckInterval),
	}

	// Expect three fetches, one after the initial sleep time, and two a regular interval later.
	fetchTimes := make([]time.Time, len(deadlines))
	firstCall := mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, deadlines[0]).
		DoAndReturn(func(ctx context.Context, timeout time.Time) ([]*persistence.Task, error) {
			fetchTimes[0] = mocks.clock.Now().UTC()
			return []*persistence.Task{}, nil
		})

	secondCall := mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, deadlines[1]).
		DoAndReturn(func(ctx context.Context, timeout time.Time) ([]*persistence.Task, error) {
			fetchTimes[1] = mocks.clock.Now().UTC()
			// Return a database error. This shouldn't break the check loop.
			return []*persistence.Task{}, errors.New("testing what errors do")
		}).
		After(firstCall)

	mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, deadlines[2]).
		DoAndReturn(func(ctx context.Context, timeout time.Time) ([]*persistence.Task, error) {
			fetchTimes[2] = mocks.clock.Now().UTC()
			return []*persistence.Task{}, nil
		}).
		After(secondCall)

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
		Model: gorm.Model{ID: 47},
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

	mocks.persist.EXPECT().FetchTimedOutTasks(mocks.ctx, gomock.Any()).
		Return([]*persistence.Task{&taskUnassigned, &taskUnknownWorker, &taskAssigned}, nil)

	mocks.taskStateMachine.EXPECT().TaskStatusChange(mocks.ctx, &taskUnassigned, api.TaskStatusFailed)
	mocks.taskStateMachine.EXPECT().TaskStatusChange(mocks.ctx, &taskUnknownWorker, api.TaskStatusFailed)
	mocks.taskStateMachine.EXPECT().TaskStatusChange(mocks.ctx, &taskAssigned, api.TaskStatusFailed)

	mocks.logStorage.EXPECT().WriteTimestamped(gomock.Any(), job.UUID, taskUnassigned.UUID,
		"Task timed out. It was assigned to worker -unassigned-, but untouched since 1969-12-31T23:00:00Z")
	mocks.logStorage.EXPECT().WriteTimestamped(gomock.Any(), job.UUID, taskUnknownWorker.UUID,
		"Task timed out. It was assigned to worker -unknown-, but untouched since 1969-12-31T23:00:00Z")
	mocks.logStorage.EXPECT().WriteTimestamped(gomock.Any(), job.UUID, taskAssigned.UUID,
		"Task timed out. It was assigned to worker Tester (WORKER-UUID), but untouched since 1969-12-31T23:00:00Z")

	// All the timeouts should be handled after the initial sleep.
	mocks.clock.Add(timeoutInitialSleep)
}

type TimeoutCheckerMocks struct {
	clock            *clock.Mock
	persist          *mocks.MockPersistenceService
	taskStateMachine *mocks.MockTaskStateMachine
	logStorage       *mocks.MockLogStorage

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

// run starts a goroutine to call ttc.Run(mocks.ctx).
func (mocks *TimeoutCheckerMocks) run(ttc *TimeoutChecker) {
	mocks.wg.Add(1)
	go func() {
		defer mocks.wg.Done()
		ttc.Run(mocks.ctx)
	}()
}

func timeoutCheckerTestFixtures(t *testing.T) (*TimeoutChecker, func(), *TimeoutCheckerMocks) {
	mockCtrl := gomock.NewController(t)

	mocks := &TimeoutCheckerMocks{
		clock:            clock.NewMock(),
		persist:          mocks.NewMockPersistenceService(mockCtrl),
		taskStateMachine: mocks.NewMockTaskStateMachine(mockCtrl),
		logStorage:       mocks.NewMockLogStorage(mockCtrl),

		wg: new(sync.WaitGroup),
	}

	// mockedNow, err := time.Parse(time.RFC3339, "2022-06-09T16:52:04+02:00")
	// if err != nil {
	// 	panic(err)
	// }
	// mocks.clock.Set(mockedNow)

	ctx, cancel := context.WithCancel(context.Background())
	mocks.ctx = ctx
	mocks.cancel = cancel

	// This should be called at the end of each unit test.
	finish := func() {
		mocks.cancel()
		mocks.wg.Wait()
		mockCtrl.Finish()
	}

	sm := New(
		taskTimeout,
		mocks.clock,
		mocks.persist,
		mocks.taskStateMachine,
		mocks.logStorage,
	)
	return sm, finish, mocks
}
