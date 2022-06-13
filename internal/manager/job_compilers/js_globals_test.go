package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameChunkerHappyBlenderStyle(t *testing.T) {
	chunks, err := jsFrameChunker("1..10,20..25,40,3..8", 4)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1-4", "5-8", "9,10,20,21", "22-25", "40"}, chunks)
}

func TestFrameChunkerHappySmallInput(t *testing.T) {
	// No frames, should be an error
	_, err := jsFrameChunker("   ", 4)
	assert.ErrorIs(t, err, ErrInvalidRange{Message: "empty range"})

	// Just one frame.
	chunks, err := jsFrameChunker("47", 4)
	assert.NoError(t, err)
	assert.Equal(t, []string{"47"}, chunks)

	// Just one range of exactly one chunk.
	chunks, err = jsFrameChunker("1-3", 3)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1-3"}, chunks)
}

func TestFrameChunkerHappyRegularStyle(t *testing.T) {
	chunks, err := jsFrameChunker("1-10,20-25,40", 4)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1-4", "5-8", "9,10,20,21", "22-25", "40"}, chunks)
}

func TestFrameChunkerHappyExtraWhitespace(t *testing.T) {
	chunks, err := jsFrameChunker(" 1  .. 10,\t20..25\n,40   ", 4)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1-4", "5-8", "9,10,20,21", "22-25", "40"}, chunks)
}

func TestFrameRangeExplode(t *testing.T) {
	frames, err := frameRangeExplode("1..10,20..25,40")
	assert.NoError(t, err)
	assert.Equal(t, []int{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		20, 21, 22, 23, 24, 25, 40,
	}, frames)
}
