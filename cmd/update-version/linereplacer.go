package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

// updateLines calls replacer() on each line of the given file, and replaces it
// with the returned value.
// Returns whether the file changed at all.
func updateLines(filename string, replacer func(string) string) (bool, error) {
	logger := log.With().Str("filename", filename).Logger()
	logger.Info().Msg("updating file")

	// Read the file contents:
	input, err := os.ReadFile(filename)
	if err != nil {
		return false, fmt.Errorf("reading from %s: %w", filename, err)
	}

	// Replace the lines:
	anythingChanged := false
	lines := strings.Split(string(input), "\n")
	for idx := range lines {
		replaced := replacer(lines[idx])
		if replaced == lines[idx] {
			continue
		}

		logger.Info().
			Str("old", strings.TrimSpace(lines[idx])).
			Str("new", strings.TrimSpace(replaced)).
			Msg("replacing line")
		lines[idx] = replaced
		anythingChanged = true
	}

	if !anythingChanged {
		logger.Info().Msg("file did not change, will not touch it")
		return false, nil
	}

	// Write the file contents to a temporary location:
	output := strings.Join(lines, "\n")
	tempname := filename + "~"
	err = os.WriteFile(tempname, []byte(output), 0644)
	if err != nil {
		return false, fmt.Errorf("writing to %s: %w", tempname, err)
	}

	// Move the temporary file onto the input filename:
	if err := os.Remove(filename); err != nil {
		return false, fmt.Errorf("removing %s: %w", filename, err)
	}
	if err := os.Rename(tempname, filename); err != nil {
		return false, fmt.Errorf("renaming %s to %s: %w", tempname, filename, err)
	}

	return true, nil
}
