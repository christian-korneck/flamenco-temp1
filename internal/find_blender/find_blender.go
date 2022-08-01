package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/crosspath"
	"github.com/rs/zerolog/log"
)

var (
	ErrNotAvailable        = errors.New("not available on this platform")
	ErrAssociationNotFound = errors.New("no program is associated with .blend files")
	ErrNotBlender          = errors.New("not a Blender executable")
	ErrTimedOut            = errors.New("version check took too long")
)

// blenderVersionTimeout is how long `blender --version` is allowed to take,
// before timing out. This can be much slower than expected, when loading
// Blender from shared storage on a not-so-fast NAS.
const blenderVersionTimeout = 10 * time.Second

type CheckBlenderResult struct {
	Input          string // What was the original 'exename' CheckBlender was told to find.
	FoundLocation  string
	BlenderVersion string
	Source         api.BlenderPathSource
}

// Find returns the path of a `blender` executable,
// If there is one associated with .blend files, and the current platform is
// supported to query those, that one is used. Otherwise $PATH is searched.
func Find(ctx context.Context) (CheckBlenderResult, error) {
	return CheckBlender(ctx, "")
}

// FileAssociation returns the full path of a Blender executable, by inspecting file association with .blend files.
// `ErrNotAvailable` is returned if no "blender finder" is available for the current platform.
func FileAssociation() (string, error) {
	// findBlender() is implemented in one of the platform-dependent files.
	return fileAssociation()
}

func CheckBlender(ctx context.Context, exename string) (CheckBlenderResult, error) {
	if exename == "" {
		// exename is not given, see if we can use .blend file association.
		fullPath, err := fileAssociation()
		switch {
		case errors.Is(err, ErrNotAvailable):
			// Association finder not available, act as if "blender" was given as exename.
			return CheckBlender(ctx, "blender")
		case err != nil:
			// Some other error occurred, better to report it.
			return CheckBlenderResult{}, fmt.Errorf("error finding .blend file association: %w", err)
		default:
			// The full path was found, report the Blender version.
			return getResultWithVersion(ctx, exename, fullPath, api.BlenderPathSourceFileAssociation)
		}
	}

	if crosspath.Dir(exename) != "." {
		// exename is some form of path, see if it works for us.
		return checkBlenderAtPath(ctx, exename)
	}

	// Try to find exename on $PATH
	fullPath, err := exec.LookPath(exename)
	if err != nil {
		return CheckBlenderResult{}, err
	}
	return getResultWithVersion(ctx, exename, fullPath, api.BlenderPathSourcePathEnvvar)
}

func checkBlenderAtPath(ctx context.Context, path string) (CheckBlenderResult, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return CheckBlenderResult{}, err
	}
	if !stat.IsDir() {
		// Simple case, it's not a directory so let's just try and execute it.
		return getResultWithVersion(ctx, path, path, api.BlenderPathSourceInputPath)
	}

	// Try appending the Blender executable name.
	log.Debug().
		Str("path", path).
		Str("executable", blenderExeName).
		Msg("blender finder: given path is directory, going to find Blender executable")
	exepath := filepath.Join(path, blenderExeName)
	return getResultWithVersion(ctx, path, exepath, api.BlenderPathSourceInputPath)
}

// getResultWithVersion tries to run the command to get Blender's version.
// The result is returned as a `CheckBlenderResult` struct.
func getResultWithVersion(
	ctx context.Context,
	input,
	commandline string,
	source api.BlenderPathSource,
) (CheckBlenderResult, error) {
	result := CheckBlenderResult{
		Input:         input,
		FoundLocation: commandline,
		Source:        source,
	}

	version, err := getBlenderVersion(ctx, commandline)
	if err != nil {
		return result, err
	}

	result.BlenderVersion = version
	return result, nil
}

func getBlenderVersion(ctx context.Context, commandline string) (string, error) {
	logger := log.With().Str("commandline", commandline).Logger()

	// Make sure that command execution doesn't hang indefinitely.
	cmdCtx, cmdCtxCancel := context.WithTimeout(ctx, blenderVersionTimeout)
	defer cmdCtxCancel()

	cmd := exec.CommandContext(cmdCtx, commandline, "--version")
	stdoutStderr, err := cmd.CombinedOutput()
	switch {
	case errors.Is(cmdCtx.Err(), context.DeadlineExceeded):
		logger.Warn().Stringer("timeout", blenderVersionTimeout).Msg("command timed out")
		return "", fmt.Errorf("%s: %w", commandline, ErrTimedOut)
	case err != nil:
		logger.Info().Err(err).Str("output", string(stdoutStderr)).Msg("error running command")
		return "", err
	}

	version := string(stdoutStderr)
	lines := strings.Split(version, "\n")
	for idx := range lines {
		line := strings.TrimSpace(lines[idx])
		if strings.HasPrefix(line, "Blender ") {
			return line, nil
		}
	}
	return "", fmt.Errorf("%s: %w", commandline, ErrNotBlender)
}
