package local_storage

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	"os"
	"path/filepath"

	"git.blender.org/flamenco/pkg/crosspath"
	"github.com/rs/zerolog/log"
)

type StorageInfo struct {
	rootPath string
}

// NewNextToExe returns a storage representation that sits next to the
// currently-running executable. If that directory cannot be determined, falls
// back to the current working directory.
func NewNextToExe(subdir string) StorageInfo {
	exeDir := getSuitableStorageRoot()
	storagePath := filepath.Join(exeDir, subdir)

	return StorageInfo{
		rootPath: storagePath,
	}
}

// ForJob returns the directory path for storing job-related files.
func (si StorageInfo) ForJob(jobUUID string) string {
	return filepath.Join(si.rootPath, pathForJob(jobUUID))
}

// Erase removes the entire storage directory from disk.
func (si StorageInfo) Erase() error {
	// A few safety measures before erasing the planet.
	if si.rootPath == "" {
		return fmt.Errorf("%+v.Erase(): refusing to erase empty directory", si)
	}
	if crosspath.IsRoot(si.rootPath) {
		return fmt.Errorf("%+v.Erase(): refusing to erase root directory", si)
	}
	if home, found := os.LookupEnv("HOME"); found && home == si.rootPath {
		return fmt.Errorf("%+v.Erase(): refusing to erase home directory %s", si, home)
	}

	log.Debug().Str("path", si.rootPath).Msg("erasing storage directory")
	return os.RemoveAll(si.rootPath)
}

// MustErase removes the entire storage directory from disk, and panics if it
// cannot do that. This is primarily aimed at cleaning up at the end of unit
// tests.
func (si StorageInfo) MustErase() {
	err := si.Erase()
	if err != nil {
		panic(err)
	}
}

// Returns a sub-directory suitable for files of this job.
// Note that this is intentionally in sync with the `filepath()` function in
// `internal/manager/task_logs/task_logs.go`.
func pathForJob(jobUUID string) string {
	if jobUUID == "" {
		return "jobless"
	}
	return filepath.Join("job-"+jobUUID[:4], jobUUID)
}

func getSuitableStorageRoot() string {
	exename, err := os.Executable()
	if err == nil {
		return filepath.Dir(exename)
	}
	log.Error().Err(err).Msg("unable to determine the path of the currently running executable")

	// Fall back to current working directory.
	cwd, err := os.Getwd()
	if err == nil {
		return cwd
	}
	log.Error().Err(err).Msg("unable to determine the current working directory")

	// Fall back to "." if all else fails.
	return "."
}
