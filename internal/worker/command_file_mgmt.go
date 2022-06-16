package worker

// SPDX-License-Identifier: GPL-3.0-or-later

/* This file contains the commands in the "file-management" type group. */

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"git.blender.org/flamenco/pkg/api"
)

// cmdMoveDirectory executes the "move-directory" command.
// It moves directory 'src' to 'dest'; if 'dest' already exists, it's moved to 'dest-{timestamp}'.
func (ce *CommandExecutor) cmdMoveDirectory(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {
	var src, dest string
	var ok bool

	if src, ok = cmdParameter[string](cmd, "src"); !ok || src == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'src' parameter")
		return NewParameterMissingError("src", cmd)
	}
	if dest, ok = cmdParameter[string](cmd, "dest"); !ok || dest == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'dest' parameter")
		return NewParameterMissingError("dest", cmd)
	}

	logger = logger.With().
		Str("src", src).
		Str("dest", dest).
		Logger()
	if !fileExists(src) {
		logger.Warn().Msg("source path does not exist, not moving anything")
		msg := fmt.Sprintf("%s: source path %q does not exist, not moving anything", cmd.Name, src)
		if err := ce.listener.LogProduced(ctx, taskID, msg); err != nil {
			return err
		}
		return NewParameterInvalidError("src", cmd, "path does not exist")
	}

	if fileExists(dest) {
		backup, err := timestampedPath(dest)
		if err != nil {
			logger.Error().Err(err).Str("path", dest).Msg("unable to determine timestamp of directory")
			return err
		}

		if fileExists(backup) {
			logger.Debug().Str("backup", backup).Msg("backup destination also exists, finding one that does not")
			backup, err = uniquePath(backup)
			if err != nil {
				return err
			}
		}

		logger.Info().
			Str("toBackup", backup).
			Msg("dest directory exists, moving to backup")
		if err := ce.moveAndLog(ctx, taskID, cmd.Name, dest, backup); err != nil {
			return err
		}
	}

	// self._log.info("Moving %s to %s", src, dest)
	// await self.worker.register_log(
	// 		"%s: Moving %s to %s", self.command_name, src, dest
	// )
	// src.rename(dest)
	logger.Info().Msg("moving directory")
	return ce.moveAndLog(ctx, taskID, cmd.Name, src, dest)
}

// moveAndLog renames a file/directory from `src` to `dest`, and logs the moveAndLog.
// The other parameters are just for logging.
func (ce *CommandExecutor) moveAndLog(ctx context.Context, taskID, cmdName, src, dest string) error {
	msg := fmt.Sprintf("%s: moving %q to %q", cmdName, src, dest)
	if err := ce.listener.LogProduced(ctx, taskID, msg); err != nil {
		return err
	}

	if err := os.Rename(src, dest); err != nil {
		msg := fmt.Sprintf("%s: could not move %q to %q: %v", cmdName, src, dest, err)
		if err := ce.listener.LogProduced(ctx, taskID, msg); err != nil {
			return err
		}
		return err
	}

	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// timestampedPath returns the path with its modification time appended to the name.
func timestampedPath(filepath string) (string, error) {
	stat, err := os.Stat(filepath)
	if err != nil {
		return "", fmt.Errorf("getting mtime of %s: %w", filepath, err)
	}

	// Round away the milliseconds, as those aren't all that interesting.
	// Uniqueness can ensured by calling unique_path() later.
	mtime := stat.ModTime().Round(time.Second)

	iso := mtime.Local().Format("2006-01-02_150405") // YYYY-MM-DD_HHMMSS
	return fmt.Sprintf("%s-%s", filepath, iso), nil
}

// uniquePath returns the path, or if it exists, the path with a unique suffix.
func uniquePath(path string) (string, error) {
	matches, err := filepath.Glob(path + "-*")
	if err != nil {
		return "", err
	}

	suffixRe, err := regexp.Compile("-([0-9]+)$")
	if err != nil {
		return "", fmt.Errorf("compiling regular expression: %w", err)
	}

	var maxSuffix int64
	for _, path := range matches {
		matches := suffixRe.FindStringSubmatch(path)
		if len(matches) < 2 {
			continue
		}
		suffix := matches[1]
		value, err := strconv.ParseInt(suffix, 10, 64)
		if err != nil {
			// Non-numeric suffixes are fine; they just don't count for this function.
			continue
		}

		if value > maxSuffix {
			maxSuffix = value
		}
	}

	return fmt.Sprintf("%s-%03d", path, maxSuffix+1), nil
}
