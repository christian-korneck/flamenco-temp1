package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/crosspath"
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

func TestReplaceVariablesInterfaceArrays(t *testing.T) {
	worker := persistence.Worker{Platform: "linux"}
	conf := config.GetTestConfig()

	task := jsonWash(varreplTestTask())
	replacedTask := replaceTaskVariables(&conf, task, worker)

	// Due to the conversion via JSON, arrays of strings are now arrays of
	// interface{} and still need to be handled properly.
	assert.Equal(t,
		[]interface{}{"--render-out", "/shared/flamenco/render/long/sybren/blender-cloud-addon/flamenco-test__intermediate/render-smpl-0001-0084-frm-######"},
		replacedTask.Commands[2].Parameters["args"],
	)
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

	var storagePath string
	switch runtime.GOOS {
	case "windows":
		storagePath = `C:\path\to\flamenco-storage`
	default:
		storagePath = "/path/to/flamenco-storage"
	}

	{ // Test with Shaman enabled.
		conf := config.GetTestConfig(func(c *config.Conf) {
			c.SharedStoragePath = storagePath
			c.Shaman.Enabled = true
		})

		replacedTask := replaceTaskVariables(&conf, task, worker)
		expectPath := crosspath.Join(crosspath.ToSlash(conf.Shaman.CheckoutPath()), "path/in/storage.blend")
		assert.Equal(t, expectPath, replacedTask.Commands[2].Parameters["filepath"])
	}

	{ // Test without Shaman.
		conf := config.GetTestConfig(func(c *config.Conf) {
			c.SharedStoragePath = storagePath
			c.Shaman.Enabled = false
		})

		replacedTask := replaceTaskVariables(&conf, task, worker)
		expectPath := crosspath.Join(storagePath, "jobs", "path/in/storage.blend")
		assert.Equal(t, expectPath, replacedTask.Commands[2].Parameters["filepath"])
	}
}

func TestReplaceTwoWayVariables(t *testing.T) {
	c := config.DefaultConfig(func(c *config.Conf) {
		// Mock that the Manager is running Linux.
		c.MockCurrentGOOSForTests("linux")

		// Register one variable in the same way that the implicit 'jobs' variable is registered.
		c.Variables["locally-set-path"] = config.Variable{
			Values: []config.VariableValue{
				{Value: "/render/frames", Platform: config.VariablePlatformAll, Audience: config.VariableAudienceAll},
			},
		}
		c.Variables["unused"] = config.Variable{
			Values: []config.VariableValue{
				{Value: "Ignore it, it'll be faaaain!", Platform: config.VariablePlatformAll, Audience: config.VariableAudienceAll},
			},
		}
		// These two-way variables should be used to translate the path as well.
		c.Variables["project"] = config.Variable{
			IsTwoWay: true,
			Values: []config.VariableValue{
				{Value: "/projects/sprite-fright", Platform: config.VariablePlatformAll, Audience: config.VariableAudienceAll},
			},
		}
		c.Variables["render"] = config.Variable{
			IsTwoWay: true,
			Values: []config.VariableValue{
				{Value: "/render", Platform: config.VariablePlatformLinux, Audience: config.VariableAudienceWorkers},
				{Value: "/Volumes/render", Platform: config.VariablePlatformDarwin, Audience: config.VariableAudienceWorkers},
				{Value: "R:", Platform: config.VariablePlatformWindows, Audience: config.VariableAudienceWorkers},
			},
		}
	})

	// Test job without settings or metadata.
	{
		original := varReplSubmittedJob()
		original.Settings = nil
		original.Metadata = nil
		replaced := varReplSubmittedJob()
		replaced.Settings = nil
		replaced.Metadata = nil
		replaceTwoWayVariables(&c, replaced)

		assert.Equal(t, original.Type, replaced.Type, "two-way variable replacement shouldn't happen on the Type property")
		assert.Equal(t, original.Name, replaced.Name, "two-way variable replacement shouldn't happen on the Name property")
		assert.Equal(t, original.Priority, replaced.Priority, "two-way variable replacement shouldn't happen on the Priority property")
		assert.Equal(t, original.SubmitterPlatform, replaced.SubmitterPlatform)
		assert.Nil(t, replaced.Settings)
		assert.Nil(t, replaced.Metadata)
	}

	// Test with settings & metadata.
	{
		original := varReplSubmittedJob()
		replaced := jsonWash(varReplSubmittedJob())
		replaceTwoWayVariables(&c, replaced)

		expectSettings := map[string]interface{}{
			"blender_cmd":           "{blender}",
			"filepath":              "{render}/jobs/sf/scene123.blend",
			"render_output_root":    "{render}/frames/sf/scene123",
			"render_output_path":    "{render}/frames/sf/scene123/Substituição variável bidirecional/######",
			"different_prefix_path": "/backup/render/frames/sf/scene123", // two-way variables should only apply to prefixes.
			"frames":                "1-10",
			"chunk_size":            float64(3),  // Changed type due to the JSON-washing.
			"fps":                   float64(24), // Changed type due to the JSON-washing.
			"extract_audio":         true,
			"images_or_video":       "images",
			"format":                "PNG",
			"output_file_extension": ".png",
		}
		expectMetadata := map[string]string{
			"user.name": "Sybren Stüvel",
			"project":   "Sprite Fright",
			"root":      "{project}",
			"scene":     "{project}/scenes/123",
		}

		assert.Equal(t, original.Type, replaced.Type, "two-way variable replacement shouldn't happen on the Type property")
		assert.Equal(t, original.Name, replaced.Name, "two-way variable replacement shouldn't happen on the Name property")
		assert.Equal(t, original.Priority, replaced.Priority, "two-way variable replacement shouldn't happen on the Priority property")
		assert.Equal(t, original.SubmitterPlatform, replaced.SubmitterPlatform)
		assert.Equal(t, expectSettings, replaced.Settings.AdditionalProperties)
		assert.Equal(t, expectMetadata, replaced.Metadata.AdditionalProperties)
	}
}

func varReplSubmittedJob() api.SubmittedJob {
	return api.SubmittedJob{
		Type:              "simple-blender-render",
		Name:              "Ignore it, it'll be faaaain!",
		Priority:          50,
		SubmitterPlatform: "linux",
		Settings: &api.JobSettings{AdditionalProperties: map[string]interface{}{
			"blender_cmd":           "{blender}",
			"filepath":              "/render/jobs/sf/scene123.blend",
			"render_output_root":    "/render/frames/sf/scene123",
			"render_output_path":    "/render/frames/sf/scene123/Substituição variável bidirecional/######",
			"different_prefix_path": "/backup/render/frames/sf/scene123",
			"frames":                "1-10",
			"chunk_size":            3,
			"fps":                   24,
			"extract_audio":         true,
			"images_or_video":       "images",
			"format":                "PNG",
			"output_file_extension": ".png",
		}},
		Metadata: &api.JobMetadata{AdditionalProperties: map[string]string{
			"user.name": "Sybren Stüvel",
			"project":   "Sprite Fright",
			"root":      "/projects/sprite-fright",
			"scene":     "/projects/sprite-fright/scenes/123",
		}},
	}
}

// jsonWash converts the given value to JSON and back.
// This makes sure the types are as closed to what the API will handle as
// possible, making the difference between "array of strings" and "array of
// interface{}s that happen to be strings".
func jsonWash[T any](value T) T {
	bytes, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	var jsonWashedValue T
	err = json.Unmarshal(bytes, &jsonWashedValue)
	if err != nil {
		panic(err)
	}

	return jsonWashedValue
}
