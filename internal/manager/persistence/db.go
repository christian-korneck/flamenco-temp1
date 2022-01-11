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
	"fmt"

	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

// TODO : have this configurable from the CLI.
const dbURI = "flamenco-manager.sqlite"

// DB provides the database interface.
type DB struct {
	sqldb *sql.DB
}

func OpenDB(ctx context.Context) (*DB, error) {
	log.Info().Str("uri", dbURI).Msg("opening database")

	sqldb, err := sql.Open("sqlite", dbURI)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	if err := sqldb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error accessing database %s: %w", dbURI, err)
	}

	db := DB{
		sqldb: sqldb,
	}
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return &db, err
}
