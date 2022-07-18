package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"time"
)

const bgContextTimeout = 10 * time.Second

// bgContext returns a background context for background processing. This
// context MUST be used when a database query is meant to be independent of any
// API call that triggered it.
func bgContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), bgContextTimeout)
}
