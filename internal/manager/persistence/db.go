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

func OpenDB(ctx context.Context, dsn string) (*DB, error) {
	log.Info().Str("dsn", dsn).Msg("opening database")

	db, err := openDB(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.migrate(); err != nil {
		return nil, err
	}
	log.Debug().Msg("database automigration succesful")

	return db, nil
}

func openDB(ctx context.Context, uri string) (*DB, error) {
	globalLogLevel := log.Logger.GetLevel()
	dblogger := NewDBLogger(log.Level(globalLogLevel))

	config := gorm.Config{
		Logger: dblogger,
	}

	return openDBWithConfig(uri, &config)
}

func openDBWithConfig(uri string, config *gorm.Config) (*DB, error) {
	dialector := sqlite.Open(uri)
	gormDB, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}

	db := DB{
		gormDB: gormDB,
	}
	return &db, nil
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
		tx := db.gormDB.Exec("vacuum")
		if tx.Error != nil {
			log.Error().Err(tx.Error).Msg("error vacuuming database")
		}
	}
}
