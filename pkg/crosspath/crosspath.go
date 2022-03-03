// Package crosspath deals with file/directory paths in a cross-platform way.
//
// This package tries to understand Windows paths on UNIX and vice versa.
// Returned paths may be using forward slashes as separators.
package crosspath

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

import (
	path_module "path" // import under other name so that parameters can be called 'path'
	"strings"
)

// Base returns the last element of path. Trailing slashes are removed before
// extracting the last element. If the path is empty, Base returns ".". If the
// path consists entirely of slashes, Base returns "/".
func Base(path string) string {
	slashed := ToSlash(path)
	return path_module.Base(slashed)
}

// Dir returns all but the last element of path, typically the path's directory.
// If the path is empty, Dir returns ".".
func Dir(path string) string {
	if path == "" {
		return "."
	}

	slashed := ToSlash(path)

	// Don't use path.Dir(), as that cleans up the path and removes double
	// slashes. However, Windows UNC paths start with double blackslashes, which
	// will translate to double slashes and should not be removed.
	dir, _ := path_module.Split(slashed)
	switch {
	case dir == "":
		return "."
	case len(dir) > 1:
		// Remove trailing slash.
		return dir[:len(dir)-1]
	default:
		return dir
	}
}

func Join(elem ...string) string {
	return ToSlash(path_module.Join(elem...))
}

// Stem returns the filename without extension.
func Stem(path string) string {
	base := Base(path)
	ext := path_module.Ext(base)
	return base[:len(base)-len(ext)]
}

// ToSlash replaces all backslashes with forward slashes.
// Contrary to filepath.ToSlash(), this also happens on Linux; it does not
// expect `path` to be in platform-native notation.
func ToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
