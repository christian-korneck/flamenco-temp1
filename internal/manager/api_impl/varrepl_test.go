package api_impl

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
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/config"
	"gitlab.com/blender/flamenco-ng-poc/internal/manager/persistence"
)

func varreplTestTask() persistence.Task {
	return persistence.Task{
		Commands: []persistence.Command{
			{Name: "echo", Parameters: persistence.StringInterfaceMap{
				"message": "Running Blender from {blender} {blender}"}},
			{Name: "sleep", Parameters: persistence.StringInterfaceMap{
				"{blender}": 3}},
			{
				Name: "blender_render",
				Parameters: persistence.StringInterfaceMap{
					"filepath":     "{job_storage}/sybren/2017-06-08-181223.625800-sybren-flamenco-test.flamenco/flamenco-test.flamenco.blend",
					"exe":          "{blender}",
					"otherpath":    "{hey}/haha",
					"frames":       "47",
					"cycles_chunk": 1.0,
					"args":         []string{"--render-out", "{render_long}/sybren/blender-cloud-addon/flamenco-test__intermediate/render-smpl-0001-0084-frm-######"},
				},
			},
		},
	}
}

func TestReplaceVariables(t *testing.T) {
	worker := persistence.Worker{Platform: "linux"}
	task := varreplTestTask()
	conf := config.GetTestConfig()
	replacedTask := replaceTaskVariables(&conf, task, worker)

	// Single string value.
	assert.Equal(t,
		"/opt/myblenderbuild/blender",
		replacedTask.Commands[2].Parameters["exe"],
	)

	// Array value.
	assert.Equal(t,
		[]string{"--render-out", "/shared/flamenco/render/long/sybren/blender-cloud-addon/flamenco-test__intermediate/render-smpl-0001-0084-frm-######"},
		replacedTask.Commands[2].Parameters["args"],
	)

	// Substitution should happen as often as needed.
	assert.Equal(t,
		"Running Blender from /opt/myblenderbuild/blender /opt/myblenderbuild/blender",
		replacedTask.Commands[0].Parameters["message"],
	)

	// No substitution should happen on keys, just on values.
	assert.Equal(t, 3, replacedTask.Commands[1].Parameters["{blender}"])
}

func TestReplacePathsWindows(t *testing.T) {
	worker := persistence.Worker{Platform: "windows"}
	task := varreplTestTask()
	conf := config.GetTestConfig()
	replacedTask := replaceTaskVariables(&conf, task, worker)

	assert.Equal(t,
		"s:/flamenco/jobs/sybren/2017-06-08-181223.625800-sybren-flamenco-test.flamenco/flamenco-test.flamenco.blend",
		replacedTask.Commands[2].Parameters["filepath"],
	)
	assert.Equal(t,
		[]string{"--render-out", "s:/flamenco/render/long/sybren/blender-cloud-addon/flamenco-test__intermediate/render-smpl-0001-0084-frm-######"},
		replacedTask.Commands[2].Parameters["args"],
	)
	assert.Equal(t, "{hey}/haha", replacedTask.Commands[2].Parameters["otherpath"])
}

func TestReplacePathsUnknownOS(t *testing.T) {
	worker := persistence.Worker{Platform: "autumn"}
	task := varreplTestTask()
	conf := config.GetTestConfig()
	replacedTask := replaceTaskVariables(&conf, task, worker)

	assert.Equal(t,
		"hey/sybren/2017-06-08-181223.625800-sybren-flamenco-test.flamenco/flamenco-test.flamenco.blend",
		replacedTask.Commands[2].Parameters["filepath"],
	)
	assert.Equal(t,
		[]string{"--render-out", "{render_long}/sybren/blender-cloud-addon/flamenco-test__intermediate/render-smpl-0001-0084-frm-######"},
		replacedTask.Commands[2].Parameters["args"],
	)
	assert.Equal(t, "{hey}/haha", replacedTask.Commands[2].Parameters["otherpath"])
}
