// Package job_compilers contains functionality to convert a Flamenco job
// definition into concrete tasks and commands to execute by Workers.
package job_compilers

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
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
			"render_output":         "/render/sf/frames/scene123",
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
		t.Logf("job compiler failed: %v", err)
		t.FailNow()
	}
	assert.NotNil(t, aj)
	if aj == nil {
		// Don't bother with the rest of the test, it'll dereference a nil pointer anyway.
		return
	}

	// Properties should be copied as-is.
	assert.Equal(t, sj.Name, aj.Name)
	assert.Equal(t, sj.Type, aj.JobType)
	assert.Equal(t, sj.Priority, aj.Priority)
	assert.EqualValues(t, sj.Settings.AdditionalProperties, aj.Settings)
	assert.EqualValues(t, sj.Metadata.AdditionalProperties, aj.Metadata)

	settings := sj.Settings.AdditionalProperties

	// Tasks should have been created to render the frames.
	assert.Equal(t, 4, len(aj.Tasks))
	t0 := aj.Tasks[0]
	expectCliArgs := []interface{}{ // They are strings, but Goja doesn't know that and will produce an []interface{}.
		"--render-output", "/render/sf__intermediate-2006-01-02_090405/frames",
		"--render-format", settings["format"].(string),
		"--render-frame", "1-3",
	}
	assert.Equal(t, "render-1-3", t0.Name)
	assert.Equal(t, 1, len(t0.Commands))
	assert.Equal(t, "blender-render", t0.Commands[0].Type)
	assert.EqualValues(t, AuthoredCommandParameters{
		"exe":       "{blender}",
		"blendfile": settings["filepath"].(string),
		"args":      expectCliArgs,
	}, t0.Commands[0].Parameters)
}
