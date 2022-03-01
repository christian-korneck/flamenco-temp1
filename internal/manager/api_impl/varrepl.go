package api_impl

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
	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

type VariableReplacer interface {
	ExpandVariables(valueToExpand string, audience config.VariableAudience, platform string) string
}

// replaceTaskVariables performs variable replacement for worker tasks.
func replaceTaskVariables(replacer VariableReplacer, task api.AssignedTask, worker persistence.Worker) api.AssignedTask {
	repl := func(value string) string {
		return replacer.ExpandVariables(value, "workers", worker.Platform)
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
