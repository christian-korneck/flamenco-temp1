package worker

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
	"context"
	"errors"
	"os"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker/mocks"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"

	_ "modernc.org/sqlite"
)

const testBufferDBFilename = "test-flamenco-worker.db"

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
	err := ub.Close(context.Background())
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
	queueSize, err := ub.queueSize(ctx)
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
	queueSize, err = ub.queueSize(ctx)
	ub.dbMutex.Unlock()
	assert.NoError(t, err)
	assert.Equal(t, 0, queueSize)

	assert.NoError(t, ub.Close(ctx))
}
