package last_rendered

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	// Import for side-effect of registering decoder.
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/rs/zerolog/log"
)

var (
	supportedMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	ErrMimeTypeUnsupported = errors.New("mime type unsupported")
)

// decodeImage checks the payload mime type, and if okay, decodes the image and returns it.
// Returns `ErrMimeTypeUnsupported` if the mime type is unsupported.
func decodeImage(payload Payload) (image.Image, error) {
	if !supportedMimeTypes[payload.MimeType] {
		return nil, ErrMimeTypeUnsupported
	}

	reader := bytes.NewReader(payload.Image)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}
	return img, nil
}

// saveJPEG writes the given image to a JPEG file.
func saveJPEG(targetpath string, img image.Image) error {
	// Ensure the directory exists.
	targetdir := filepath.Dir(targetpath)
	err := os.MkdirAll(targetdir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating directory %s: %w", targetdir, err)
	}

	file, err := os.Create(targetpath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Warn().Err(err).Str("filename", targetpath).Msg("last-rendered: error closing file")
		}
	}()

	options := jpeg.Options{
		Quality: thumbnailJPEGQuality,
	}
	err = jpeg.Encode(file, img, &options)
	if err != nil {
		return fmt.Errorf("encoding as JPEG: %w", err)
	}

	return nil
}

func downscaleImage(spec Thumbspec, img image.Image) image.Image {
	// Fill out the entire frame, cropping the image if necessary:
	// return imaging.Fill(img, spec.maxWidth, spec.maxHeight, imaging.Center, imaging.Lanczos)

	// Fit the image to the frame, potentially resulting in either a narrower or lower image:
	return imaging.Fit(img, spec.MaxWidth, spec.MaxHeight, imaging.Lanczos)
}
