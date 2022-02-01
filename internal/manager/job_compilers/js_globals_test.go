package job_compilers

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameChunkerHappyBlenderStyle(t *testing.T) {
	chunks, err := jsFrameChunker("1..10,20..25,40,3..8", 4)
	assert.Nil(t, err)
	assert.Equal(t, []string{"1-4", "5-8", "9,10,20,21", "22-25", "40"}, chunks)
}

func TestFrameChunkerHappySmallInput(t *testing.T) {
	// No frames, should be an error
	_, err := jsFrameChunker("   ", 4)
	assert.ErrorIs(t, err, ErrInvalidRange{Message: "empty range"})

	// Just one frame.
	chunks, err := jsFrameChunker("47", 4)
	assert.Nil(t, err)
	assert.Equal(t, []string{"47"}, chunks)

	// Just one range of exactly one chunk.
	chunks, err = jsFrameChunker("1-3", 3)
	assert.Nil(t, err)
	assert.Equal(t, []string{"1-3"}, chunks)
}

func TestFrameChunkerHappyRegularStyle(t *testing.T) {
	chunks, err := jsFrameChunker("1-10,20-25,40", 4)
	assert.Nil(t, err)
	assert.Equal(t, []string{"1-4", "5-8", "9,10,20,21", "22-25", "40"}, chunks)
}

func TestFrameChunkerHappyExtraWhitespace(t *testing.T) {
	chunks, err := jsFrameChunker(" 1  .. 10,\t20..25\n,40   ", 4)
	assert.Nil(t, err)
	assert.Equal(t, []string{"1-4", "5-8", "9,10,20,21", "22-25", "40"}, chunks)
}

func TestFrameRangeExplode(t *testing.T) {
	frames, err := frameRangeExplode("1..10,20..25,40")
	assert.Nil(t, err)
	assert.Equal(t, []int{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		20, 21, 22, 23, 24, 25, 40,
	}, frames)
}
