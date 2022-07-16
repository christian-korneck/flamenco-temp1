package task_logs

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"git.blender.org/flamenco/internal/manager/webupdates"
	"git.blender.org/flamenco/pkg/api"
	"github.com/benbjohnson/clock"
	"github.com/rs/zerolog"
)

const (
	// tailSize is the maximum number of bytes read by the Tail() function.
	tailSize int64 = 2048
)

// Storage can write data to task logs, rotate logs, etc.
type Storage struct {
	localStorage LocalStorage

	clock       clock.Clock
	broadcaster ChangeBroadcaster

	// Locks to only allow one goroutine at a time to handle the logs of a certain task.
	mutex     *sync.Mutex
	taskLocks map[string]*sync.Mutex
}

// Generate mock implementations of these interfaces.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/interfaces_mock.gen.go -package mocks git.blender.org/flamenco/internal/manager/task_logs LocalStorage,ChangeBroadcaster

type LocalStorage interface {
	// ForJob returns the absolute directory path for storing job-related files.
	ForJob(jobUUID string) string
}

type ChangeBroadcaster interface {
	// BroadcastTaskLogUpdate sends the task log update to SocketIO clients.
	BroadcastTaskLogUpdate(taskLogUpdate api.SocketIOTaskLogUpdate)
}

// ChangeBroadcaster should be a subset of webupdates.BiDirComms
var _ ChangeBroadcaster = (*webupdates.BiDirComms)(nil)

// NewStorage creates a new log storage rooted at `basePath`.
func NewStorage(
	localStorage LocalStorage,
	clock clock.Clock,
	broadcaster ChangeBroadcaster,
) *Storage {
	return &Storage{
		localStorage: localStorage,
		clock:        clock,
		broadcaster:  broadcaster,
		mutex:        new(sync.Mutex),
		taskLocks:    make(map[string]*sync.Mutex),
	}
}

// Write appends text to a task's log file, and broadcasts the log lines via SocketIO.
func (s *Storage) Write(logger zerolog.Logger, jobID, taskID string, logText string) error {
	if err := s.writeToDisk(logger, jobID, taskID, logText); err != nil {
		return err
	}

	// Broadcast the task log to SocketIO clients.
	taskUpdate := webupdates.NewTaskLogUpdate(taskID, logText)
	s.broadcaster.BroadcastTaskLogUpdate(taskUpdate)
	return nil
}

// Write appends text, prefixed with the current date & time, to a task's log file,
// and broadcasts the log lines via SocketIO.
func (s *Storage) WriteTimestamped(logger zerolog.Logger, jobID, taskID string, logText string) error {
	now := s.clock.Now().Format(time.RFC3339)
	return s.Write(logger, jobID, taskID, now+" "+logText)
}

func (s *Storage) writeToDisk(logger zerolog.Logger, jobID, taskID string, logText string) error {
	// Shortcut to avoid creating an empty log file. It also solves an
	// index out of bounds error further down when we check the last character.
	if logText == "" {
		return nil
	}

	s.taskLock(taskID)
	defer s.taskUnlock(taskID)

	filepath := s.filepath(jobID, taskID)
	logger = logger.With().Str("filepath", filepath).Logger()

	if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
		logger.Error().Err(err).Msg("unable to create directory for log file")
		return fmt.Errorf("creating directory: %w", err)
	}

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error().Err(err).Msg("unable to open log file for append/create/write")
		return fmt.Errorf("unable to open log file for append/create/write: %w", err)
	}

	if n, err := file.WriteString(logText); n < len(logText) || err != nil {
		logger.Error().
			Int("written", n).
			Int("totalLength", len(logText)).
			Err(err).
			Msg("could only write partial log file")
		file.Close()
		return fmt.Errorf("could only write partial log file: %w", err)
	}

	if logText[len(logText)-1] != '\n' {
		if n, err := file.WriteString("\n"); n < 1 || err != nil {
			logger.Error().Err(err).Msg("could not append line end")
			file.Close()
			return err
		}
	}

	if err := file.Close(); err != nil {
		logger.Error().Err(err).Msg("error closing log file")
		return err
	}
	return nil
}

// RotateFile rotates the task's log file, ignoring (but logging) any errors that occur.
func (s *Storage) RotateFile(logger zerolog.Logger, jobID, taskID string) {
	logpath := s.filepath(jobID, taskID)
	logger = logger.With().Str("logpath", logpath).Logger()

	s.taskLock(taskID)
	defer s.taskUnlock(taskID)

	err := rotateLogFile(logger, logpath)
	if err != nil {
		// rotateLogFile() has already logged something, so we can ignore `err`.
		logger.Trace().Err(err).Msg("ignoring error from log rotation")
	}
}

// filepath returns the file path suitable to write a log file.
// Note that this intentionally shares the behaviour of `pathForJob()` in
// `internal/manager/local_storage/local_storage.go`; it is intended that the
// file handling code in this source file is migrated to use the `local_storage`
// package at some point.
func (s *Storage) filepath(jobID, taskID string) string {
	dirpath := s.localStorage.ForJob(jobID)
	filename := fmt.Sprintf("task-%v.txt", taskID)
	return path.Join(dirpath, filename)
}

// TaskLog reads the entire log file.
func (s *Storage) TaskLog(jobID, taskID string) (string, error) {
	filepath := s.filepath(jobID, taskID)

	s.taskLock(taskID)
	defer s.taskUnlock(taskID)

	buffer, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("reading log file of job %q task %q: %w", jobID, taskID, err)
	}
	return string(buffer), nil
}

// Tail reads the final few lines of a task log.
func (s *Storage) Tail(jobID, taskID string) (string, error) {
	filepath := s.filepath(jobID, taskID)

	s.taskLock(taskID)
	defer s.taskUnlock(taskID)

	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("unable to open log file for reading: %w", err)
	}

	fileSize, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		return "", fmt.Errorf("unable to seek to end of log file: %w", err)
	}

	// Number of bytes to read.
	var (
		buffer   []byte
		numBytes int
	)
	if fileSize <= tailSize {
		// The file is small, just read all of it.
		_, err = file.Seek(0, os.SEEK_SET)
		if err != nil {
			return "", fmt.Errorf("unable to seek to start of log file: %w", err)
		}
		buffer, err = io.ReadAll(file)
	} else {
		// Read the last 'tailSize' number of bytes.
		_, err = file.Seek(-tailSize, os.SEEK_END)
		if err != nil {
			return "", fmt.Errorf("unable to seek in log file: %w", err)
		}
		buffer = make([]byte, tailSize)
		numBytes, err = file.Read(buffer)

		// Try to remove contents up to the first newline, as it's very likely we just
		// seeked into the middle of a line.
		firstNewline := bytes.IndexByte(buffer, byte('\n'))
		if 0 <= firstNewline && firstNewline < numBytes-1 {
			buffer = buffer[firstNewline+1:]
		} else {
			// The file consists of a single line of text; don't strip the first line.
		}
	}
	if err != nil {
		return "", fmt.Errorf("error reading log file: %w", err)
	}

	return string(buffer), nil
}

func (s *Storage) taskLock(taskID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	mutex := s.taskLocks[taskID]
	if mutex == nil {
		mutex = new(sync.Mutex)
		s.taskLocks[taskID] = mutex
	}
	mutex.Lock()
}

func (s *Storage) taskUnlock(taskID string) {
	// This code doesn't modify s.taskLocks, and the task should have been locked
	// already by now.
	mutex := s.taskLocks[taskID]
	if mutex == nil {
		panic("trying to unlock task that is not yet locked")
	}
	mutex.Unlock()
}
