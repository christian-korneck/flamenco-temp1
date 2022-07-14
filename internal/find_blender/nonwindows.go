//go:build !windows

package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

// fileAssociation isn't implemented on non-Windows platforms.
func fileAssociation() (string, error) {
	return "", ErrNotAvailable
}
