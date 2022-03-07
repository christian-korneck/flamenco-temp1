package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"strings"
	"time"
)

// When the buffer grows beyond this many bytes, flush.
const defaultLogChunkerBufferFlushSize = 1024

// When the last flush was this long ago, flush.
const defaultLogChunkerFlushMaxInterval = 30 * time.Second

// LogChunker gathers log lines in memory and sends them to a CommandListener.
// NOTE: LogChunker is not thread-safe.
type LogChunker struct {
	taskID string

	listener    CommandListener
	timeService TimeService

	buffer          strings.Builder
	bufferFlushSize int // When the buffer grows beyond this many bytes, flush.

	lastFlush  time.Time
	flushAfter time.Duration
}

func NewLogChunker(taskID string, listerer CommandListener, timeService TimeService) *LogChunker {
	return &LogChunker{
		taskID: taskID,

		listener:    listerer,
		timeService: timeService,

		buffer:          strings.Builder{},
		bufferFlushSize: defaultLogChunkerBufferFlushSize,

		lastFlush:  timeService.Now(),
		flushAfter: defaultLogChunkerFlushMaxInterval,
	}
}

// Flush sends any buffered logs to the listener.
func (lc *LogChunker) Flush(ctx context.Context) error {
	if lc.buffer.Len() == 0 {
		return nil
	}

	err := lc.listener.LogProduced(ctx, lc.taskID, lc.buffer.String())
	lc.buffer.Reset()
	lc.lastFlush = time.Now()
	return err
}

// Append log lines to the buffer, sending to the listener when the buffer gets too large.
func (lc *LogChunker) Append(ctx context.Context, logLines ...string) error {
	for idx := range logLines {
		lc.buffer.WriteString(logLines[idx])
		lc.buffer.WriteByte('\n')
	}

	if lc.shouldFlush() {
		return lc.Flush(ctx)
	}

	return nil
}

func (lc *LogChunker) shouldFlush() bool {
	if lc.buffer.Len() > lc.bufferFlushSize {
		return true
	}

	if lc.timeService.Now().Sub(lc.lastFlush) > lc.flushAfter {
		return true
	}

	return false
}
