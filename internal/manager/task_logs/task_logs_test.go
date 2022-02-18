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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func tempStorage() *Storage {
	temppath, err := ioutil.TempDir("", "testlogs")
	if err != nil {
		panic(err)
	}
	return &Storage{temppath}
}

func TestLogWriting(t *testing.T) {
	s := tempStorage()
	defer os.RemoveAll(s.BasePath)

	err := s.Write(zerolog.Nop(),
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"20ff9d06-53ec-4019-9e2e-1774f05f170a",
		"Ovo je priča")
	assert.NoError(t, err)

	err = s.Write(zerolog.Nop(),
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"20ff9d06-53ec-4019-9e2e-1774f05f170a",
		"Ima dvije linije")
	assert.NoError(t, err)

	filename := filepath.Join(
		s.BasePath,
		"job-25c5",
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"task-20ff9d06-53ec-4019-9e2e-1774f05f170a.txt")

	contents, err := ioutil.ReadFile(filename)
	assert.NoError(t, err, "the log file should exist")
	assert.Equal(t, "Ovo je priča\nIma dvije linije\n", string(contents))
}

func TestLogRotation(t *testing.T) {
	s := tempStorage()
	defer os.RemoveAll(s.BasePath)

	err := s.Write(zerolog.Nop(),
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"20ff9d06-53ec-4019-9e2e-1774f05f170a",
		"Ovo je priča")
	assert.NoError(t, err)

	s.RotateFile(zerolog.Nop(),
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"20ff9d06-53ec-4019-9e2e-1774f05f170a")

	filename := filepath.Join(
		s.BasePath,
		"job-25c5",
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"task-20ff9d06-53ec-4019-9e2e-1774f05f170a.txt")
	rotatedFilename := filename + ".1"

	contents, err := ioutil.ReadFile(rotatedFilename)
	assert.NoError(t, err, "the rotated log file should exist")
	assert.Equal(t, "Ovo je priča\n", string(contents))

	_, err = os.Stat(filename)
	assert.True(t, os.IsNotExist(err))
}
