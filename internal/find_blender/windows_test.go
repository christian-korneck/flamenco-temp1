//go:build windows

package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var withBlender = flag.Bool("withBlender", false, "run test that requires Blender to be installed")

// TestFindBlender is a "weak" test, which actually accepts both happy and unhappy flows.
// It would be too fragile to always require a file association to be set up with Blender.
func TestFindBlender(t *testing.T) {
	exe, err := findBlender()
	if err != nil {
		assert.Empty(t, exe)

		if *withBlender {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	assert.NotEmpty(t, exe)
	assert.NotContains(t, exe,
		"blender-launcher",
		"FindBlender should find blender.exe, not blender-launcher.exe")
}
