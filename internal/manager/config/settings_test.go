package config

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSettings(t *testing.T) {
	config, err := loadConf("nonexistant.yaml")
	assert.NotNil(t, err) // should indicate an error to open the file.

	// The settings should contain the defaults, though.
	assert.Equal(t, latestConfigVersion, config.Meta.Version)
	assert.Equal(t, defaultConfig.TaskLogsPath, config.TaskLogsPath)
	assert.Equal(t, defaultConfig.DatabaseDSN, config.DatabaseDSN)

	assert.Contains(t, config.Variables, "job_storage")
	assert.Contains(t, config.Variables, "render")
	assert.Equal(t, "oneway", config.Variables["ffmpeg"].Direction)
	assert.Equal(t, "ffmpeg", config.Variables["ffmpeg"].Values[0].Value)
	assert.Equal(t, "linux", config.Variables["ffmpeg"].Values[0].Platform)

	linuxPVars, ok := config.VariablesLookup["workers"]["linux"]
	assert.True(t, ok, "workers/linux should have variables: %v", config.VariablesLookup)
	assert.Equal(t, "/shared/flamenco/jobs", linuxPVars["job_storage"])

	winPVars, ok := config.VariablesLookup["users"]["windows"]
	assert.True(t, ok)
	assert.Equal(t, "S:/flamenco/jobs", winPVars["job_storage"])
}

func TestVariableValidation(t *testing.T) {
	c := DefaultConfig()

	platformless := c.Variables["blender"]
	platformless.Values = VariableValues{
		VariableValue{Value: "/path/to/blender"},
		VariableValue{Platform: "linux", Value: "/valid/path/blender"},
	}
	c.Variables["blender"] = platformless

	c.checkVariables()

	assert.Equal(t, c.Variables["blender"].Values[0].Value, "/path/to/blender")
	assert.Equal(t, c.Variables["blender"].Values[1].Value, "/valid/path/blender")
}
