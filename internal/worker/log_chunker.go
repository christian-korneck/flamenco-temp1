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
	"strings"
)

const defaultLogChunkerBufferFlushSize = 1024

// LogChunker gathers log lines in memory and sends them to a CommandListener.
// NOTE: LogChunker is not thread-safe.
type LogChunker struct {
	taskID          string
	listener        CommandListener
	buffer          strings.Builder
	bufferFlushSize int // When the buffer grows beyond this many bytes, flush.
}

func NewLogChunker(taskID string, listerer CommandListener) *LogChunker {
	return &LogChunker{
		taskID:          taskID,
		listener:        listerer,
		buffer:          strings.Builder{},
		bufferFlushSize: defaultLogChunkerBufferFlushSize,
	}
}

// Flush sends any buffered logs to the listener.
func (lc *LogChunker) Flush(ctx context.Context) error {
	if lc.buffer.Len() == 0 {
		return nil
	}

	err := lc.listener.LogProduced(ctx, lc.taskID, lc.buffer.String())
	lc.buffer.Reset()
	return err
}

func (lc *LogChunker) Append(ctx context.Context, logLines ...string) error {
	for idx := range logLines {
		lc.buffer.WriteString(logLines[idx])
		lc.buffer.WriteByte('\n')
	}

	if lc.buffer.Len() > lc.bufferFlushSize {
		return lc.Flush(ctx)
	}

	return nil
}
