package last_rendered

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"git.blender.org/flamenco/internal/manager/local_storage"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	storage := local_storage.NewNextToExe("lrp")
	defer storage.MustErase()

	lrp := New(storage)
	assert.Equal(t, lrp.storage, storage)
	assert.NotNil(t, lrp.queue)
}

func TestQueueImage(t *testing.T) {
	// Canary test.
	if !assert.Equal(t, queueSize, 3) {
		t.Fatalf("queueSize must be 3 for this test, but is %v", queueSize)
	}

	payload := Payload{
		JobUUID:  "2205227c-593c-46ac-a0d7-e115d4e80dd4",
		MimeType: "image/png",
		Image:    []byte("PNG file contents"),
	}

	storage := local_storage.NewNextToExe("lrp")
	defer storage.MustErase()
	lrp := New(storage)

	assert.NoError(t, lrp.QueueImage(payload))
	assert.NoError(t, lrp.QueueImage(payload))
	assert.NoError(t, lrp.QueueImage(payload))
	assert.ErrorIs(t, lrp.QueueImage(payload), ErrQueueFull)
}

func TestProcessImage(t *testing.T) {
	// Load the test image. Note that this intentionally has an approximate 21:9
	// ratio, whereas the thumbnail specs define a 16:9 ratio.
	imgBytes, err := os.ReadFile("last_rendered_test.jpg")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	jobID := "e078438b-c9f5-43e6-9e86-52f8be91dd12"
	payload := Payload{
		JobUUID:  jobID,
		MimeType: "image/jpeg",
		Image:    imgBytes,
	}

	storage := local_storage.NewNextToExe("lrp")
	defer storage.MustErase()
	lrp := New(storage)

	callbackCount := 0
	payload.Callback = func() {
		callbackCount++
	}

	// Sanity check: the thumbnails shouldn't exist yet.
	jobdir := storage.ForJob(jobID)
	assert.NoFileExists(t, filepath.Join(jobdir, "last-rendered.jpg"))
	assert.NoFileExists(t, filepath.Join(jobdir, "last-rendered-small.jpg"))
	assert.NoFileExists(t, filepath.Join(jobdir, "last-rendered-tiny.jpg"))

	lrp.processImage(payload)

	// The files should exist now.
	assert.FileExists(t, filepath.Join(jobdir, "last-rendered.jpg"))
	assert.FileExists(t, filepath.Join(jobdir, "last-rendered-small.jpg"))
	assert.FileExists(t, filepath.Join(jobdir, "last-rendered-tiny.jpg"))

	assert.Equal(t, callbackCount, 1, "the 'done' callback should be called exactly once")

	// Check the sizes, they should match the thumbspec.
	assertImageSize := func(spec Thumbspec) {
		path := filepath.Join(jobdir, spec.Filename)
		file, err := os.Open(path)
		if !assert.NoError(t, err, "thumbnail %s should be openable", spec.Filename) {
			return
		}
		defer file.Close()

		img, format, err := image.Decode(file)
		if !assert.NoErrorf(t, err, "thumbnail %s should be decodable", spec.Filename) {
			return
		}

		assert.Equalf(t, "jpeg", format, "thumbnail %s not written in the expected format", spec.Filename)
		assert.LessOrEqualf(t, img.Bounds().Dx(), spec.MaxWidth, "thumbnail %s has wrong width", spec.Filename)
		assert.LessOrEqualf(t, img.Bounds().Dy(), spec.MaxHeight, "thumbnail %s has wrong height", spec.Filename)
	}

	for _, spec := range thumbnails {
		assertImageSize(spec)
	}
}
