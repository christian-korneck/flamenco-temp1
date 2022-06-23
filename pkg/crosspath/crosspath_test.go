package crosspath

// SPDX-License-Identifier: GPL-3.0-or-later

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

func TestIsRoot(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"empty", "", false},

		{"UNIX", "/", true},
		{"Drive only", "C:", true},
		{"Drive and slash", "C:/", true},
		{"Drive and backslash", `C:\`, true},

		{"backslash", `\`, false},
		{"just letters", "just letters", false},
		{"subdir of root", "/subdir", false},
		{"subdir of drive", "C:\\subdir", false},

		{"indirectly root", "/subdir/..", false},
		{"UNC notation", `\\NAS\Share\`, false},
		{"Slashed UNC notation", `//NAS/Share/`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRoot(tt.path); got != tt.want {
				t.Errorf("IsRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}
