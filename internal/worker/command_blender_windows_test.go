//go:build windows

// SPDX-License-Identifier: GPL-3.0-or-later
package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFindBlender is a "weak" test, which actually accepts both happy and unhappy flows.
// It would be too fragile to always require a file association to be set up with Blender.
func TestFindBlender(t *testing.T) {
	exe, err := FindBlender()
	if err == nil {
		assert.NotEmpty(t, exe)
	} else {
		assert.Empty(t, exe)
	}
}
