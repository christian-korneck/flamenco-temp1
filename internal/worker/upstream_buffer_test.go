package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/worker/mocks"
	"git.blender.org/flamenco/pkg/api"
)

type UpstreamBufferDBMocks struct {
	client *mocks.MockFlamencoClient
	clock  *clock.Mock
}

func mockUpstreamBufferDB(t *testing.T, mockCtrl *gomock.Controller) (*UpstreamBufferDB, *UpstreamBufferDBMocks) {
	mocks := UpstreamBufferDBMocks{
		client: mocks.NewMockFlamencoClient(mockCtrl),
		clock:  clock.NewMock(),
	}

	ub, err := NewUpstreamBuffer(mocks.client, mocks.clock)
	if err != nil {
		t.Fatalf("unable to create upstream buffer: %v", err)
	}

	return ub, &mocks
}

// sqliteTestDBName returns a DSN for SQLite that separates tests from each
// other, but lets all connections made within the same test to connect to the
// same in-memory instance.
func sqliteTestDBName(t *testing.T) string {
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
}

func TestUpstreamBufferCloseUnopened(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ub, _ := mockUpstreamBufferDB(t, mockCtrl)
	err := ub.Close()
	assert.NoError(t, err, "Closing without opening should be OK")
}

func TestUpstreamBufferManagerUnavailable(t *testing.T) {
	// FIXME: This test is unreliable. The `wg.Wait()` function below can wait
	// indefinitely in some situations, which points at a timing issue between
	// various goroutines.
	t.Skip("Skipping test, it is unreliable.")

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()

	ub, mocks := mockUpstreamBufferDB(t, mockCtrl)
	assert.NoError(t, ub.OpenDB(ctx, sqliteTestDBName(t)))

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

	// Make it possible to wait for the queued item to be sent to the Manager.
	wg := sync.WaitGroup{}
	wg.Add(1)
	mocks.client.EXPECT().
		TaskUpdateWithResponse(ctx, taskID, update).
		DoAndReturn(func(ctx context.Context, taskID string, body api.TaskUpdateJSONRequestBody, editors ...api.RequestEditorFn) (*api.TaskUpdateResponse, error) {
			wg.Done()
			return &api.TaskUpdateResponse{}, nil
		}).
		After(managerCallFail)

	err := ub.SendTaskUpdate(ctx, taskID, update)
	assert.NoError(t, err)

	mocks.clock.Add(defaultUpstreamFlushInterval)

	// Do the actual waiting.
	wg.Wait()

	// Queue should be empty now.
	queueSize, err := ub.QueueSize()
	assert.NoError(t, err)
	assert.Equal(t, 0, queueSize)

	assert.NoError(t, ub.Close())
}

func TestStressingBuffer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping potentially heavy test due to -short CLI arg")
		return
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := context.Background()

	ub, mocks := mockUpstreamBufferDB(t, mockCtrl)
	assert.NoError(t, ub.OpenDB(ctx, sqliteTestDBName(t)))

	// Queue task updates much faster than the Manager can handle.
	taskID := "3960dec4-978e-40ab-bede-bfa6428c6ebc"
	update := api.TaskUpdateJSONRequestBody{
		Activity:   ptr("Testing da ünits"),
		Log:        ptr("¿Unicode logging should work?"),
		TaskStatus: ptr(api.TaskStatusActive),
	}

	// Make the Manager slow to respond.
	const managerResponseTime = 250 * time.Millisecond
	mocks.client.EXPECT().
		TaskUpdateWithResponse(ctx, taskID, update).
		DoAndReturn(func(ctx context.Context, taskID string, body api.TaskUpdateJSONRequestBody, editors ...api.RequestEditorFn) (*api.TaskUpdateResponse, error) {
			time.Sleep(managerResponseTime)
			return &api.TaskUpdateResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
			}, nil
		}).
		AnyTimes()

	// Send updates MUCH faster than the slowed-down Manager can handle.
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			err := ub.SendTaskUpdate(ctx, taskID, update)
			assert.NoError(t, err)
		}()

		// Also mix in a bunch of flushes.
		go func() {
			defer wg.Done()
			_, err := ub.flushFirstItem(ctx)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

}
