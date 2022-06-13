// Package persistence provides the database interface for Flamenco Manager.
package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	// sqlite "git.blender.org/flamenco/pkg/gorm-modernc-sqlite"
	"github.com/glebarez/sqlite"
)

const vacuumPeriod = 1 * time.Hour

// DB provides the database interface.
type DB struct {
	gormDB *gorm.DB
}

// Model contains the common database fields for most model structs.
// It is a copy of the gorm.Model struct, but without the `DeletedAt` field.
// Soft deletion is not used by Flamenco. If it ever becomes necessary to
// support soft-deletion, see https://gorm.io/docs/delete.html#Soft-Delete
type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func OpenDB(ctx context.Context, dsn string) (*DB, error) {
	log.Info().Str("dsn", dsn).Msg("opening database")

	db, err := openDB(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := setBusyTimeout(db.gormDB, 5*time.Second); err != nil {
		return nil, err
	}

	// Perfom some maintenance at startup.
	db.vacuum()

	if err := db.migrate(); err != nil {
		return nil, err
	}
	log.Debug().Msg("database automigration succesful")

	return db, nil
}

func openDB(ctx context.Context, dsn string) (*DB, error) {
	globalLogLevel := log.Logger.GetLevel()
	dblogger := NewDBLogger(log.Level(globalLogLevel))

	config := gorm.Config{
		Logger:  dblogger,
		NowFunc: nowFunc,
	}

	return openDBWithConfig(dsn, &config)
}

func openDBWithConfig(dsn string, config *gorm.Config) (*DB, error) {
	dialector := sqlite.Open(dsn)
	gormDB, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}

	// Use the generic sql.DB interface to set some connection pool options.
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	// Only allow a single database connection, to avoid SQLITE_BUSY errors.
	// It's not certain that this'll improve the situation, but it's worth a try.
	sqlDB.SetMaxIdleConns(1) // Max num of connections in the idle connection pool.
	sqlDB.SetMaxOpenConns(1) // Max num of open connections to the database.

	db := DB{
		gormDB: gormDB,
	}

	return &db, nil
}

// nowFunc returns 'now' in UTC, so that GORM-managed times (createdAt,
// deletedAt, updatedAt) are stored in UTC.
func nowFunc() time.Time {
	return time.Now().UTC()
}

// PeriodicMaintenanceLoop periodically vacuums the database.
// This function only returns when the context is done.
func (db *DB) PeriodicMaintenanceLoop(ctx context.Context) {
	log.Debug().Msg("periodic database maintenance loop starting")
	defer log.Debug().Msg("periodic database maintenance loop stopping")

	var waitTime time.Duration

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitTime):
			waitTime = vacuumPeriod
		}

		log.Debug().Msg("vacuuming database")
		db.vacuum()
	}
}

// vacuum executes the SQL "VACUUM" command, and logs any errors.
func (db *DB) vacuum() {
	tx := db.gormDB.Exec("vacuum")
	if tx.Error != nil {
		log.Error().Err(tx.Error).Msg("error vacuuming database")
	}
}
