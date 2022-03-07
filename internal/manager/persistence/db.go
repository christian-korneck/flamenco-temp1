// Package persistence provides the database interface for Flamenco Manager.
package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	// sqlite "git.blender.org/flamenco/pkg/gorm-modernc-sqlite"
	"github.com/glebarez/sqlite"
)

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
