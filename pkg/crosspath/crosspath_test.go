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
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	tests := []struct {
		expect, input string
	}{
		{".", ""},
		{"justafile.txt", "justafile.txt"},
		{"with spaces.txt", "/Linux path/with spaces.txt"},
		{"awésom.tar.gz", "C:\\ünicode\\is\\awésom.tar.gz"},
		{"Resource with ext.ension", "\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension"},
	}
	for _, test := range tests {
		assert.Equal(t, test.expect, Base(test.input))
	}
}

func TestDir(t *testing.T) {
	// Just to show how path.Dir() behaves:
	assert.Equal(t, ".", path.Dir(""))
	assert.Equal(t, ".", path.Dir("justafile.txt"))

	tests := []struct {
		expect, input string
	}{
		// Follow path.Dir() when it comes to empty directories:
		{".", ""},
		{".", "justafile.txt"},

		{"/", "/"},
		{"/", "/file-at-root"},
		{"C:", "C:\\file-at-root"},
		{"/Linux path", "/Linux path/with spaces.txt"},
		{"/Mixed path/with", "/Mixed path\\with/slash.txt"},
		{"C:/ünicode/is", "C:\\ünicode\\is\\awésom.tar.gz"},
		{"//SERVER/ünicode/is", "\\\\SERVER\\ünicode\\is\\awésom.tar.gz"},
		{"//?/UNC/ComputerName/SharedFolder", "\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource"},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, Dir(test.input),
			"for input %q", test.input)
	}
}

func TestJoin(t *testing.T) {
	// Just to show how path.Join() behaves:
	assert.Equal(t, "", path.Join())
	assert.Equal(t, "", path.Join(""))
	assert.Equal(t, "", path.Join("", ""))
	assert.Equal(t, "a/b", path.Join("", "", "a", "", "b", ""))

	tests := []struct {
		expect string
		input  []string
	}{
		// Should behave the same as path.Join():
		{"", []string{}},
		{"", []string{""}},
		{"", []string{"", ""}},
		{"a/b", []string{"", "", "a", "", "b", ""}},

		{"/file-at-root", []string{"/", "file-at-root"}},
		{"C:/file-at-root", []string{"C:", "file-at-root"}},

		{"/Linux path/with spaces.txt", []string{"/Linux path", "with spaces.txt"}},
		{"C:/ünicode/is/awésom.tar.gz", []string{"C:\\ünicode", "is\\awésom.tar.gz"}},
		{"//SERVER/mount/dir/file.txt", []string{"\\\\SERVER", "mount", "dir", "file.txt"}},
		{"//?/UNC/ComputerName/SharedFolder/Resource", []string{"\\\\?\\UNC\\ComputerName", "SharedFolder\\Resource"}},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, Join(test.input...),
			"for input %q", test.input)
	}
}

func TestStem(t *testing.T) {
	tests := []struct {
		expect, input string
	}{
		{"", ""},
		{"stem", "stem.txt"},
		{"stem.tar", "stem.tar.gz"},
		{"file", "/path/to/file.txt"},
		{"file", "C:\\path\\to\\file.txt"},
		{"file", "C:\\path\\to/mixed/slashes/file.txt"},
		{"file", "C:\\path/to\\mixed/slashes\\file.txt"},
		{"Resource with ext", "\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension"},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, Stem(test.input),
			"for input %q", test.input)
	}
}

func TestToNative_native_backslash(t *testing.T) {
	if filepath.Separator != '\\' {
		t.Skipf("skipping backslash-specific test on %q with path separator %q",
			runtime.GOOS, filepath.Separator)
	}

	tests := []struct {
		expect, input string
	}{
		{"", ""},
		{".", "."},
		{"\\some\\simple\\path", "/some/simple/path"},
		{"C:\\path\\to\\file.txt", "C:\\path\\to\\file.txt"},
		{"C:\\path\\to\\mixed\\slashes\\file.txt", "C:\\path\\to/mixed/slashes/file.txt"},
		{"\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension",
			"\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension"},
		{"\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension",
			"//?/UNC/ComputerName/SharedFolder/Resource with ext.ension"},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, ToNative(test.input),
			"for input %q", test.input)
	}
}

func TestToNative_native_slash(t *testing.T) {
	if filepath.Separator != '/' {
		t.Skipf("skipping backslash-specific test on %q with path separator %q",
			runtime.GOOS, filepath.Separator)
	}

	tests := []struct {
		expect, input string
	}{
		{"", ""},
		{".", "."},
		{"/some/simple/path", "/some/simple/path"},
		{"C:/path/to/file.txt", "C:\\path\\to\\file.txt"},
		{"C:/path/to/mixed/slashes/file.txt", "C:\\path\\to/mixed/slashes/file.txt"},
		{"//?/UNC/ComputerName/SharedFolder/Resource with ext.ension",
			"\\\\?\\UNC\\ComputerName\\SharedFolder\\Resource with ext.ension"},
		{"//?/UNC/ComputerName/SharedFolder/Resource with ext.ension",
			"//?/UNC/ComputerName/SharedFolder/Resource with ext.ension"},
	}
	for _, test := range tests {
		assert.Equal(t,
			test.expect, ToNative(test.input),
			"for input %q", test.input)
	}
}

// This test should be skipped on every platform. It's there just to detect that
// the above two tests haven't run.
func TestToNative_unsupported(t *testing.T) {
	if filepath.Separator == '/' || filepath.Separator == '\\' {
		t.Skipf("skipping test on %q with path separator %q",
			runtime.GOOS, filepath.Separator)
	}

	t.Fatalf("ToNative not supported on this platform %q with path separator %q",
		runtime.GOOS, filepath.Separator)
}
