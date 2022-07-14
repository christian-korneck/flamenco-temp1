package find_blender

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/crosspath"
	"github.com/rs/zerolog/log"
)

var ErrNotAvailable = errors.New("not available on this platform")

type CheckBlenderResult struct {
	Input          string // What was the original 'exename' CheckBlender was told to find.
	FoundLocation  string
	BlenderVersion string
	Source         api.BlenderPathSource
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
		case err == ErrNotAvailable:
			// Association finder not available, act as if "blender" was given as exename.
			return CheckBlender(ctx, "blender")
		case err != nil:
			// Some other error occurred, better to report it.
			return CheckBlenderResult{}, err
		default:
			// The full path was found, report the Blender version.
			return getResultWithVersion(ctx, exename, fullPath, api.BlenderPathSourceFileAssociation)
		}
	}

	if crosspath.Dir(exename) != "." {
		// exename is some form of path, see if it works directly as executable.
		return getResultWithVersion(ctx, exename, exename, api.BlenderPathSourceInputPath)
	}

	// Try to find exename on $PATH
	fullPath, err := exec.LookPath(exename)
	if err != nil {
		return CheckBlenderResult{}, err
	}
	return getResultWithVersion(ctx, exename, fullPath, api.BlenderPathSourcePathEnvvar)
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

	cmd := exec.CommandContext(ctx, commandline, "--version")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logger.Info().Err(err).Str("output", string(stdoutStderr)).Msg("error running command")
		return "", err
	}

	version := strings.TrimSpace(string(stdoutStderr))
	return version, nil
}
