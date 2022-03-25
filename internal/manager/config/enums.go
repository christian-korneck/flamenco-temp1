package config

// SPDX-License-Identifier: GPL-3.0-or-later

const (
	// The "audience" of task variables.
	VariableAudienceAll     VariableAudience = "all"
	VariableAudienceWorkers VariableAudience = "workers"
	VariableAudienceUsers   VariableAudience = "users"
)

type VariableAudience string

const (
	// the "platform" of task variables. It's a free-form string field, but it has
	// some predefined values here.
	VariablePlatformLinux   VariablePlatform = "linux"
	VariablePlatformWindows VariablePlatform = "windows"
	VariablePlatformDarwin  VariablePlatform = "darwin"
)

type VariablePlatform string
