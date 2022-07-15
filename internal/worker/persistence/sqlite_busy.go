// SPDX-License-Identifier: GPL-3.0-or-later
package persistence

import (
	"errors"
	"fmt"
	"time"

	"github.com/glebarez/go-sqlite"
	"gorm.io/gorm"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	// errDatabaseBusy is returned by this package when the operation could not be
	// performed due to SQLite being busy.
	errDatabaseBusy = errors.New("database busy")
)

// ErrIsDBBusy returns true when the error is a "database busy" error.
func ErrIsDBBusy(err error) bool {
	return errors.Is(err, errDatabaseBusy) || isDatabaseBusyError(err)
}

// isDatabaseBusyError returns true when the error returned by GORM is a
// SQLITE_BUSY error.
func isDatabaseBusyError(err error) bool {
	sqlErr, ok := err.(*sqlite.Error)
	return ok && sqlErr.Code() == sqlite3.SQLITE_BUSY
}

// setBusyTimeout sets the SQLite busy_timeout busy handler.
// See https://sqlite.org/pragma.html#pragma_busy_timeout
func setBusyTimeout(gormDB *gorm.DB, busyTimeout time.Duration) error {
	if tx := gormDB.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", busyTimeout.Milliseconds())); tx.Error != nil {
		return fmt.Errorf("setting busy_timeout: %w", tx.Error)
	}
	return nil
}
