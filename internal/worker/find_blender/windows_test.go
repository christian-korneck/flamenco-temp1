//go:build windows

// SPDX-License-Identifier: GPL-3.0-or-later
package find_blender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFindBlender is a "weak" test, which actually accepts both happy and unhappy flows.
// It would be too fragile to always require a file association to be set up with Blender.
func TestFindBlender(t *testing.T) {
	exe, err := FindBlender()
	if err != nil {
		assert.Empty(t, exe)
		return
	}
	assert.NotEmpty(t, exe)
	assert.NotContains(t, exe,
		"blender-launcher",
		"FindBlender should find blender.exe, not blender-launcher.exe")
}
