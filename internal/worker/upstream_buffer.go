package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/worker/persistence"
	"git.blender.org/flamenco/pkg/api"
)

// TODO: pull the SQLite stuff out of this file into a more global place, so
// that other areas of Flamenco Worker can also use it.

// Note that there are two contexts used in this file. One (`dbCtx`) is for
// database access and is a local, short-lived, background context. The other
// (`ctx`) is the one that's passed from the caller, which should indicate the
// global worker context. If that context is done, queueing updates in the
// database will still work, but all communication with Flamenco Manager will
// halt.

// UpstreamBufferDB implements the UpstreamBuffer interface using a database as backend.
type UpstreamBufferDB struct {
	db      UpstreamBufferPersistence
	dbMutex *sync.Mutex // Protects from "database locked" errors

	client        FlamencoClient
	clock         TimeService
	flushInterval time.Duration

	done chan struct{}
	wg   *sync.WaitGroup
}

type UpstreamBufferPersistence interface {
	UpstreamBufferQueueSize(ctx context.Context) (int, error)
	UpstreamBufferQueue(ctx context.Context, taskID string, apiTaskUpdate api.TaskUpdateJSONRequestBody) error
	UpstreamBufferFrontItem(ctx context.Context) (*persistence.TaskUpdate, error)
	UpstreamBufferDiscard(ctx context.Context, queuedTaskUpdate *persistence.TaskUpdate) error
	Close() error
}

const defaultUpstreamFlushInterval = 30 * time.Second
const databaseContextTimeout = 10 * time.Second
const flushOnShutdownTimeout = 5 * time.Second

var _ UpstreamBuffer = (*UpstreamBufferDB)(nil)

func NewUpstreamBuffer(client FlamencoClient, clock TimeService) (*UpstreamBufferDB, error) {
	ub := UpstreamBufferDB{
		db:      nil,
		dbMutex: new(sync.Mutex),

		client:        client,
		clock:         clock,
		flushInterval: defaultUpstreamFlushInterval,

		done: make(chan struct{}),
		wg:   new(sync.WaitGroup),
	}
	return &ub, nil
}

// OpenDB opens the database. Must be called once before using.
func (ub *UpstreamBufferDB) OpenDB(dbCtx context.Context, databaseFilename string) error {
	if ub.db != nil {
		return errors.New("upstream buffer database already opened")
	}

	db, err := persistence.OpenDB(dbCtx, databaseFilename)
	if err != nil {
		return fmt.Errorf("opening %s: %w", databaseFilename, err)
	}
	ub.db = db

	ub.wg.Add(1)
	go ub.periodicFlushLoop()

	return nil
}

func (ub *UpstreamBufferDB) SendTaskUpdate(ctx context.Context, taskID string, update api.TaskUpdateJSONRequestBody) error {
	ub.dbMutex.Lock()
	defer ub.dbMutex.Unlock()

	queueSize, err := ub.queueSize(ctx)
	if err != nil {
		return fmt.Errorf("unable to determine upstream queue size: %w", err)
	}

	// Immediately queue if there is already stuff queued, to ensure the order of updates is maintained.
	if queueSize > 0 {
		log.Debug().Int("queueSize", queueSize).
			Msg("task updates already queued, immediately queueing new update")
		return ub.queueTaskUpdate(taskID, update)
	}

	// Try to deliver the update.
	resp, err := ub.client.TaskUpdateWithResponse(ctx, taskID, update)
	if err != nil {
		log.Warn().Err(err).Str("task", taskID).
			Msg("error communicating with Manager, going to queue task update for sending later")
		return ub.queueTaskUpdate(taskID, update)
	}

	// The Manager responded, so no need to queue this update, even when there was an error.
	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusConflict:
		return ErrTaskReassigned
	default:
		return fmt.Errorf("unknown error from Manager, code %d: %v",
			resp.StatusCode(), resp.JSONDefault)
	}
}

// Close performs one final flush, then releases the database.
func (ub *UpstreamBufferDB) Close() error {
	if ub.db == nil {
		return nil
	}

	// Stop the periodic flush loop.
	close(ub.done)
	ub.wg.Wait()

	// Attempt one final flush, if it's fast enough:
	log.Info().Msg("upstream buffer shutting down, doing one final flush")
	flushCtx, ctxCancel := context.WithTimeout(context.Background(), flushOnShutdownTimeout)
	defer ctxCancel()
	if err := ub.Flush(flushCtx); err != nil {
		log.Warn().Err(err).Msg("error flushing upstream buffer at shutdown")
	}

	// Close the database.
	return ub.db.Close()
}

