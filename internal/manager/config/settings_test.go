package config

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"git.blender.org/flamenco/pkg/crosspath"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSettings(t *testing.T) {
	config, err := loadConf("nonexistant.yaml")
	assert.NotNil(t, err) // should indicate an error to open the file.

	// The settings should contain the defaults, though.
	assert.Equal(t, latestConfigVersion, config.Meta.Version)
	assert.Equal(t, defaultConfig.LocalManagerStoragePath, config.LocalManagerStoragePath)
	assert.Equal(t, defaultConfig.DatabaseDSN, config.DatabaseDSN)

	assert.Equal(t, false, config.Variables["ffmpeg"].IsTwoWay)
	assert.Equal(t, "ffmpeg", config.Variables["ffmpeg"].Values[0].Value)
	assert.Equal(t, VariablePlatformLinux, config.Variables["ffmpeg"].Values[0].Platform)

	assert.Greater(t, config.BlocklistThreshold, 0)
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

func TestStorageImplicitVariablesWithShaman(t *testing.T) {
	c := DefaultConfig(func(c *Conf) {
		// Having the Shaman enabled should create an implicit variable "{jobs}" at the Shaman checkout path.
		c.SharedStoragePath = "/path/to/shaman/storage"
		c.Shaman.Enabled = true

		c.Variables["jobs"] = Variable{
			IsTwoWay: true,
			Values: []VariableValue{
				{
					Audience: VariableAudienceAll,
					Platform: VariablePlatformAll,
					Value:    "this value should not be seen",
				},
			},
		}
	})

	assert.NotContains(t, c.Variables, "jobs", "implicit variables should erase existing variables with the same name")
	if !assert.Contains(t, c.implicitVariables, "jobs") {
		t.FailNow()
	}
	assert.False(t, c.implicitVariables["jobs"].IsTwoWay)
	assert.Equal(t,
		crosspath.ToSlash(c.Shaman.CheckoutPath()),
		c.implicitVariables["jobs"].Values[0].Value)
}

func TestStorageImplicitVariablesWithoutShaman(t *testing.T) {
	c := DefaultConfig(func(c *Conf) {
		// Having the Shaman disabled should create an implicit variable "{jobs}" at the storage path.
		c.SharedStoragePath = "/path/to/shaman/storage"
		c.Shaman.Enabled = false

		c.Variables["jobs"] = Variable{
			IsTwoWay: true,
			Values: []VariableValue{
				{
					Audience: VariableAudienceAll,
					Platform: VariablePlatformAll,
					Value:    "this value should not be seen",
				},
			},
		}
	})

	assert.NotContains(t, c.Variables, "jobs", "implicit variables should erase existing variables with the same name")
	if !assert.Contains(t, c.implicitVariables, "jobs") {
		t.FailNow()
	}
	assert.False(t, c.implicitVariables["jobs"].IsTwoWay)
	assert.Equal(t,
		crosspath.ToSlash(c.SharedStoragePath),
		c.implicitVariables["jobs"].Values[0].Value)
}
