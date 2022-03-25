package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

type VariableReplacer interface {
	ExpandVariables(valueToExpand string, audience config.VariableAudience, platform config.VariablePlatform) string
}

// replaceTaskVariables performs variable replacement for worker tasks.
func replaceTaskVariables(replacer VariableReplacer, task api.AssignedTask, worker persistence.Worker) api.AssignedTask {
	repl := func(value string) string {
		return replacer.ExpandVariables(value, "workers", config.VariablePlatform(worker.Platform))
	}

	for cmdIndex, cmd := range task.Commands {
		for key, value := range cmd.Parameters {
			switch v := value.(type) {
			case string:
				task.Commands[cmdIndex].Parameters[key] = repl(v)
			case []string:
				replaced := make([]string, len(v))
				for idx := range v {
					replaced[idx] = repl(v[idx])
				}
				task.Commands[cmdIndex].Parameters[key] = replaced
			default:
				continue
			}
		}
	}

	return task
}
