package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

import "errors"

var ErrNotAvailable = errors.New("not available on this platform")

// FindBlender returns the full path of a Blender executable.
// `ErrNotAvailable` is returned if no "blender finder" is available for the current platform.
func FindBlender() (string, error) {
	// findBlender() is implemented in one of the platform-dependent files.
	return findBlender()
}
