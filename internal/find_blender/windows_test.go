//go:build windows

package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFileAssociation is a "weak" test, which actually accepts both happy and unhappy flows.
// It would be too fragile to always require a file association to be set up with Blender.
func TestFileAssociation(t *testing.T) {
	exe, err := fileAssociation()
	if err != nil {
		assert.Empty(t, exe)

		if *withBlender {
			t.Fatalf("unexpected error: %v", err)
		} else {
			t.Skip("skipping test, -withBlender arg not passed")
		}
		return
	}
	assert.NotEmpty(t, exe)
	assert.NotContains(t, exe,
		"blender-launcher",
		"FindBlender should find blender.exe, not blender-launcher.exe")
}
