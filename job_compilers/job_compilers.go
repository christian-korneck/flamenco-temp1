package job_compilers

import (
	"errors"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

var ErrJobTypeUnknown = errors.New("job type unknown")

type GojaJobCompiler struct {
	vm *goja.Runtime

	jobtypes map[string]*goja.Program
}

func Load() (*GojaJobCompiler, error) {
	compiler := GojaJobCompiler{
		vm:       newGojaVM(),
		jobtypes: map[string]*goja.Program{},
	}

	if err := compiler.loadScripts(); err != nil {
		return nil, err
	}

	return &compiler, nil
}

func newGojaVM() *goja.Runtime {
	vm := goja.New()
	vm.Set("print", func(call goja.FunctionCall) goja.Value {
		log.Info().Str("args", call.Argument(0).String()).Msg("print")
		return goja.Undefined()
	})
	vm.Set("create_task", func(call goja.FunctionCall) goja.Value {
		log.Info().Interface("args", call.Arguments).Msg("create_task")
		return goja.Undefined()
	})
	vm.Set("alert", func(call goja.FunctionCall) goja.Value {
		log.Warn().Interface("args", call.Arguments).Msg("alert")
		return goja.Undefined()
	})
	return vm
}

func (c *GojaJobCompiler) Run(jobType string) error {
	program, ok := c.jobtypes[jobType]
	if !ok {
		return ErrJobTypeUnknown
	}

	_, err := c.vm.RunProgram(program)
	return err
}
