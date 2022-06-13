package timeout_checker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/timeout_checker/mocks"
)

type TimeoutCheckerMocks struct {
	clock            *clock.Mock
	persist          *mocks.MockPersistenceService
	taskStateMachine *mocks.MockTaskStateMachine
	logStorage       *mocks.MockLogStorage
	broadcaster      *mocks.MockChangeBroadcaster

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
		broadcaster:      mocks.NewMockChangeBroadcaster(mockCtrl),

		wg: new(sync.WaitGroup),
	}

	mockedNow, err := time.Parse(time.RFC3339, "2022-06-09T12:00:00+00:00")
	if err != nil {
		panic(err)
	}
	mocks.clock.Set(mockedNow)

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
		workerTimeout,
		mocks.clock,
		mocks.persist,
		mocks.taskStateMachine,
		mocks.logStorage,
		mocks.broadcaster,
	)
	return sm, finish, mocks
}

// canaryTest will abort the current test if timing constants do not have the
// expected value. Unit tests will fail rather cryptically by themselves if the
// timing of the timeout checker is not what is expected.
func canaryTest(t *testing.T) {
	if assert.Equal(t, 5*time.Minute, timeoutInitialSleep, "timeoutInitialSleep does not have the expected value") &&
		assert.Equal(t, 1*time.Minute, timeoutCheckInterval, "timeoutCheckInterval does not have the expected value") {
		return
	}
	t.Fatal("timing-related constants are not as expected by the unit test, preemptively aborting.")
	t.FailNow()
}
