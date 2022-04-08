package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"

	"git.blender.org/flamenco/internal/worker/mocks"
	"git.blender.org/flamenco/pkg/api"
)

const testBufferDBFilename = ":memory:"

type UpstreamBufferDBMocks struct {
	client *mocks.MockFlamencoClient
	clock  *clock.Mock
}

func mockUpstreamBufferDB(t *testing.T, mockCtrl *gomock.Controller) (*UpstreamBufferDB, *UpstreamBufferDBMocks) {
	mocks := UpstreamBufferDBMocks{
		client: mocks.NewMockFlamencoClient(mockCtrl),
		clock:  clock.NewMock(),
	}

	// Always start tests with a fresh database.
	os.Remove(testBufferDBFilename)

	ub, err := NewUpstreamBuffer(mocks.client, mocks.clock)
	if err != nil {
		t.Fatalf("unable to create upstream buffer: %v", err)
	}

	return ub, &mocks
}

func TestUpstreamBufferCloseUnopened(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ub, _ := mockUpstreamBufferDB(t, mockCtrl)
	err := ub.Close()
	assert.NoError(t, err, "Closing without opening should be OK")
}

func TestUpstreamBufferManagerUnavailable(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()

	ub, mocks := mockUpstreamBufferDB(t, mockCtrl)
	assert.NoError(t, ub.OpenDB(ctx, testBufferDBFilename))

	// Send a task update without Manager available.
	taskID := "3960dec4-978e-40ab-bede-bfa6428c6ebc"
	update := api.TaskUpdateJSONRequestBody{
		Activity:   ptr("Testing da ünits"),
		Log:        ptr("¿Unicode logging should work?"),
		TaskStatus: ptr(api.TaskStatusActive),
	}

	updateError := errors.New("mock manager unavailable")
	managerCallFail := mocks.client.EXPECT().
		TaskUpdateWithResponse(ctx, taskID, update).
		Return(nil, updateError)

	err := ub.SendTaskUpdate(ctx, taskID, update)
	assert.NoError(t, err)

	// Check the queue size, it should have an item queued.
	queueSize, err := ub.queueSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, queueSize)

	// Wait for the flushing with Manager available.
	resp := &api.TaskUpdateResponse{}
	mocks.client.EXPECT().
		TaskUpdateWithResponse(ctx, taskID, update).
		Return(resp, nil).
		After(managerCallFail)

	// Only add exactly the flush interval, as that maximises the chances of
	// getting conflicts on the database level (if we didn't have the
	// database-protection mutex).
	mocks.clock.Add(defaultUpstreamFlushInterval)

	// Queue should be empty now.
	ub.dbMutex.Lock()
	queueSize, err = ub.queueSize()
	ub.dbMutex.Unlock()
	assert.NoError(t, err)
	assert.Equal(t, 0, queueSize)

	assert.NoError(t, ub.Close())
}
