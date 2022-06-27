package main

import (
	"archive/zip"
	"compress/flate"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/appinfo"
)

var cliArgs struct {
	// Do-and-quit flags.
	version bool

	// Logging level flags.
	quiet, debug, trace bool

	filename string
}

func main() {
	parseCliArgs()
	if cliArgs.version {
		fmt.Println(appinfo.ApplicationVersion)
		return
	}

	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)
	configLogLevel()

	outfile, err := filepath.Abs(cliArgs.filename)
	if err != nil {
		log.Fatal().Err(err).Str("filepath", cliArgs.filename).Msg("unable make output file path absolute")
	}

	// Open the output file.
	logger := log.With().Str("zipname", outfile).Logger()
	logger.Info().Msg("creating ZIP file")
	zipFile, err := os.Create(outfile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error creating file")
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)

	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	// CD to the addon/flamenco dir.
	if err := os.Chdir("addon/flamenco"); err != nil {
		log.Fatal().Err(err).Msg("unable to cd to addon/flamenco")
	}

	basePath, err := os.Getwd()
	if err != nil {
		logger.Fatal().Err(err).Msg("error getting current working directory")
	}

	// Copy all the files into the ZIP.
	addToZip := func(path string, d fs.DirEntry, err error) error {
		sublog := log.With().Str("path", path).Logger()

		if err != nil {
			sublog.Error().Err(err).Msg("error received from filepath.WalkDir, aborting")
			return err
		}

		// Construct the path inside the ZIP file.
		relpath, err := filepath.Rel(basePath, path)
		if err != nil {
			return fmt.Errorf("making %s relative to %s: %w", path, basePath, err)
		}

		if d.IsDir() {
			switch {
			case filepath.Base(path) == "__pycache__":
				return fs.SkipDir
			case relpath == filepath.Join("manager", "docs"):
				return fs.SkipDir
			case strings.HasPrefix(filepath.Base(path), "."):
				// Skip directories like .mypy_cache, etc.
				return fs.SkipDir
			default:
				// Just recurse into this directory.
				return nil
			}
		}

		sublog.Debug().Str("path", relpath).Msg("adding file to ZIP")

		// Read the file's contents. These are just Python files and maybe a Wheel,
		// nothing huge.
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Write into the ZIP file.
		fileInZip, err := zipWriter.Create(relpath)
		if err != nil {
			return fmt.Errorf("creating %s in ZIP: %w", relpath, err)
		}
		_, err = fileInZip.Write(contents)
		if err != nil {
			return fmt.Errorf("writing to %s in ZIP: %w", relpath, err)
		}

		return nil
	}

	logger.Debug().Str("cwd", basePath).Msg("walking directory")

	if err := filepath.WalkDir(basePath, addToZip); err != nil {
		logger.Fatal().Err(err).Msg("error filling ZIP file")
	}

	comment := fmt.Sprintf("%s add-on for Blender, version %s",
		appinfo.ApplicationName,
		appinfo.ApplicationVersion,
	)
	if err := zipWriter.SetComment(comment); err != nil {
		logger.Fatal().Err(err).Msg("error setting ZIP comment")
	}

	if err := zipWriter.Close(); err != nil {
		logger.Fatal().Err(err).Msg("error closing ZIP file")
	}
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.version, "version", false, "Shows the application version, then exits.")
	flag.BoolVar(&cliArgs.quiet, "quiet", false, "Only log warning-level and worse.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")
	flag.BoolVar(&cliArgs.trace, "trace", false, "Enable trace-level logging.")
	flag.StringVar(&cliArgs.filename, "filename", "web/static/flamenco-addon.zip", "Filename to save the add-on to.")
	flag.Parse()
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
