package cli_runner

// SPDX-License-Identifier: GPL-3.0-or-later

import "context"

type LogChunker interface {
	// Flush sends any buffered logs to the listener.
	Flush(ctx context.Context) error
	// Append log lines to the buffer, sending to the listener when the buffer gets too large.
	Append(ctx context.Context, logLines ...string) error
}
