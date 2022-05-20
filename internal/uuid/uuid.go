// uuid is a thin wrapper around github.com/google/uuid.
package uuid

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"github.com/google/uuid"
)

// New generates a random UUID.
func New() string {
	return uuid.New().String()
}

// IsValid returns true when the string can be parsed as UUID.
func IsValid(value string) bool {
	// uuid.Parse() accepts a few different notations for UUIDs, but Flamenco only
	// works with the xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx format.
	if len(value) != 36 {
		return false
	}

	_, err := uuid.Parse(value)
	return err == nil
}
