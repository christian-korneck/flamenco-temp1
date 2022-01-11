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
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func init() {
	source.Register("embedfs", &EmbedFS{})
}

type EmbedFS struct {
	iofs.PartialDriver
}

//go:embed migrations/*.sql
var embedFS embed.FS

func (db *DB) migrate() error {
	driver, err := sqlite.WithInstance(db.sqldb, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("cannot create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("embedfs://", "sqlite", driver)
	if err != nil {
		return fmt.Errorf("cannot create migration instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("cannot migrate database: %w", err)
	}
	return nil
}

func (f *EmbedFS) Open(url string) (source.Driver, error) {
	nf := &EmbedFS{}
	if err := nf.Init(embedFS, "migrations"); err != nil {
		return nil, err
	}
	return nf, nil
}
