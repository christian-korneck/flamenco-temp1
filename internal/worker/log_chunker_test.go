package worker

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/worker/mocks"
)

// SPDX-License-Identifier: GPL-3.0-or-later

type LogChunkerMocks struct {
	listener *mocks.MockCommandListener
	clock    *clock.Mock
}

func mockedLogChunker(t *testing.T, mockCtrl *gomock.Controller) (*LogChunker, *LogChunkerMocks) {
	mocks := LogChunkerMocks{
		clock:    mockedClock(t),
		listener: mocks.NewMockCommandListener(mockCtrl),
	}
	lc := NewLogChunker("taskID", mocks.listener, mocks.clock)
	return lc, &mocks
}

func TestLogChunkerEmpty(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	lc, _ := mockedLogChunker(t, mockCtrl)

	// Note: no call to mockListener is expected.
	err := lc.Flush(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 0, lc.buffer.Len(), "buffer should be empty")
}

func TestLogChunkerSimple(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	lc, mocks := mockedLogChunker(t, mockCtrl)

	ctx := context.Background()
	mocks.listener.EXPECT().LogProduced(ctx, "taskID", "just one line\n")

	err := lc.Append(ctx, "just one line")
	assert.NoError(t, err)

	err = lc.Flush(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, lc.buffer.Len(), "buffer should be empty")
}

func TestLogChunkerMuchLogging(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	lc, mocks := mockedLogChunker(t, mockCtrl)
	lc.bufferFlushSize = 12

	ctx := context.Background()

	err := lc.Append(ctx, "één regel") // 9 runes, 11 bytes, 12 with newline, within buffer size.
	assert.NoError(t, err)

	mocks.listener.EXPECT().LogProduced(ctx, "taskID", "één regel\nsecond line\n")

	err = lc.Append(ctx, "second line") // this pushes the buffer over its max size.
	assert.NoError(t, err)

	assert.Equal(t, 0, lc.buffer.Len(), "buffer should be empty")
}

func TestLogChunkerTimedFlush(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	lc, mocks := mockedLogChunker(t, mockCtrl)
	lc.flushAfter = 2 * time.Second

	ctx := context.Background()

	err := lc.Append(ctx, "één regel") // No flush yet
	assert.NoError(t, err)

	mocks.clock.Add(2000 * time.Millisecond) // Exactly the threshold
	err = lc.Append(ctx, "second line")      // No flush yet
	assert.NoError(t, err)

	mocks.clock.Add(1 * time.Millisecond) // Juuuuust a bit longer than the threshold.

	mocks.listener.EXPECT().LogProduced(ctx, "taskID", "één regel\nsecond line\nthird line\n")

	err = lc.Append(ctx, "third line") // This should flush due to the long wait.
	assert.NoError(t, err)

	assert.Equal(t, 0, lc.buffer.Len(), "buffer should be empty")
}
