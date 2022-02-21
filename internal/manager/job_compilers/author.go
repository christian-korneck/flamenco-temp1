package job_compilers

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
	"errors"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

// Author allows scripts to author tasks and commands.
type Author struct {
	runtime *goja.Runtime
}

type AuthoredJob struct {
	JobID    string
	Name     string
	JobType  string
	Priority int
	Status   api.JobStatus

	Created time.Time

	Settings JobSettings
	Metadata JobMetadata

	Tasks []AuthoredTask
}

type JobSettings map[string]interface{}
type JobMetadata map[string]string

type AuthoredTask struct {
	// Tasks already get their UUID in the authoring stage. This makes it simpler
	// to store the dependencies, as the code doesn't have to worry about value
	// vs. pointer semantics. Tasks can always be unambiguously referenced by
	// their UUID.
	UUID     string
	Name     string
	Type     string
	Priority int
	Commands []AuthoredCommand

	// Dependencies are tasks that need to be completed before this one can run.
	Dependencies []*AuthoredTask
}

type AuthoredCommand struct {
	Name       string
	Parameters AuthoredCommandParameters
}
type AuthoredCommandParameters map[string]interface{}

func (a *Author) Task(name string, taskType string) (*AuthoredTask, error) {
	name = strings.TrimSpace(name)
	taskType = strings.TrimSpace(taskType)
	if name == "" {
		return nil, errors.New("author.Task(name, type): name is required")
	}
	if taskType == "" {
		return nil, errors.New("author.Task(name, type): type is required")
	}

	at := AuthoredTask{
		uuid.New().String(),
		name,
		taskType,
		50, // TODO: handle default priority somehow.
		make([]AuthoredCommand, 0),
		make([]*AuthoredTask, 0),
	}
	return &at, nil
}

func (a *Author) Command(cmdName string, parameters AuthoredCommandParameters) (*AuthoredCommand, error) {
	ac := AuthoredCommand{cmdName, parameters}
	return &ac, nil
}

// AuthorModule exports the Author module members to Goja.
func AuthorModule(r *goja.Runtime, module *goja.Object) {
	a := &Author{
		runtime: r,
	}
	obj := module.Get("exports").(*goja.Object)
	mustExport := func(name string, value interface{}) {
		err := obj.Set(name, value)
		if err != nil {
			log.Panic().Err(err).Msgf("unable to register '%s' in Goja 'author' module", name)
		}
	}

	mustExport("Task", a.Task)
	mustExport("Command", a.Command)
}

func (aj *AuthoredJob) AddTask(at *AuthoredTask) {
	log.Debug().Str("job", at.Name).Interface("task", at).Msg("add task")
	aj.Tasks = append(aj.Tasks, *at)
}

func (at *AuthoredTask) AddCommand(ac *AuthoredCommand) {
	at.Commands = append(at.Commands, *ac)
}

func (at *AuthoredTask) AddDependency(dep *AuthoredTask) error {
	// TODO: check for dependency cycles, return error if there.
	at.Dependencies = append(at.Dependencies, dep)
	return nil
}
