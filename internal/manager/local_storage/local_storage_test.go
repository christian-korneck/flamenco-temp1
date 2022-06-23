package local_storage

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNextToExe(t *testing.T) {
	si := NewNextToExe("nø ASCïÏ")

	// Unit test executables typically are in `/tmp/go-build{random number}`.
	assert.Contains(t, si.rootPath, "go-build")
	assert.Equal(t, filepath.Base(si.rootPath), "nø ASCïÏ",
		"the real path should end in the given directory name")
}

func TestNewNextToExe_noSubdir(t *testing.T) {
	exePath, err := os.Executable()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	exeName := filepath.Base(exePath)

	// The filesystem in an empty "subdirectory" next to the executable should
	// contain the executable.
	si := NewNextToExe("")
	_, err = os.Stat(filepath.Join(si.rootPath, exeName))
	assert.NoErrorf(t, err, "should be able to stat this executable %s", exeName)
}

func TestForJob(t *testing.T) {
	si := NewNextToExe("task-logs")
	jobPath := si.ForJob("08e126ef-d773-468b-8bab-19a8213cf2ff")

	expectedSuffix := filepath.Join("task-logs", "job-08e1", "08e126ef-d773-468b-8bab-19a8213cf2ff")
	hasSuffix := strings.HasSuffix(jobPath, expectedSuffix)
	assert.Truef(t, hasSuffix, "expected %s to have suffix %s", jobPath, expectedSuffix)
}

func TestErase(t *testing.T) {
	si := NewNextToExe("task-logs")
	assert.NoDirExists(t, si.rootPath, "creating a StorageInfo should not create the directory")

	jobPath := si.ForJob("08e126ef-d773-468b-8bab-19a8213cf2ff")
	assert.NoDirExists(t, jobPath, "getting a path should not create it")

	assert.NoError(t, os.MkdirAll(jobPath, os.ModePerm))
	assert.DirExists(t, jobPath, "os.MkdirAll is borked")

	assert.NoError(t, si.Erase())
	assert.NoDirExists(t, si.rootPath, "Erase() should erase the root path, and everything in it")
}
