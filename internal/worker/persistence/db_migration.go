package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
)

func (db *DB) migrate() error {
	err := db.gormDB.AutoMigrate(
		&TaskUpdate{},
	)
	if err != nil {
		return fmt.Errorf("failed to automigrate database: %v", err)
	}
	return nil
}
