//go:build !windows

package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

// findBlender returns ErrNotAvailable, as the file association lookup is only available on Windows.
func findBlender() (string, error) {
	return "", ErrNotAvailable
}
