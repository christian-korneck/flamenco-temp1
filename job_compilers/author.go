package job_compilers

import (
	"time"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// Author allows scripts to author tasks and commands.
type Author struct {
	runtime *goja.Runtime
}

type AuthoredJob struct {
	JobID    string
	Name     string
	JobType  string
	Priority int8

	Created time.Time

	Settings JobSettings
	Metadata JobMetadata

	Tasks []AuthoredTask
}

type JobSettings map[string]interface{}
type JobMetadata map[string]string

type AuthoredTask struct {
	Name     string
	Commands []AuthoredCommand

	// Dependencies are tasks that need to be completed before this one can run.
	Dependencies []*AuthoredTask
}

type AuthoredCommand struct {
	Type       string
	Parameters map[string]string
}

func (a *Author) Task(name string) (*AuthoredTask, error) {
	at := AuthoredTask{
		name,
		make([]AuthoredCommand, 0),
		make([]*AuthoredTask, 0),
	}
	return &at, nil
}

func (a *Author) Command(cmdType string, parameters map[string]string) (*AuthoredCommand, error) {
	ac := AuthoredCommand{cmdType, parameters}
	return &ac, nil
}

func AuthorModule(r *goja.Runtime, module *goja.Object) {
	a := &Author{
		runtime: r,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("Task", a.Task)
	obj.Set("Command", a.Command)
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
