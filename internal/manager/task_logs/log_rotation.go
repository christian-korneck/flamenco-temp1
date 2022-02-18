package task_logs

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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

type numberedPath struct {
	path     string
	number   int
	basepath string
}

// byNumber implements the sort.Interface for numberedPath objects,
// and sorts in reverse (so highest number first).
type byNumber []numberedPath

func (a byNumber) Len() int           { return len(a) }
func (a byNumber) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byNumber) Less(i, j int) bool { return a[i].number > a[j].number }

func createNumberedPath(path string) numberedPath {
	dotIndex := strings.LastIndex(path, ".")
	if dotIndex < 0 {
		return numberedPath{path, -1, path}
	}
	asInt, err := strconv.Atoi(path[dotIndex+1:])
	if err != nil {
		return numberedPath{path, -1, path}
	}
	return numberedPath{path, asInt, path[:dotIndex]}
}

// rotateLogFile renames 'logpath' to 'logpath.1', and increases numbers for already-existing files.
// NOTE: not thread-safe when calling with the same `logpath`.
func rotateLogFile(logger zerolog.Logger, logpath string) error {
	// Don't do anything if the file doesn't exist yet.
	_, err := os.Stat(logpath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debug().Msg("log file does not exist, no need to rotate")
			return nil
		}
		logger.Warn().Err(err).Msg("unable to stat logfile")
		return err
	}

	// Rotate logpath.3 to logpath.2, logpath.1 to logpath.2, etc.
	pattern := logpath + ".*"
	existing, err := filepath.Glob(pattern)
	if err != nil {
		logger.Warn().Err(err).Str("glob", pattern).Msg("rotateLogFile: unable to glob")
		return err
	}
	if existing == nil {
		logger.Debug().Msg("rotateLogFile: no existing files to rotate")
	} else {
		// Rotate the files in reverse numerical order (so renaming n→n+1 comes after n+1→n+2)
		var numbered = make(byNumber, len(existing))
		for idx := range existing {
			numbered[idx] = createNumberedPath(existing[idx])
		}
		sort.Sort(numbered)

		for _, numberedPath := range numbered {
			newName := numberedPath.basepath + "." + strconv.Itoa(numberedPath.number+1)
			err := os.Rename(numberedPath.path, newName)
			if err != nil {
				logger.Error().
					Str("from_path", numberedPath.path).
					Str("to_path", newName).
					Err(err).
					Msg("rotateLogFile: unable to rename log file")
			}
		}
	}

	// Rotate the pointed-to file.
	newName := logpath + ".1"
	if err := os.Rename(logpath, newName); err != nil {
		logger.Error().Str("new_name", newName).Err(err).Msg("rotateLogFile: unable to rename log file for rotating")
		return err
	}

	return nil
}
