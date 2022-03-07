package task_logs

// SPDX-License-Identifier: GPL-3.0-or-later

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
