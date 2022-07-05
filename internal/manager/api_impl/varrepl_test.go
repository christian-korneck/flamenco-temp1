package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

func varreplTestTask() api.AssignedTask {
	return api.AssignedTask{
		Commands: []api.Command{
			{Name: "echo", Parameters: map[string]interface{}{
				"message": "Running Blender from {blender} {blender}"}},
			{Name: "sleep", Parameters: map[string]interface{}{
				"{blender}": 3}},
			{
				Name: "blender_render",
				Parameters: map[string]interface{}{
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

func TestReplaceJobsVariable(t *testing.T) {
	worker := persistence.Worker{Platform: "linux"}

	task := varreplTestTask()
	task.Commands[2].Parameters["filepath"] = "{jobs}/path/in/storage.blend"

	// An implicit variable "{jobs}" should be created, regardless of whether
	// Shaman is enabled or not.

	{ // Test with Shaman enabled.
		conf := config.GetTestConfig(func(c *config.Conf) {
			c.SharedStoragePath = "/path/to/flamenco-storage"
			c.Shaman.Enabled = true
		})

		replacedTask := replaceTaskVariables(&conf, task, worker)
		expectPath := path.Join(conf.Shaman.CheckoutPath(), "path/in/storage.blend")
		assert.Equal(t, expectPath, replacedTask.Commands[2].Parameters["filepath"])
	}

	{ // Test without Shaman.
		conf := config.GetTestConfig(func(c *config.Conf) {
			c.SharedStoragePath = "/path/to/flamenco-storage"
			c.Shaman.Enabled = false
		})

		replacedTask := replaceTaskVariables(&conf, task, worker)
		expectPath := "/path/to/flamenco-storage/jobs/path/in/storage.blend"
		assert.Equal(t, expectPath, replacedTask.Commands[2].Parameters["filepath"])
	}
}
