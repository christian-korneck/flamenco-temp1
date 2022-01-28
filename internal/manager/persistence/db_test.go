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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

const testURI = "host=localhost user=flamenco password=flamenco dbname=flamenco-test TimeZone=Europe/Amsterdam"

func createTestDB(t *testing.T) *DB {
	// Creating a new database should be fast.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	db, err := openDB(ctx, testURI)
	assert.NoError(t, err)

	// Erase everything in the database.
	var tx *gorm.DB
	tx = db.gormDB.Exec("DROP SCHEMA public CASCADE")
	assert.NoError(t, tx.Error)
	tx = db.gormDB.Exec("CREATE SCHEMA public")
	assert.NoError(t, tx.Error)

	// Restore default grants (source: https://stackoverflow.com/questions/3327312/how-can-i-drop-all-the-tables-in-a-postgresql-database)
	tx = db.gormDB.Exec("GRANT ALL ON SCHEMA public TO postgres")
	assert.NoError(t, tx.Error)
	tx = db.gormDB.Exec("GRANT ALL ON SCHEMA public TO public")
	assert.NoError(t, tx.Error)

	err = db.migrate()
	assert.NoError(t, err)

	return db
}
