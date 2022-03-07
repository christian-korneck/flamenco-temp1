package config

// SPDX-License-Identifier: GPL-3.0-or-later

const (
	// The "audience" of task variables.
	VariableAudienceAll     VariableAudience = "all"
	VariableAudienceWorkers VariableAudience = "workers"
	VariableAudienceUsers   VariableAudience = "users"
)

type VariableAudience string
