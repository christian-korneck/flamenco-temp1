// Package job_compilers contains functionality to convert a Flamenco job
// definition into concrete tasks and commands to execute by Workers.
package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/pkg/api"
)

func exampleSubmittedJob() api.SubmittedJob {
	settings := api.JobSettings{
		AdditionalProperties: map[string]interface{}{
			"blender_cmd":           "{blender}",
			"chunk_size":            3,
			"extract_audio":         true,
			"filepath":              "/render/sf/jobs/scene123.blend",
			"format":                "PNG",
			"fps":                   24,
			"frames":                "1-10",
			"images_or_video":       "images",
			"output_file_extension": ".png",
			"render_output":         "/render/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2/######",
		}}
	metadata := api.JobMetadata{
		AdditionalProperties: map[string]string{
			"project":    "Sprite Fright",
			"user.email": "sybren@blender.org",
			"user.name":  "Sybren Stüvel",
		}}
	sj := api.SubmittedJob{
		Name:     "3Д рендеринг",
		Priority: 50,
		Type:     "simple-blender-render",
		Settings: &settings,
		Metadata: &metadata,
	}
	return sj
}

func mockedClock(t *testing.T) clock.Clock {
	c := clock.NewMock()
	now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	assert.NoError(t, err)
	c.Set(now)
	return c
}

func TestSimpleBlenderRenderHappy(t *testing.T) {
	c := mockedClock(t)

	s, err := Load(c)
	assert.NoError(t, err)

	// Compiling a job should be really fast.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sj := exampleSubmittedJob()
	aj, err := s.Compile(ctx, sj)
	if err != nil {
		t.Fatalf("job compiler failed: %v", err)
	}
	if aj == nil {
		t.Fatalf("job compiler returned nil but no error")
	}

	// Properties should be copied as-is.
	assert.Equal(t, sj.Name, aj.Name)
	assert.Equal(t, sj.Type, aj.JobType)
	assert.Equal(t, sj.Priority, aj.Priority)
	assert.EqualValues(t, sj.Settings.AdditionalProperties, aj.Settings)
	assert.EqualValues(t, sj.Metadata.AdditionalProperties, aj.Metadata)

	settings := sj.Settings.AdditionalProperties

	// Tasks should have been created to render the frames: 1-3, 4-6, 7-9, 10, video-encoding
	assert.Equal(t, 5, len(aj.Tasks))
	t0 := aj.Tasks[0]
	expectCliArgs := []interface{}{ // They are strings, but Goja doesn't know that and will produce an []interface{}.
		"--render-output", "/render/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/######",
		"--render-format", settings["format"].(string),
		"--render-frame", "1-3",
	}
	assert.NotEmpty(t, t0.UUID)
	assert.Equal(t, "render-1-3", t0.Name)
	assert.Equal(t, 1, len(t0.Commands))
	assert.Equal(t, "blender-render", t0.Commands[0].Name)
	assert.EqualValues(t, AuthoredCommandParameters{
		"exe":        "{blender}",
		"blendfile":  settings["filepath"].(string),
		"args":       expectCliArgs,
		"argsBefore": make([]interface{}, 0),
	}, t0.Commands[0].Parameters)

	tVideo := aj.Tasks[4] // This should be a video encoding task
	assert.NotEmpty(t, tVideo.UUID)
	assert.Equal(t, "create-video", tVideo.Name)
	assert.Equal(t, 1, len(tVideo.Commands))
	assert.Equal(t, "create-video", tVideo.Commands[0].Name)
	assert.EqualValues(t, AuthoredCommandParameters{
		"input_files": "/render/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/*.png",
		"output_file": "/render/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/scene123-1-10.mp4",
		"fps":         int64(24),
	}, tVideo.Commands[0].Parameters)

	for index, task := range aj.Tasks {
		if index == 0 {
			continue
		}
		assert.NotEqual(t, t0.UUID, task.UUID, "Task UUIDs should be unique")
	}

	// Check dependencies
	assert.Empty(t, aj.Tasks[0].Dependencies)
	assert.Empty(t, aj.Tasks[1].Dependencies)
	assert.Empty(t, aj.Tasks[2].Dependencies)
	assert.Equal(t, 4, len(tVideo.Dependencies))
	expectDeps := []*AuthoredTask{
		&aj.Tasks[0], &aj.Tasks[1], &aj.Tasks[2], &aj.Tasks[3],
	}
	assert.Equal(t, expectDeps, tVideo.Dependencies)
}

func TestSimpleBlenderRenderWindowsPaths(t *testing.T) {
	c := mockedClock(t)

	s, err := Load(c)
	assert.NoError(t, err)

	// Compiling a job should be really fast.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sj := exampleSubmittedJob()

	// Adjust the job to get paths in Windows notation.
	sj.Settings.AdditionalProperties["filepath"] = "R:\\sf\\jobs\\scene123.blend"
	sj.Settings.AdditionalProperties["render_output"] = "R:\\sprites\\farm_output\\promo\\square_ellie\\square_ellie.lighting_light_breakdown2\\######"

	aj, err := s.Compile(ctx, sj)
	if err != nil {
		t.Fatalf("job compiler failed: %v", err)
	}
	if aj == nil {
		t.Fatalf("job compiler returned nil but no error")
	}

	// Properties should be copied as-is, so also with filesystem paths as-is.
	assert.Equal(t, sj.Name, aj.Name)
	assert.Equal(t, sj.Type, aj.JobType)
	assert.Equal(t, sj.Priority, aj.Priority)
	assert.EqualValues(t, sj.Settings.AdditionalProperties, aj.Settings)
	assert.EqualValues(t, sj.Metadata.AdditionalProperties, aj.Metadata)

	settings := sj.Settings.AdditionalProperties

	// Tasks should have been created to render the frames: 1-3, 4-6, 7-9, 10, video-encoding
	assert.Equal(t, 5, len(aj.Tasks))
	t0 := aj.Tasks[0]
	expectCliArgs := []interface{}{ // They are strings, but Goja doesn't know that and will produce an []interface{}.
		// The render output is constructed by the job compiler, and thus transforms to forward slashes.
		"--render-output", "R:/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/######",
		"--render-format", settings["format"].(string),
		"--render-frame", "1-3",
	}
	assert.NotEmpty(t, t0.UUID)
	assert.Equal(t, "render-1-3", t0.Name)
	assert.Equal(t, 1, len(t0.Commands))
	assert.Equal(t, "blender-render", t0.Commands[0].Name)
	assert.EqualValues(t, AuthoredCommandParameters{
		"exe":        "{blender}",
		"blendfile":  "R:\\sf\\jobs\\scene123.blend", // The blendfile parameter is just copied as-is, so keeps using backslash notation.
		"args":       expectCliArgs,
		"argsBefore": make([]interface{}, 0),
	}, t0.Commands[0].Parameters)

	tVideo := aj.Tasks[4] // This should be a video encoding task
	assert.NotEmpty(t, tVideo.UUID)
	assert.Equal(t, "create-video", tVideo.Name)
	assert.Equal(t, 1, len(tVideo.Commands))
	assert.Equal(t, "create-video", tVideo.Commands[0].Name)
	assert.EqualValues(t, AuthoredCommandParameters{
		"input_files": "R:/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/*.png",
		"output_file": "R:/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/scene123-1-10.mp4",
		"fps":         int64(24),
	}, tVideo.Commands[0].Parameters)
}
