package last_rendered

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// MaxImageSizeBytes is the maximum size in bytes allowed for to-be-processed images.
	MaxImageSizeBytes int64 = 10 * 1024 * 1024

	// queueSize determines how many images can be queued in memory before rejecting
	// new requests to process.
	queueSize = 3

	thumbnailJPEGQuality = 80
)

var (
	ErrQueueFull = errors.New("queue full")

	// thumbnails specifies the thumbnail sizes. For efficiency, they should be
	// listed from large to small, as each thumbnail is the input for the next
	// one.
	thumbnails = []Thumbspec{
		{"last-rendered.jpg", 1920, 1080},
		{"last-rendered-small.jpg", 600, 338},
		{"last-rendered-tiny.jpg", 200, 112},
	}
)

type Storage interface {
	// ForJob returns the directory path for storing job-related files.
	ForJob(jobUUID string) string
}

// LastRenderedProcessor processes "last-rendered" images and stores them with
// the job.
type LastRenderedProcessor struct {
	storage Storage

	// TODO: expand this queue to be per job, so that one spammy job doesn't block
	// the queue for other jobs.
	queue chan Payload

	processingDonecallback func(jobUUID string)
}

// Payload contains the actual image to process.
type Payload struct {
	JobUUID    string // Used to determine the directory to store the image.
	WorkerUUID string // Just for logging.
	MimeType   string
	Image      []byte
}

// Thumbspec specifies a thumbnail size & filename.
type Thumbspec struct {
	Filename  string
	MaxWidth  int
	MaxHeight int
}

func New(storage Storage) *LastRenderedProcessor {
	return &LastRenderedProcessor{
		storage: storage,
		queue:   make(chan Payload, queueSize),
	}
}

// SetCallback registers a 'done' callback, which will be called after the job
// received a new last-rendered image.
// There is only one such callback, so calling this will overwrite the
// previously-set callback function. Pass `nil` to un-set.
func (lrp *LastRenderedProcessor) SetCallback(processingDonecallback func(jobUUID string)) {
	lrp.processingDonecallback = processingDonecallback
}

// Run is the main loop for the processing of images. It will keep running until
// the context is closed.
func (lrp *LastRenderedProcessor) Run(ctx context.Context) {
	log.Debug().Msg("last-rendered: queue runner running")
	defer log.Debug().Msg("last-rendered: queue runner shutting down")

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-lrp.queue:
			lrp.processImage(payload)
		}
	}
}

// QueueImage queues an image for processing.
// Returns `ErrQueueFull` if there is no more space in the queue for new images.
func (lrp *LastRenderedProcessor) QueueImage(payload Payload) error {
	logger := payload.sublogger(log.Logger)
	select {
	case lrp.queue <- payload:
		logger.Debug().Msg("last-rendered: queued image for processing")
		return nil
	default:
		logger.Debug().Msg("last-rendered: unable to queue image for processing")
		return ErrQueueFull
	}
}

// PathForJob returns the base path for this job's last-rendered images.
func (lrp *LastRenderedProcessor) PathForJob(jobUUID string) string {
	return lrp.storage.ForJob(jobUUID)
}

// ThumbSpecs returns the thumbnail specifications.
func (lrp *LastRenderedProcessor) ThumbSpecs() []Thumbspec {
	// Return a copy so modification of the returned slice won't affect the global
	// `thumbnails` variable.
	copied := make([]Thumbspec, len(thumbnails))
	copy(copied, thumbnails)
	return copied
}

// processImage down-scales the image to a few thumbnails for presentation in
// the web interface, and stores those in a job-specific directory.
//
// Because this is intended as internal queue-processing function, errors are
// logged but not returned.
func (lrp *LastRenderedProcessor) processImage(payload Payload) {
	jobDir := lrp.PathForJob(payload.JobUUID)

	logger := log.With().Str("jobDir", jobDir).Logger()
	logger = payload.sublogger(logger)

	// Decode the image.
	image, err := decodeImage(payload)
	if err != nil {
		logger.Error().Err(err).Msg("last-rendered: unable to decode image")
		return
	}

	// Generate the thumbnails.
	for _, spec := range thumbnails {
		thumbLogger := spec.sublogger(logger)
		thumbLogger.Trace().Msg("last-rendered: creating thumbnail")

		image = downscaleImage(spec, image)

		imgpath := filepath.Join(jobDir, spec.Filename)
		if err := saveJPEG(imgpath, image); err != nil {
			thumbLogger.Error().Err(err).Msg("last-rendered: error saving thumbnail")
			break
		}
	}

	// Call the callback, if provided.
	if lrp.processingDonecallback != nil {
		lrp.processingDonecallback(payload.JobUUID)
	}
}

func (p Payload) sublogger(logger zerolog.Logger) zerolog.Logger {
	return logger.With().
		Str("job", p.JobUUID).
		Str("producedByWorker", p.WorkerUUID).
		Str("mime", p.MimeType).
		Logger()
}

func (spec Thumbspec) sublogger(logger zerolog.Logger) zerolog.Logger {
	return logger.With().
		Int("width", spec.MaxWidth).
		Int("height", spec.MaxHeight).
		Str("filename", spec.Filename).
		Logger()
}
