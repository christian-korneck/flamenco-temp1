package config

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"sync"
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

func TestStorageImplicitVariablesWithShaman(t *testing.T) {
	c := DefaultConfig(func(c *Conf) {
		// Having the Shaman enabled should create an implicit variable "{jobs}" at the Shaman checkout path.
		c.SharedStoragePath = "/path/to/shaman/storage"
		c.Shaman.Enabled = true

		c.Variables["jobs"] = Variable{
			IsTwoWay: false,
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
			IsTwoWay: false,
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

func TestExpandVariables(t *testing.T) {
	c := DefaultConfig(func(c *Conf) {
		c.Variables["demo"] = Variable{
			Values: []VariableValue{
				{Value: "demo-value", Audience: VariableAudienceAll, Platform: VariablePlatformDarwin},
			},
		}
		c.Variables["ffmpeg"] = Variable{
			Values: []VariableValue{
				{Value: "/path/to/ffmpeg", Audience: VariableAudienceUsers, Platform: VariablePlatformLinux},
				{Value: "/path/to/ffmpeg/on/darwin", Audience: VariableAudienceUsers, Platform: VariablePlatformDarwin},
				{Value: "C:/flamenco/ffmpeg", Audience: VariableAudienceUsers, Platform: VariablePlatformWindows},
			},
		}
	})

	feeder := make(chan string, 1)
	receiver := make(chan string, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.ExpandVariables(feeder, receiver, VariableAudienceUsers, VariablePlatformWindows)
	}()

	feeder <- "unchanged value"
	assert.Equal(t, "unchanged value", <-receiver)

	feeder <- "{ffmpeg}"
	assert.Equal(t, "C:/flamenco/ffmpeg", <-receiver)

	feeder <- "{demo}"
	assert.Equal(t, "{demo}", <-receiver, "missing value on the platform should not be replaced")

	close(feeder)
	wg.Wait()
	close(receiver)
}

func TestExpandVariablesWithTwoWay(t *testing.T) {

	c := DefaultConfig(func(c *Conf) {
		// Mock that the Manager is running on Linux right now.
		c.currentGOOS = VariablePlatformLinux

		// Register one variable in the same way that the implicit 'jobs' variable is registered.
		c.Variables["locally-set-path"] = Variable{
			Values: []VariableValue{
				{Value: "/path/on/linux", Platform: VariablePlatformAll, Audience: VariableAudienceAll},
			},
		}
		// This two-way variable should be used to translate the path as well.
		c.Variables["platform-specifics"] = Variable{
			IsTwoWay: true,
			Values: []VariableValue{
				{Value: "/path/on/linux", Platform: VariablePlatformLinux, Audience: VariableAudienceWorkers},
				{Value: "/path/on/darwin", Platform: VariablePlatformDarwin, Audience: VariableAudienceWorkers},
				{Value: "C:/path/on/windows", Platform: VariablePlatformWindows, Audience: VariableAudienceWorkers},
			},
		}
	})

	feeder := make(chan string, 1)
	receiver := make(chan string, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Always target a different-than-current target platform.
		c.ExpandVariables(feeder, receiver, VariableAudienceWorkers, VariablePlatformWindows)
	}()

	// Simple two-way variable replacement.
	feeder <- "/path/on/linux/file.txt"
	assert.Equal(t, "C:/path/on/windows/file.txt", <-receiver)

	// {locally-set-path} expands to a value that's then further replaced by a two-way variable.
	feeder <- "{locally-set-path}/should/be/remapped"
	assert.Equal(t, "C:/path/on/windows/should/be/remapped", <-receiver)

	close(feeder)
	wg.Wait()
	close(receiver)
}
