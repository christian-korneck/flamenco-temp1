package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"

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
	db      *sql.DB
	dbMutex *sync.Mutex // Protects from "database locked" errors

	client        FlamencoClient
	clock         TimeService
	flushInterval time.Duration

	done chan struct{}
	wg   *sync.WaitGroup
}

const defaultUpstreamFlushInterval = 30 * time.Second
const databaseContextTimeout = 10 * time.Second

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

	db, err := sql.Open("sqlite", databaseFilename)
	if err != nil {
		return fmt.Errorf("opening %s: %w", databaseFilename, err)
	}

	if err := db.PingContext(dbCtx); err != nil {
		return fmt.Errorf("accessing %s: %w", databaseFilename, err)
	}

	ub.db = db

	if err := ub.prepareDatabase(dbCtx); err != nil {
		return err
	}

	ub.wg.Add(1)
	go ub.periodicFlushLoop()

	return nil
}

func (ub *UpstreamBufferDB) SendTaskUpdate(ctx context.Context, taskID string, update api.TaskUpdateJSONRequestBody) error {
	ub.dbMutex.Lock()
	defer ub.dbMutex.Unlock()

	queueSize, err := ub.queueSize()
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

// Close releases the database. It does not try to flush any pending items.
func (ub *UpstreamBufferDB) Close() error {
	if ub.db == nil {
		return nil
	}

	// Stop the periodic flush loop.
	close(ub.done)
	ub.wg.Wait()

	// Close the database.
	return ub.db.Close()
}

// prepareDatabase creates the database schema, if necessary.
func (ub *UpstreamBufferDB) prepareDatabase(dbCtx context.Context) error {
	ub.dbMutex.Lock()
	defer ub.dbMutex.Unlock()

	tx, err := ub.db.BeginTx(dbCtx, nil)
	if err != nil {
		return fmt.Errorf("beginning database transaction: %w", err)
	}
	defer rollbackTransaction(tx)

	stmt := `CREATE TABLE IF NOT EXISTS task_update_queue(task_id VARCHAR(36), payload BLOB)`
	log.Debug().Str("sql", stmt).Msg("creating database table")

	if _, err := tx.ExecContext(dbCtx, stmt); err != nil {
		return fmt.Errorf("creating database table: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commiting creation of database table: %w", err)
	}

	return nil
}

func (ub *UpstreamBufferDB) queueSize() (int, error) {
	if ub.db == nil {
		log.Panic().Msg("no database opened, unable to inspect upstream queue")
	}

	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), databaseContextTimeout)
	defer dbCtxCancel()

	var queueSize int

	err := ub.db.
		QueryRowContext(dbCtx, "SELECT count(*) FROM task_update_queue").
		Scan(&queueSize)

	switch {
	case err == sql.ErrNoRows:
		return 0, nil
	case err != nil:
		return 0, err
	default:
		return queueSize, nil
	}
}

func (ub *UpstreamBufferDB) queueTaskUpdate(taskID string, update api.TaskUpdateJSONRequestBody) error {
	if ub.db == nil {
		log.Panic().Msg("no database opened, unable to queue task updates")
	}

	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), databaseContextTimeout)
	defer dbCtxCancel()

	tx, err := ub.db.BeginTx(dbCtx, nil)
	if err != nil {
		return fmt.Errorf("beginning database transaction: %w", err)
	}
	defer rollbackTransaction(tx)

	blob, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("converting task update to JSON: %w", err)
	}

	stmt := `INSERT INTO task_update_queue (task_id, payload) VALUES (?, ?)`
	log.Debug().Str("sql", stmt).Str("task", taskID).Msg("inserting task update")

	if _, err := tx.ExecContext(dbCtx, stmt, taskID, blob); err != nil {
		return fmt.Errorf("queueing task update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing queued task update: %w", err)
	}

	return nil
}

func (ub *UpstreamBufferDB) QueueSize() (int, error) {
	ub.dbMutex.Lock()
	defer ub.dbMutex.Unlock()
	return ub.queueSize()
}

func (ub *UpstreamBufferDB) Flush(ctx context.Context) error {
	if ub.db == nil {
		log.Panic().Msg("no database opened, unable to queue task updates")
	}

	// See if we need to flush at all.
	ub.dbMutex.Lock()
	queueSize, err := ub.queueSize()
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
	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), databaseContextTimeout)
	defer dbCtxCancel()

	tx, err := ub.db.BeginTx(dbCtx, nil)
	if err != nil {
		return false, fmt.Errorf("beginning database transaction: %w", err)
	}
	defer rollbackTransaction(tx)

	stmt := `SELECT rowid, task_id, payload FROM task_update_queue ORDER BY rowid LIMIT 1`
	log.Trace().Str("sql", stmt).Msg("fetching queued task updates")

	var rowID int64
	var taskID string
	var blob []byte

	err = tx.QueryRowContext(dbCtx, stmt).Scan(&rowID, &taskID, &blob)
	switch {
	case err == sql.ErrNoRows:
		// Flush operation is done.
		log.Debug().Msg("task update queue empty")
		return true, nil
	case err != nil:
		return false, fmt.Errorf("querying task update queue: %w", err)
	}

	logger := log.With().Str("task", taskID).Logger()

	var update api.TaskUpdateJSONRequestBody
	if err := json.Unmarshal(blob, &update); err != nil {
		// If we can't unmarshal the queued task update, there is little else to do
		// than to discard it and ignore it ever happened.
		logger.Warn().Err(err).
			Msg("unable to unmarshal queued task update, discarding")
		if err := ub.discardRow(tx, rowID); err != nil {
			return false, err
		}
		return false, tx.Commit()
	}

	// actually attempt delivery.
	resp, err := ub.client.TaskUpdateWithResponse(ctx, taskID, update)
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

	if err := ub.discardRow(tx, rowID); err != nil {
		return false, err
	}
	return false, tx.Commit()
}

func (ub *UpstreamBufferDB) discardRow(tx *sql.Tx, rowID int64) error {
	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), databaseContextTimeout)
	defer dbCtxCancel()

	stmt := `DELETE FROM task_update_queue WHERE rowid = ?`
	log.Trace().Str("sql", stmt).Int64("rowID", rowID).Msg("un-queueing task update")

	_, err := tx.ExecContext(dbCtx, stmt, rowID)
	if err != nil {
		return fmt.Errorf("un-queueing task update: %w", err)
	}
	return nil
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

func rollbackTransaction(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		// log.Error().Err(err).Msg("rolling back transaction")
		log.Panic().Err(err).Msg("rolling back transaction")
	}
}
