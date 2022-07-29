package main

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/find_blender"
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

// findBlender tries to find Blender, in order to show its version (if found) or a message (if not).
func findBlender() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := find_blender.Find(ctx)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		log.Warn().Msg("Blender could not be found, Flamenco Manager will have to supply a full path")
	case err != nil:
		log.Warn().Err(err).Msg("there was an unexpected error finding Blender on this system, Flamenco Manager will have to supply a full path")
	default:
		log.Info().
			Str("path", result.FoundLocation).
			Str("version", result.BlenderVersion).
			Msg("Blender found on this system, it will be used unless Flamenco Manager specifies a path to a different Blender")
	}
}
