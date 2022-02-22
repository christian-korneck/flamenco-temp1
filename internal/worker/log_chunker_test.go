package worker

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker/mocks"
)

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

func TestLogChunkerEmpty(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockListener := mocks.NewMockCommandListener(mockCtrl)
	lc := NewLogChunker("taskID", mockListener)

	// Note: no call to mockListener is expected.
	err := lc.Flush(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 0, lc.buffer.Len(), "buffer should be empty")
}

func TestLogChunkerSimple(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockListener := mocks.NewMockCommandListener(mockCtrl)
	lc := NewLogChunker("taskID", mockListener)

	ctx := context.Background()
	mockListener.EXPECT().LogProduced(ctx, "taskID", "just one line\n")

	err := lc.Append(ctx, "just one line")
	assert.NoError(t, err)

	err = lc.Flush(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, lc.buffer.Len(), "buffer should be empty")
}

func TestLogChunkerMuchLogging(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockListener := mocks.NewMockCommandListener(mockCtrl)
	lc := NewLogChunker("taskID", mockListener)
	lc.bufferFlushSize = 12

	ctx := context.Background()

	err := lc.Append(ctx, "één regel") // 9 runes, 11 bytes, 12 with newline, within buffer size.
	assert.NoError(t, err)

	mockListener.EXPECT().LogProduced(ctx, "taskID", "één regel\nsecond line\n")

	err = lc.Append(ctx, "second line") // this pushes the buffer over its max size.
	assert.NoError(t, err)

	assert.Equal(t, 0, lc.buffer.Len(), "buffer should be empty")
}
