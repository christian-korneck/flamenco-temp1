package main

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var cliArgs struct {
	// Logging level flags.
	quiet, debug, trace bool

	newVersion     string
	updateMakefile bool
}

func main() {
	parseCliArgs()
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)
	configLogLevel()

	log.Info().Str("version", cliArgs.newVersion).Msg("updating Flamenco version")

	var anyFileWasChanged bool
	if cliArgs.updateMakefile {
		anyFileWasChanged = anyFileWasChanged || updateMakefile()
	}
	anyFileWasChanged = anyFileWasChanged || updateAddon()

	if !anyFileWasChanged {
		log.Warn().Msg("nothing changed")
		os.Exit(42)
		return
	}

	log.Info().Msg("file replacement done")
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.quiet, "quiet", false, "Only log warning-level and worse.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")
	flag.BoolVar(&cliArgs.trace, "trace", false, "Enable trace-level logging.")
	flag.BoolVar(&cliArgs.updateMakefile, "makefile", false,
		"Also update the Makefile. Normally this application is invoked from the Makefile itself, "+
			"and thus it does not change that file without this CLI argument.")

	flag.Parse()

	cliArgs.newVersion = flag.Arg(0)
	if cliArgs.newVersion == "" {
		os.Stderr.WriteString(fmt.Sprintf("Usage: %s [-quiet|-debug|-trace] {new Flamenco version number}\n", os.Args[0]))
		os.Stderr.WriteString("\n")
		flag.PrintDefaults()
		os.Stderr.WriteString("\n")
		os.Stderr.WriteString("This program updates Makefile and some other files to set the new Flamenco version.\n")
		os.Stderr.WriteString("\n")
		os.Exit(47)
	}
}

func configLogLevel() {
	var logLevel zerolog.Level
	switch {
	case cliArgs.trace:
		logLevel = zerolog.TraceLevel
	case cliArgs.debug:
		logLevel = zerolog.DebugLevel
	case cliArgs.quiet:
		logLevel = zerolog.WarnLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
}
