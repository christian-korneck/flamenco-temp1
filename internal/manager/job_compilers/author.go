package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
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
	Dependencies []*AuthoredTask `json:"omitempty" yaml:"omitempty"`
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
		uuid.New(),
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
	// Construct a task without dependencies, and a separate list of dependency
	// UUIDs, just for logging. Simply logging `at` would recursively include all
	// dependent tasks, which gets way too big to log.
	if log.Debug().Enabled() {
		deps := make([]string, len(at.Dependencies))
		for i := range at.Dependencies {
			deps[i] = at.Dependencies[i].UUID
		}
		logTask := *at
		logTask.Dependencies = nil

		log.Debug().
			Str("job", aj.Name).
			Strs("depTaskIDs", deps).
			Interface("task", logTask).Msg("add task")
	}

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
