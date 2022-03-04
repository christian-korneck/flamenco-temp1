// Package persistence provides the database interface for Flamenco Manager.
package persistence

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
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Change this to a filename if you want to run a single test and inspect the
// resulting database.
const TestDSN = "file::memory:"

func CreateTestDB(t *testing.T) (db *DB, closer func()) {
	// Delete the SQLite file if it exists on disk.
	if _, err := os.Stat(TestDSN); err == nil {
		if err := os.Remove(TestDSN); err != nil {
			t.Fatalf("unable to remove %s: %v", TestDSN, err)
		}
	}

	var err error

	dblogger := NewDBLogger(log.Level(zerolog.InfoLevel).Output(os.Stdout))

	// Open the database ourselves, so that we have a low-level connection that
	// can be closed when the unit test is done running.
	sqliteConn, err := sql.Open(sqlite.DriverName, TestDSN)
	if err != nil {
		t.Fatalf("opening SQLite connection: %v", err)
	}

	config := gorm.Config{
		Logger:   dblogger,
		ConnPool: sqliteConn,
	}

	db, err = openDBWithConfig(TestDSN, &config)
	if err != nil {
		t.Fatalf("opening DB: %v", err)
	}

	err = db.migrate()
	if err != nil {
		t.Fatalf("migrating DB: %v", err)
	}

	closer = func() {
		if err := sqliteConn.Close(); err != nil {
			t.Fatalf("closing DB: %v", err)
		}
	}

	return db, closer
}

// persistenceTestFixtures creates a test database and returns it and a context.
// Tests should call the returned cancel function when they're done.
func persistenceTestFixtures(t *testing.T, testContextTimeout time.Duration) (context.Context, context.CancelFunc, *DB) {
	db, dbCloser := CreateTestDB(t)
	ctx, ctxCancel := context.WithTimeout(context.Background(), testContextTimeout)

	cancel := func() {
		ctxCancel()
		dbCloser()
	}

	return ctx, cancel, db
}
