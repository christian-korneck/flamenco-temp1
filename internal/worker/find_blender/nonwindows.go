//go:build !windows

// SPDX-License-Identifier: GPL-3.0-or-later
package find_blender

import "errors"

// FindBlender returns an error, as the file association lookup is only available on Windows.
func FindBlender() (string, error) {
	return "", errors.New("file association lookup is only available on Windows")
}
