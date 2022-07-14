package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"flag"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

var withBlender = flag.Bool("withBlender", false, "run test that requires Blender to be installed")

func TestGetBlenderVersion(t *testing.T) {
	if !*withBlender {
		t.Skip("skipping test, -withBlender arg not passed")
	}

	path, err := exec.LookPath("blender")
	if !assert.NoError(t, err) {
		t.Fatal("running with -withBlender requires having a `blender` command on $PATH")
	}

	ctx := context.Background()

	// Try finding version from "/path/to/blender":
	version, err := getBlenderVersion(ctx, path)
	if assert.NoError(t, err) {
		assert.Contains(t, version, "Blender")
	}

	// Try non-existing executable:
	version, err = getBlenderVersion(ctx, "This-Blender-Executable-Does-Not-Exist")
	assert.ErrorIs(t, err, exec.ErrNotFound)
	assert.Empty(t, version)
}
