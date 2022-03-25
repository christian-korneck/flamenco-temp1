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

	assert.Equal(t, false, config.Variables["ffmpeg"].IsTwoWay)
	assert.Equal(t, "ffmpeg", config.Variables["ffmpeg"].Values[0].Value)
	assert.Equal(t, VariablePlatformLinux, config.Variables["ffmpeg"].Values[0].Platform)
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

// TODO: Test two-way variables. Even though they're not currently in the
// default configuration, they should work.

func TestShamanImplicitVariables(t *testing.T) {
	c := DefaultConfig(func(c *Conf) {
		// Having the Shaman enabled should create an implicit variable "{jobs}".
		c.Shaman.Enabled = true
		c.Shaman.StoragePath = "/path/to/shaman/storage"
	})

	if !assert.Contains(t, c.Variables, "jobs") {
		t.FailNow()
	}
	assert.False(t, c.Variables["jobs"].IsTwoWay)
	assert.Equal(t, c.Shaman.CheckoutPath(), c.Variables["jobs"].Values[0].Value)
}
