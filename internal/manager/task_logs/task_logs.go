package task_logs

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// tailSize is the maximum number of bytes read by the Tail() function.
	tailSize int64 = 2048
)

// Storage can write data to task logs, rotate logs, etc.
type Storage struct {
	BasePath string // Directory where task logs are stored.
}

// NewStorage creates a new log storage rooted at `basePath`.
func NewStorage(basePath string) *Storage {
	if !filepath.IsAbs(basePath) {
		absPath, err := filepath.Abs(basePath)
		if err != nil {
			log.Panic().Err(err).Str("path", basePath).Msg("cannot resolve relative path to task logs")
		}
		basePath = absPath
	}

	log.Info().
		Str("path", basePath).
		Msg("task logs")

	return &Storage{
		BasePath: basePath,
	}
}

func (s *Storage) Write(logger zerolog.Logger, jobID, taskID string, logText string) error {
	// Shortcut to avoid creating an empty log file. It also solves an
	// index out of bounds error further down when we check the last character.
	if logText == "" {
		return nil
	}

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

	err := rotateLogFile(logger, logpath)
	if err != nil {
		// rotateLogFile() has already logged something, so we can ignore `err`.
		logger.Trace().Err(err).Msg("ignoring error from log rotation")
	}
}

// filepath returns the file path suitable to write a log file.
func (s *Storage) filepath(jobID, taskID string) string {
	var dirpath string
	if jobID == "" {
		dirpath = path.Join(s.BasePath, "jobless")
	} else {
		dirpath = path.Join(s.BasePath, "job-"+jobID[:4], jobID)
	}
	filename := fmt.Sprintf("task-%v.txt", taskID)
	return path.Join(dirpath, filename)
}

// Tail reads the final few lines of a task log.
func (s *Storage) Tail(jobID, taskID string) (string, error) {
	filepath := s.filepath(jobID, taskID)

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
