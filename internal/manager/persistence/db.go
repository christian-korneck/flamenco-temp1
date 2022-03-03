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

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/glebarez/sqlite"
)

// DB provides the database interface.
type DB struct {
	gormDB *gorm.DB
}

func OpenDB(ctx context.Context, dsn string) (*DB, error) {
	db, err := openDB(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

func openDB(ctx context.Context, uri string) (*DB, error) {
	// TODO: don't log the password.
	log.Info().Str("dsn", uri).Msg("opening database")

	connection := sqlite.Open(uri)
	// connection := postgres.Open(uri)
	gormDB, err := gorm.Open(connection, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db := DB{
		gormDB: gormDB,
	}
	return &db, nil
}

// GormDB returns the GORM interface.
// This should only be used for the one Task Scheduler Monster Query. Other
// operations should just be implemented as a function on DB.
func (db *DB) GormDB() *gorm.DB {
	return db.gormDB
}