func (ub *UpstreamBufferDB) queueSize(ctx context.Context) (int, error) {
	if ub.db == nil {
		log.Panic().Msg("no database opened, unable to inspect upstream queue")
	}

	dbCtx, dbCtxCancel := context.WithTimeout(ctx, databaseContextTimeout)
	defer dbCtxCancel()

	return ub.db.UpstreamBufferQueueSize(dbCtx)
}

func (ub *UpstreamBufferDB) queueTaskUpdate(taskID string, update api.TaskUpdateJSONRequestBody) error {
	if ub.db == nil {
		log.Panic().Msg("no database opened, unable to queue task updates")
	}

	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), databaseContextTimeout)
	defer dbCtxCancel()

	return ub.db.UpstreamBufferQueue(dbCtx, taskID, update)
}

func (ub *UpstreamBufferDB) QueueSize() (int, error) {
	ub.dbMutex.Lock()
	defer ub.dbMutex.Unlock()
	return ub.queueSize(context.Background())
}

func (ub *UpstreamBufferDB) Flush(ctx context.Context) error {
	if ub.db == nil {
		log.Panic().Msg("no database opened, unable to queue task updates")
	}

	// See if we need to flush at all.
	ub.dbMutex.Lock()
	queueSize, err := ub.queueSize(ctx)
	ub.dbMutex.Unlock()

	switch {
	case err != nil:
		return fmt.Errorf("unable to determine queue size: %w", err)
	case queueSize == 0:
		log.Debug().Msg("task update queue empty, nothing to flush")
		return nil
	}

	// Keep flushing until the queue is empty or there is an error.
	var done bool
	for !done {
		ub.dbMutex.Lock()
		done, err = ub.flushFirstItem(ctx)
		ub.dbMutex.Unlock()

		if err != nil {
			return err
		}
	}

	return nil
}

func (ub *UpstreamBufferDB) flushFirstItem(ctx context.Context) (done bool, err error) {
	dbCtx, dbCtxCancel := context.WithTimeout(ctx, databaseContextTimeout)
	defer dbCtxCancel()

	queued, err := ub.db.UpstreamBufferFrontItem(dbCtx)
	if err != nil {
		return false, fmt.Errorf("finding first queued task update: %w", err)
	}
	if queued == nil {
		// Nothing is queued.
		return true, nil
	}

	logger := log.With().Str("task", queued.TaskID).Logger()

	apiTaskUpdate, err := queued.Unmarshal()
	if err != nil {
		// If we can't unmarshal the queued task update, there is little else to do
		// than to discard it and ignore it ever happened.
		logger.Warn().Err(err).
			Msg("unable to unmarshal queued task update, discarding")
		return false, ub.db.UpstreamBufferDiscard(dbCtx, queued)
	}

	// actually attempt delivery.
	resp, err := ub.client.TaskUpdateWithResponse(ctx, queued.TaskID, *apiTaskUpdate)
	if err != nil {
		logger.Info().Err(err).Msg("communication with Manager still problematic")
		return true, err
	}

	// Regardless of the response, there is little else to do but to discard the
	// update from the queue.
	switch resp.StatusCode() {
	case http.StatusNoContent:
		logger.Debug().Msg("queued task updated accepted by Manager")
	case http.StatusConflict:
		logger.Warn().Msg("queued task update discarded by Manager, task was already reassigned to other Worker")
	default:
		logger.Warn().
			Int("statusCode", resp.StatusCode()).
			Interface("response", resp.JSONDefault).
			Msg("queued task update discarded by Manager, unknown reason")
	}

	if err := ub.db.UpstreamBufferDiscard(dbCtx, queued); err != nil {
		return false, err
	}
	return false, nil
}

func (ub *UpstreamBufferDB) periodicFlushLoop() {
	defer ub.wg.Done()
	defer log.Debug().Msg("periodic task update flush loop stopping")
	log.Debug().Msg("periodic task update flush loop starting")

	ctx := context.Background()

	for {
		select {
		case <-ub.done:
			return
		case <-ub.clock.After(ub.flushInterval):
			log.Trace().Msg("task upstream queue: periodic flush")
			err := ub.Flush(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("error flushing task update queue")
			}
		}
	}
}
