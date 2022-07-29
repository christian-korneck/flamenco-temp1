package main

import (
	"errors"
	"io/fs"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/find_ffmpeg"
)

// findFFmpeg tries to find FFmpeg, in order to show its version (if found) or a warning (if not).
func findFFmpeg() {
	result, err := find_ffmpeg.Find()
	switch {
	case errors.Is(err, fs.ErrNotExist):
		log.Warn().Msg("FFmpeg could not be found on this system, render jobs may not run correctly")
	case err != nil:
		log.Warn().Err(err).Msg("there was an unexpected error finding FFmepg on this system, render jobs may not run correctly")
	default:
		log.Info().Str("path", result.Path).Str("version", result.Version).Msg("FFmpeg found on this system")
	}
}
