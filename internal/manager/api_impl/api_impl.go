// Package api_impl implements the OpenAPI API from pkg/api/flamenco-openapi.yaml.
package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"git.blender.org/flamenco/pkg/api"
	"github.com/labstack/echo/v4"
)

type Flamenco struct {
	jobCompiler  JobCompiler
	persist      PersistenceService
	broadcaster  ChangeBroadcaster
	logStorage   LogStorage
	config       ConfigService
	stateMachine TaskStateMachine
	shaman       Shaman
	clock        TimeService
	lastRender   LastRendered
	localStorage LocalStorage

	// The task scheduler can be locked to prevent multiple Workers from getting
	// the same task. It is also used for certain other queries, like
	// `MayWorkerRun` to prevent similar race conditions.
	taskSchedulerMutex sync.Mutex

	// done is closed by Flamenco when it wants the application to shut down and
	// restart itself from scratch.
	done chan struct{}
}

var _ api.ServerInterface = (*Flamenco)(nil)

// NewFlamenco creates a new Flamenco service.
func NewFlamenco(
	jc JobCompiler,
	jps PersistenceService,
	b ChangeBroadcaster,
	logStorage LogStorage,
	cs ConfigService,
	sm TaskStateMachine,
	sha Shaman,
	ts TimeService,
	lr LastRendered,
	localStorage LocalStorage,
) *Flamenco {
	return &Flamenco{
		jobCompiler:  jc,
		persist:      jps,
		broadcaster:  b,
		logStorage:   logStorage,
		config:       cs,
		stateMachine: sm,
		shaman:       sha,
		clock:        ts,
		lastRender:   lr,
		localStorage: localStorage,

		done: make(chan struct{}),
	}
}

// WaitForShutdown waits until Flamenco wants to shut down the application.
// Returns `true` when the application should restart.
// Returns `false` when the context closes.
func (f *Flamenco) WaitForShutdown(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-f.done:
		return true
	}
}

// requestShutdown closes the 'done' channel, signalling to callers of
// WaitForShutdown() that a shutdown is requested.
func (f *Flamenco) requestShutdown() {
	defer func() {
		// Recover the panic that happens when the channel is closed multiple times.
		// Requesting a shutdown should be possible multiple times without panicing.
		recover()
	}()
	close(f.done)
}

// sendAPIError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendAPIError(e echo.Context, code int, message string, args ...interface{}) error {
	if len(args) > 0 {
		// Only interpret 'message' as format string if there are actually format parameters.
		message = fmt.Sprintf(message, args...)
	}

	apiErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	return e.JSON(code, apiErr)
}

// sendAPIErrorDBBusy sends a HTTP 503 Service Unavailable, with a hopefully
// reasonable "retry after" header.
func sendAPIErrorDBBusy(e echo.Context, message string, args ...interface{}) error {
	if len(args) > 0 {
		// Only interpret 'message' as format string if there are actually format parameters.
		message = fmt.Sprintf(message, args)
	}

	code := http.StatusServiceUnavailable
	apiErr := api.Error{
		Code:    int32(code),
		Message: message,
	}

	retryAfter := 1 * time.Second
	seconds := int64(retryAfter.Seconds())
	e.Response().Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
	return e.JSON(code, apiErr)
}
