package job_compilers

import (
	"github.com/dop251/goja"
)

// Author allows scripts to author tasks and commands.
type Author struct {
	runtime *goja.Runtime
}

type AuthoredTask struct {
	Name     string
	Commands []AuthoredCommand
}

type AuthoredCommand struct {
	Type       string
	Parameters map[string]string
}

func (a *Author) Task(name string) (*AuthoredTask, error) {
	at := AuthoredTask{
		name,
		make([]AuthoredCommand, 0),
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

func (at *AuthoredTask) AddCommand(ac *AuthoredCommand) {
	at.Commands = append(at.Commands, *ac)
}
