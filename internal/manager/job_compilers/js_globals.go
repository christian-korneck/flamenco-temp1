package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// ----------------------------------------------------------
// Functions that start with `js` are exposed to JavaScript.
// See newGojaVM() for the actual expose-as-globals code.
// ----------------------------------------------------------

func jsPrint(call goja.FunctionCall) goja.Value {
	log.Info().Interface("args", call.Arguments).Msg("print")
	return goja.Undefined()
}

func jsAlert(call goja.FunctionCall) goja.Value {
	log.Warn().Interface("args", call.Arguments).Msg("alert")
	return goja.Undefined()
}

// jsFormatTimestampLocal returns the timestamp formatted as local time in a way that's compatible with filenames.
func jsFormatTimestampLocal(timestamp time.Time) string {
	return timestamp.Local().Format("2006-01-02_150405")
}

type ErrInvalidRange struct {
	Range   string // The frame range that was invalid.
	Message string // The error message
	err     error  // Any wrapped error
}

func (e ErrInvalidRange) Error() string {
	if e.err != nil {
		return fmt.Sprintf("invalid range \"%v\":  %s (%s)", e.Range, e.Message, e.Error())
	}
	return fmt.Sprintf("invalid range \"%v\": %s", e.Range, e.Message)
}

func (e ErrInvalidRange) Unwrap() error {
	return e.err
}

func errInvalidRange(theRange, message string, errs ...error) error {
	e := ErrInvalidRange{
		Range:   theRange,
		Message: message,
	}
	for _, err := range errs {
		if err != nil {
			e.err = err
			break
		}
	}
	return e
}

const (
	chunkRegular = "-"
	chunkBlender = ".."
)

// jsFrameChunker takes a range like "1..10,20..25,40" and returns chunked ranges.
//
// The returned ranges will be at most `chunkSize` frames long.
//
// Supports "regular" and "blender" notation, resp. "A-Z" and "A..Z". Returned
// chunks will always be in "regular" notation because they're more compatible
// with embedding in filenames.
func jsFrameChunker(frameRange string, chunkSize int) ([]string, error) {
	frameRange = strings.TrimSpace(frameRange)
	if len(frameRange) == 0 {
		return nil, errInvalidRange(frameRange, "empty range")
	}
	if chunkSize < 1 {
		return nil, fmt.Errorf("invalid chunk size, must be positive number: %d", chunkSize)
	}

	frames, err := frameRangeExplode(frameRange)
	if err != nil {
		return nil, err
	}
	if len(frames) == 0 {
		return nil, errInvalidRange(frameRange, "empty range")
	}

	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	var i int
	chunks := make([]string, 0)
	for i = 0; i < len(frames); i += chunkSize {
		chunkFrames := frames[i:min(i+chunkSize, len(frames))]
		chunkRange := frameRangeMerge(chunkFrames)
		chunks = append(chunks, chunkRange)
	}

	return chunks, nil
}

// Given a range of frames, return an array containing each frame number.
func frameRangeExplode(frameRange string) ([]int, error) {
	// Store as map to avoid duplicate frames.
	frames := make(map[int]struct{}, 0)

	// Convert from "blender" to "regular" range notation.
	frameRange = strings.ReplaceAll(frameRange, chunkBlender, chunkRegular)

	// parseInt first trims whitespace before converting to integer.
	parseInt := func(s string) (int64, error) {
		return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	}

	// Explode each comma-separated frame range.
	for _, part := range strings.Split(frameRange, ",") {
		startEnd := strings.Split(part, chunkRegular)
		switch len(startEnd) {
		case 1: // Single frame
			frame, err := parseInt(startEnd[0])
			if err != nil {
				return nil, errInvalidRange(frameRange, part, err)
			}
			frames[int(frame)] = struct{}{}
		case 2: // Frame range A-B
			startFrame, startErr := parseInt(startEnd[0])
			endFrame, endErr := parseInt(startEnd[1])
			if startErr != nil || endErr != nil {
				return nil, errInvalidRange(frameRange, part, startErr, endErr)
			}
			for frame := startFrame; frame <= endFrame; frame++ {
				frames[int(frame)] = struct{}{}
			}
		default:
			return nil, errInvalidRange(frameRange, part)
		}
	}

	// Convert from map to sorted array.
	frameList := make([]int, 0, len(frames))
	for frame := range frames {
		frameList = append(frameList, frame)
	}
	sort.Ints(frameList)
	return frameList, nil
}

// frameRangeMerge merges consecutive frames into ranges like "3..8,13,15..17".
func frameRangeMerge(frames []int) string {
	startFrame := frames[0]
	prevFrame := frames[0]

	ranges := make([]string, 0)

	appendRange := func(fromFrame, toFrame int) {
		switch {
		case fromFrame == toFrame: // Last range was one frame only
			ranges = append(ranges, strconv.FormatInt(int64(fromFrame), 10))
		case fromFrame+1 == toFrame: // Last range was only two frames
			ranges = append(ranges, strconv.FormatInt(int64(fromFrame), 10))
			ranges = append(ranges, strconv.FormatInt(int64(toFrame), 10))
		default:
			ranges = append(ranges, fmt.Sprintf("%v%s%v", fromFrame, chunkRegular, toFrame))
		}
	}

	var currentFrame int
	for _, currentFrame = range frames {
		if currentFrame > prevFrame+1 {
			// This frame starts a new range, so append the one we now know ended.
			appendRange(startFrame, prevFrame)
			startFrame = currentFrame
		}
		prevFrame = currentFrame
	}
	appendRange(startFrame, currentFrame)

	return strings.Join(ranges, ",")
}
