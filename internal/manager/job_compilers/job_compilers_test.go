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

// The example job is expected to result in these arguments for FFmpeg.
var expectedFramesToVideoArgs = []interface{}{
	"-c:v", "h264", "-crf", "20", "-g", "18", "-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2", "-pix_fmt", "yuv420p", "-r", int64(24), "-y",
}

func exampleSubmittedJob() api.SubmittedJob {
	settings := api.JobSettings{
		AdditionalProperties: map[string]interface{}{
			"blender_cmd":            "{blender}",
			"blendfile":              "/render/sf/jobs/scene123.blend",
			"chunk_size":             3,
			"extract_audio":          true,
			"format":                 "PNG",
			"fps":                    24.0,
			"frames":                 "1-10",
			"images_or_video":        "images",
			"image_file_extension":   ".png",
			"video_container_format": "",
			"render_output_root":     "/render/sprites/farm_output/promo/square_ellie",
			"add_path_components":    1,
			"render_output_path":     "/render/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2/######",
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

	// Tasks should have been created to render the frames: 1-3, 4-6, 7-9, 10, video-encoding, and cleanup
	assert.Len(t, aj.Tasks, 6)
	t0 := aj.Tasks[0]
	expectCliArgs := []interface{}{ // They are strings, but Goja doesn't know that and will produce an []interface{}.
		"--render-output", "/render/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/######",
		"--render-format", settings["format"].(string),
		"--render-frame", "1..3",
	}
	assert.NotEmpty(t, t0.UUID)
	assert.Equal(t, "render-1-3", t0.Name)
	assert.Equal(t, 1, len(t0.Commands))
	assert.Equal(t, "blender-render", t0.Commands[0].Name)
	assert.EqualValues(t, AuthoredCommandParameters{
		"exe":        "{blender}",
		"blendfile":  settings["blendfile"].(string),
		"args":       expectCliArgs,
		"argsBefore": make([]interface{}, 0),
	}, t0.Commands[0].Parameters)

	tVideo := aj.Tasks[4] // This should be a video encoding task
	assert.NotEmpty(t, tVideo.UUID)
	assert.Equal(t, "preview-video", tVideo.Name)
	assert.Equal(t, 1, len(tVideo.Commands))
	assert.Equal(t, "frames-to-video", tVideo.Commands[0].Name)
	assert.EqualValues(t, AuthoredCommandParameters{
		"exe":        "ffmpeg",
		"inputGlob":  "/render/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/*.png",
		"outputFile": "/render/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/scene123-1-10.mp4",
		"fps":        int64(24),
		"args":       expectedFramesToVideoArgs,
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
	sj.Settings.AdditionalProperties["blendfile"] = "R:\\sf\\jobs\\scene123.blend"
	sj.Settings.AdditionalProperties["render_output_path"] = "R:\\sprites\\farm_output\\promo\\square_ellie\\square_ellie.lighting_light_breakdown2\\######"

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

	// Tasks should have been created to render the frames: 1-3, 4-6, 7-9, 10, video-encoding, and cleanup
	assert.Len(t, aj.Tasks, 6)
	t0 := aj.Tasks[0]
	expectCliArgs := []interface{}{ // They are strings, but Goja doesn't know that and will produce an []interface{}.
		// The render output is constructed by the job compiler, and thus transforms to forward slashes.
		"--render-output", "R:/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/######",
		"--render-format", settings["format"].(string),
		"--render-frame", "1..3",
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
	assert.Equal(t, "preview-video", tVideo.Name)
	assert.Equal(t, 1, len(tVideo.Commands))
	assert.Equal(t, "frames-to-video", tVideo.Commands[0].Name)
	assert.EqualValues(t, AuthoredCommandParameters{
		"exe":        "ffmpeg",
		"inputGlob":  "R:/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/*.png",
		"outputFile": "R:/sprites/farm_output/promo/square_ellie/square_ellie.lighting_light_breakdown2__intermediate-2006-01-02_090405/scene123-1-10.mp4",
		"fps":        int64(24),
		"args":       expectedFramesToVideoArgs,
	}, tVideo.Commands[0].Parameters)
}

func TestSimpleBlenderRenderOutputPathFieldReplacement(t *testing.T) {
	c := mockedClock(t)

	s, err := Load(c)
	assert.NoError(t, err)

	// Compiling a job should be really fast.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sj := exampleSubmittedJob()
	sj.Settings.AdditionalProperties["render_output_path"] = "/root/{timestamp}/jobname/######"

	aj, err := s.Compile(ctx, sj)
	if err != nil {
		t.Fatalf("job compiler failed: %v", err)
	}
	if aj == nil {
		t.Fatalf("job compiler returned nil but no error")
	}

	// The job compiler should have replaced the {timestamp} and {ext} fields.
	assert.Equal(t, "/root/2006-01-02_090405/jobname/######", aj.Settings["render_output_path"])

	// Tasks should have been created to render the frames: 1-3, 4-6, 7-9, 10, video-encoding, and cleanup
	assert.Len(t, aj.Tasks, 6)
	t0 := aj.Tasks[0]
	expectCliArgs := []interface{}{ // They are strings, but Goja doesn't know that and will produce an []interface{}.
		"--render-output", "/root/2006-01-02_090405/jobname__intermediate-2006-01-02_090405/######",
		"--render-format", sj.Settings.AdditionalProperties["format"].(string),
		"--render-frame", "1..3",
	}
	assert.EqualValues(t, AuthoredCommandParameters{
		"exe":        "{blender}",
		"blendfile":  sj.Settings.AdditionalProperties["blendfile"].(string),
		"args":       expectCliArgs,
		"argsBefore": make([]interface{}, 0),
	}, t0.Commands[0].Parameters)

	tVideo := aj.Tasks[4] // This should be a video encoding task
	assert.EqualValues(t, AuthoredCommandParameters{
		"exe":        "ffmpeg",
		"inputGlob":  "/root/2006-01-02_090405/jobname__intermediate-2006-01-02_090405/*.png",
		"outputFile": "/root/2006-01-02_090405/jobname__intermediate-2006-01-02_090405/scene123-1-10.mp4",
		"fps":        int64(24),
		"args":       expectedFramesToVideoArgs,
	}, tVideo.Commands[0].Parameters)

}

func TestEtag(t *testing.T) {
	c := mockedClock(t)

	s, err := Load(c)
	assert.NoError(t, err)

	// Etags should be computed when the compiler VM is obtained.
	vm, err := s.compilerVMForJobType("echo-sleep-test")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	const expectEtag = "eba586e16d6b55baaa43e32f9e78ae514b457fee"
	assert.Equal(t, expectEtag, vm.jobTypeEtag)

	// A mismatching Etag should prevent job compilation.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sj := api.SubmittedJob{
		Name:              "job name",
		Type:              "echo-sleep-test",
		Priority:          50,
		SubmitterPlatform: "linux",
		Settings: &api.JobSettings{AdditionalProperties: map[string]interface{}{
			"message": "hey",
		}},
	}

	{ // Test without etag.
		aj, err := s.Compile(ctx, sj)
		if assert.NoError(t, err, "job without etag should always be accepted") {
			assert.NotNil(t, aj)
		}
	}

	{ // Test with bad etag.
		sj.TypeEtag = ptr("this is not the right etag")
		_, err := s.Compile(ctx, sj)
		assert.ErrorIs(t, err, ErrJobTypeBadEtag)
	}

	{ // Test with correct etag.
		sj.TypeEtag = ptr(expectEtag)
		aj, err := s.Compile(ctx, sj)
		if assert.NoError(t, err, "job with correct etag should be accepted") {
			assert.NotNil(t, aj)
		}
	}
}

func ptr[T any](value T) *T {
	return &value
}
