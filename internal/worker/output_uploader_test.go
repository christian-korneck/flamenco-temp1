package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/worker/mocks"
	"git.blender.org/flamenco/pkg/api"
)

func TestQueueOutput(t *testing.T) {
	ou, mocks, finish := mockedOutputUploader(t)
	defer finish()

	taskID := "094d98ba-d6e2-4765-a10b-70533604a952"

	// Run the queue process, otherwise the queued item cannot be received.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ou.Run(mocks.ctx)
	}()

	ou.OutputProduced(taskID, "filename.jpg")

	// The output should be queued for processing now.
	select {
	case item := <-ou.queue.Item():
		assert.Equal(t, item.Filename, "filename.jpg")
		assert.Equal(t, item.TaskID, taskID)
	case <-time.After(1 * time.Second):
		t.Fatal("output should be queued for processing")
	}

	mocks.ctxCancel()
	wg.Wait()
}

func TestProcess(t *testing.T) {
	ou, mocks, finish := mockedOutputUploader(t)
	defer finish()

	taskID := "094d98ba-d6e2-4765-a10b-70533604a952"
	filename := "command_ffmpeg_test_files/frame-1.png"

	item := TaskOutput{
		TaskID:   taskID,
		Filename: filename,
	}

	{
		// Test happy response from Manager.
		response := api.TaskOutputProducedResponse{
			HTTPResponse: &http.Response{
				Status:     "202 Accepted",
				StatusCode: http.StatusAccepted,
			},
		}
		mocks.client.EXPECT().TaskOutputProducedWithBodyWithResponse(
			mocks.ctx, taskID, "image/jpeg", gomock.Any()).
			Return(&response, nil)

		ou.process(mocks.ctx, item)
	}

	{
		// Test unhappy response from Manager (its queue is full).
		// The only difference with the happy flow is in the logging, so that's hard
		// to assert for here. It's a different flow though, with a different
		// non-nil pointer, so it's good to at least check it doesn't cause any
		// panics.
		response := api.TaskOutputProducedResponse{
			JSON429: &api.Error{
				Code:    http.StatusTooManyRequests,
				Message: "enhance your calm",
			},
		}
		mocks.client.EXPECT().TaskOutputProducedWithBodyWithResponse(
			mocks.ctx, taskID, "image/jpeg", gomock.Any()).
			Return(&response, nil)

		ou.process(mocks.ctx, item)
	}

}

type outputUploaderTestMocks struct {
	client    *mocks.MockFlamencoClient
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func mockedOutputUploader(t *testing.T) (*OutputUploader, *outputUploaderTestMocks, func()) {
	mockCtrl := gomock.NewController(t)

	ctx, cancel := context.WithCancel(context.Background())
	mocks := outputUploaderTestMocks{
		client:    mocks.NewMockFlamencoClient(mockCtrl),
		ctx:       ctx,
		ctxCancel: cancel,
	}
	ou := NewOutputUploader(mocks.client)

	finish := func() {
		cancel()
		mockCtrl.Finish()
	}
	return ou, &mocks, finish
}
