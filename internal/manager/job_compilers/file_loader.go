package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

var (
	// embeddedScriptsFS gives access to the embedded scripts.
	embeddedScriptsFS fs.FS

	// onDiskScriptsFS gives access to the on-disk scripts, located in a `scripts`
	// directory next to the `flamenco-manager` executable.
	onDiskScriptsFS fs.FS = nil

	fileLoaderInitialised = false
)

const scriptsDirName = "scripts"

// Scripts from the `./scripts` subdirectory are embedded into the executable
// here. Note that accessing these files still requires explicit use of the
// `scripts/` subdirectory, which is abstracted away by `embeddedScriptFS`.
//
//go:embed scripts
var _embeddedScriptsFS embed.FS

func initFileLoader() {
	if fileLoaderInitialised {
		return
	}

	initEmbeddedFS()
	initOnDiskFS()

	fileLoaderInitialised = true
}

// getAvailableFilesystems returns the filesystems to load scripts from, where
// earlier ones have priority over later ones.
func getAvailableFilesystems() []fs.FS {
	filesystems := []fs.FS{}

	if onDiskScriptsFS != nil {
		filesystems = append(filesystems, onDiskScriptsFS)
	}

	filesystems = append(filesystems, embeddedScriptsFS)
	return filesystems
}

// loadFileFromAnyFS iterates over the available filesystems to find the
// identified file, and returns its contents when found.
//
// Returns `os.ErrNotExist` if there is no filesystem that has this file.
func loadFileFromAnyFS(path string) ([]byte, error) {
	filesystems := getAvailableFilesystems()

	for _, fs := range filesystems {
		file, err := fs.Open(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s on filesystem %s: %w", path, fs, err)
		}
		return io.ReadAll(file)
	}

	return nil, os.ErrNotExist
}

func initEmbeddedFS() {
	// Find embedded filesystem. Unless there were issues with the build of
	// Flamenco Manager, this should always be here.
	var err error
	embeddedScriptsFS, err = fs.Sub(_embeddedScriptsFS, "scripts")
	if err != nil {
		panic(fmt.Sprintf("failed to find embedded 'scripts' directory: %v", err))
	}
}

func initOnDiskFS() {
	exename, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("job compiler: unable to determine the path of the currently running executable")
		return
	}
	logger := log.With().Str("executable", exename).Logger()
	logger.Debug().Msg("job compiler: searching for scripts directory next to executable")

	// Try to find the scripts next to the executable.
	scriptsDir, found := findOnDiskScriptsNextTo(exename)
	if found {
		log.Debug().Str("scriptsDir", scriptsDir).Msg("job compiler: found scripts directory next to executable")
		onDiskScriptsFS = os.DirFS(scriptsDir)
		return
	}

	// Evaluate any symlinks and see if that produces a different path to the
	// executable.
	evalLinkExe, err := filepath.EvalSymlinks(exename)
	if err != nil {
		logger.Error().Err(err).Msg("job compiler: unable to evaluate any symlinks to the running executable")
		return
	}
	if evalLinkExe == exename {
		// Evaluating any symlinks didn't produce a different path; no need to do the same search twice.
		return
	}

	scriptsDir, found = findOnDiskScriptsNextTo(evalLinkExe)
	if !found {
		logger.Debug().Msg("job compiler: did not find scripts directory next to executable")
		return
	}

	log.Debug().Str("scriptsDir", scriptsDir).Msg("job compiler: found scripts directory next to executable")
	onDiskScriptsFS = os.DirFS(scriptsDir)
}

// Find the `scripts` directory sitting next to the currently-running executable.
// Return the directory path, and a 'found' boolean indicating whether that path
// is actually a directory.
func findOnDiskScriptsNextTo(exename string) (string, bool) {
	scriptsDir := filepath.Join(filepath.Dir(exename), scriptsDirName)

	logger := log.With().Str("scriptsDir", scriptsDir).Logger()
	logger.Trace().Msg("job compiler: finding on-disk scripts")

	stat, err := os.Stat(scriptsDir)
	if os.IsNotExist(err) {
		return scriptsDir, false
	}
	if err != nil {
		logger.Warn().Err(err).Msg("job compiler: error accessing scripts directory")
		return scriptsDir, false
	}
	if !stat.IsDir() {
		logger.Debug().Msg("job compiler: ignoring 'scripts' next to executable; it is not a directory")
		return scriptsDir, false
	}

	return scriptsDir, true
}
