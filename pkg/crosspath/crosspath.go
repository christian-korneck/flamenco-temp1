// Package crosspath deals with file/directory paths in a cross-platform way.
//
// This package tries to understand Windows paths on UNIX and vice versa.
// Returned paths may be using forward slashes as separators.
package crosspath

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	path_module "path" // import under other name so that parameters can be called 'path'
	"path/filepath"
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

// ToNative replaces all path separators (forward and backward slashes) with the
// platform-native separator.
func ToNative(path string) string {
	switch filepath.Separator {
	case '/':
		return ToSlash(path)
	case '\\':
		return strings.ReplaceAll(path, "/", "\\")
	default:
		panic(fmt.Sprintf("this platform has an unknown path separator: %q", filepath.Separator))
	}
}

func validDriveLetter(r rune) bool {
	return ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z')
}

// IsRoot returns whether the given path is a root path or not.
// Paths "/", "C:", "C:\", and "C:/" are considered root, for all latin drive
// letters A-Z.
//
// NOTE: this does NOT resolve symlinks or `..` entries.
func IsRoot(path string) bool {
	switch {
	case path == "":
		return false
	case path == "/":
		return true
	// From here on, it can only be a DOS root, so must have a drive letter and a colon.
	case len(path) < 2, len(path) > 3:
		return false
	case path[1] != ':':
		return false
	// C:\ and C:/ are both considered roots.
	case len(path) == 3 && path[2] != '/' && path[2] != '\\':
		return false
	}

	runes := []rune(path)
	return validDriveLetter(runes[0])
}
