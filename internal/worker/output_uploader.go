package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"sync"

	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/last_in_one_out_queue"
)

const thumbnailJPEGQuality = 85

// OutputUploader sends (downscaled versions of) rendered images to Flamenco
// Manager. Only one image is sent at a time. A queue of a single image is kept,
// where newly queued images replace older ones.
type OutputUploader struct {
	client FlamencoClient
	queue  *last_in_one_out_queue.LastInOneOutQueue[TaskOutput]
}

type TaskOutput struct {
	TaskID   string
	Filename string
}

func NewOutputUploader(client FlamencoClient) *OutputUploader {
	return &OutputUploader{
		client: client,
		queue:  last_in_one_out_queue.New[TaskOutput](),
	}
}

// OutputProduced enqueues the given filename for processing.
func (ou *OutputUploader) OutputProduced(taskID, filename string) {
	// TODO: Before enqueueing (and thus overwriting any previously queued item),
	// check that this file can actually be handled by the Last Rendered system of
	// Flamenco. It would be a shame if a perfectly-good JPEG file is kicked off
	// the queue by an EXR file we can't handle.
	item := TaskOutput{taskID, filename}
	ou.queue.Enqueue(item)
}

func (ou *OutputUploader) Run(ctx context.Context) {
	log.Info().Msg("output uploader: running")
	defer log.Info().Msg("output uploader: shutting down")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ou.queue.Run(ctx)
	}()

runLoop:
	for {
		select {
		case <-ctx.Done():
			break runLoop
		case item := <-ou.queue.Item():
			ou.process(ctx, item)
		}
	}

	wg.Wait()
}

// process loads the given image, converts it to JPEG, and uploads it to
// Flamenco Manager.
func (ou *OutputUploader) process(ctx context.Context, item TaskOutput) {
	logger := log.With().
		Str("image", item.Filename).
		Str("task", item.TaskID).
		Logger()
	logger.Info().Msg("output uploader: processing file before uploading to Manager")

	jpegBytes := loadAsJPEG(item.Filename)
	if len(jpegBytes) == 0 {
		return // loadAsJPEG() already logged the error.
	}

	// Upload to Manager.
	jpegReader := bytes.NewReader(jpegBytes)
	resp, err := ou.client.TaskOutputProducedWithBodyWithResponse(
		ctx, item.TaskID, "image/jpeg", jpegReader)
	if err != nil {
		logger.Error().Err(err).Msg("output uploader: unable to send image to Manager")
		return
	}

	// Handle the Manager response:
	switch {
	case resp.StatusCode() == http.StatusAccepted:
		logger.Info().Msg("output uploader: Manager accepted our image")
	case resp.JSON411 != nil:
		logger.Error().
			Str("message", resp.JSON411.Message).
			Msg("output uploader: Manager rejected our request, this is a bug in Flamenco Worker")
	case resp.JSON413 != nil:
		logger.Warn().
			Str("message", resp.JSON413.Message).
			Msg("output uploader: Manager rejected our upload, it is too large")
	case resp.JSON415 != nil:
		logger.Error().
			Str("message", resp.JSON415.Message).
			Msg("output uploader: Manager rejected our upload, unsupported file type")
	case resp.JSON429 != nil:
		logger.Warn().
			Str("message", resp.JSON429.Message).
			Msg("output uploader: Manager is too busy to handle this upload")
	case resp.JSONDefault != nil:
		logger.Error().
			Str("message", resp.JSONDefault.Message).
			Msg("output uploader: error from Manager")
	default:
		logger.Error().
			Str("httpStatus", resp.Status()).
			Msg("output uploader: unexpected error from Manager")
	}
}

func loadAsJPEG(imagePath string) []byte {
	logger := log.With().Str("image", imagePath).Logger()

	// Open the output file.
	file, err := os.Open(imagePath)
	if err != nil {
		logger.Error().Err(err).Msg("output uploader: error opening file")
		return nil
	}
	defer file.Close()

	// Try to decode the file as image.
	img, fileType, err := image.Decode(file)
	if err != nil {
		logger.Error().Err(err).Msg("output uploader: cannot decode image file")
		return nil
	}
	logger.Debug().Str("type", fileType).Msg("output uploaded: image decoded")

	// Compress the image as JPEG.
	jpegBuffer := bytes.Buffer{}
	options := jpeg.Options{
		Quality: thumbnailJPEGQuality,
	}
	err = jpeg.Encode(&jpegBuffer, img, &options)
	if err != nil {
		logger.Error().Err(err).Msg("output uploader: cannot encode image as JPEG")
		return nil
	}

	return jpegBuffer.Bytes()
}
