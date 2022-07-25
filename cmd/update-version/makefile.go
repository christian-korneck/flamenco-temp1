package main

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

const makefileFile = "Makefile"

// updateMakefile changes the version number in Makefile.
// Returns whether the file actually changed.
func updateMakefile() bool {
	replacer := func(line string) string {
		if !strings.HasPrefix(line, "VERSION := ") {
			return line
		}
		return fmt.Sprintf("VERSION := %q", cliArgs.newVersion)
	}

	fileWasChanged, err := updateLines(makefileFile, replacer)
	if err != nil {
		log.Fatal().Err(err).Msg("error updating Makefile")
	}
	return fileWasChanged
}
