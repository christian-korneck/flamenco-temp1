package config

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
)

func TestDefaultSettings(t *testing.T) {
	config, err := loadConf("nonexistant.yaml")
	assert.NotNil(t, err) // should indicate an error to open the file.

	// The settings should contain the defaults, though.
	assert.Equal(t, latestConfigVersion, config.Meta.Version)
	assert.Equal(t, "./task-logs", config.TaskLogsPath)
	assert.Equal(t, "64ad4c21-6042-4378-9cdf-478f88b4f990", config.SSDPDeviceUUID)

	assert.Contains(t, config.Variables, "job_storage")
	assert.Contains(t, config.Variables, "render")
	assert.Equal(t, "oneway", config.Variables["ffmpeg"].Direction)
	assert.Equal(t, "/usr/bin/ffmpeg", config.Variables["ffmpeg"].Values[0].Value)
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

	err := c.checkVariables()
	assert.Equal(t, ErrMissingVariablePlatform, err)

	assert.Equal(t, c.Variables["blender"].Values[0].Value, "/path/to/blender")
	assert.Equal(t, c.Variables["blender"].Values[1].Value, "/valid/path/blender")
}
