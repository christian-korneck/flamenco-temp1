// package find_ffmpeg can find an FFmpeg binary on the system.
package find_ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Result struct {
	Path    string
	Version string
}

const (
	// ffmpegExename is the name of the ffmpeg executable. It will be suffixed by
	// the platform-depentent `exeSuffix`
	ffmpegExename = "ffmpeg"

	// toolsDir is the directory sitting next to the currently running executable,
	// in which tools like FFmpeg are searched for.
	toolsDir = "tools"
)

// Find returns the path of an `ffmpeg` executable,
// If there is one next to the currently running executable, that one is used.
// Otherwise $PATH is searched.
func Find() (Result, error) {
	path, err := findBundled()
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// Not finding FFmpeg next to the executable is fine, just continue searching.
	case err != nil:
		// Other errors might be more serious. Log them, but keep going.
		log.Error().Err(err).Msg("error finding FFmpeg next to the current executable")
	case path != "":
		// Found FFmpeg!
		return getVersion(path)
	}

	path, err = exec.LookPath(ffmpegExename)
	if err != nil {
		return Result{}, err
	}
	return getVersion(path)
}

func findBundled() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("finding current executable: %w", err)
	}

	exeDir := filepath.Dir(exe)

	// Subdirectories to use to find the ffmpeg executable. Should go from most to
	// least specific for the current platform.
	filenames := []string{
		fmt.Sprintf("%s-%s-%s%s", ffmpegExename, runtime.GOOS, runtime.GOARCH, exeSuffix),
		fmt.Sprintf("%s-%s%s", ffmpegExename, runtime.GOOS, exeSuffix),
		fmt.Sprintf("%s%s", ffmpegExename, exeSuffix),
	}

	var firstErr error
	for _, filename := range filenames {
		ffmpegPath := filepath.Join(exeDir, toolsDir, filename)
		_, err = os.Stat(ffmpegPath)

		switch {
		case err == nil:
			return ffmpegPath, nil
		case errors.Is(err, fs.ErrNotExist):
			log.Debug().Str("path", ffmpegPath).Msg("FFmpeg not found on this path")
		default:
			log.Debug().Err(err).Str("path", ffmpegPath).Msg("FFmpeg could not be accessed on this path")
		}

		// If every path fails, report on the first-failed path, as that's the most
		// specific one.
		if firstErr == nil {
			firstErr = err
		}
	}

	return "", firstErr
}

func getVersion(path string) (Result, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	cmd := exec.CommandContext(ctx, path, "-version")
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return Result{}, fmt.Errorf("running %s -version: %w", path, err)
	}
	output := string(outputBytes)

	lines := strings.SplitN(output, "\n", 2)
	if len(lines) < 2 {
		return Result{}, fmt.Errorf("unexpected output (only %d lines) from %s -version: %s", len(lines), path, output)
	}
	versionLine := lines[0]

	// Get the version from the first line of output.
	// ffmpeg version 4.2.7-0ubuntu0.1 Copyright (c) 2000-2022 the FFmpeg developer
	if !strings.HasPrefix(versionLine, "ffmpeg version ") {
		return Result{}, fmt.Errorf("unexpected output from %s -version: [%s]", path, versionLine)
	}
	lineParts := strings.SplitN(versionLine, " ", 4)

	return Result{
		Path:    path,
		Version: lineParts[2],
	}, nil
}
