package job_compilers

import (
	"errors"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
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

	staticFileLoader := func(path string) ([]byte, error) {
		content, err := compiler.loadScript(path)
		if err != nil {
			// The 'require' module uses this to try different variations of the path
			// in order to find it (without .js, with .js, etc.), so don't log any of
			// such errors.
			return nil, require.ModuleFileDoesNotExistError
		}
		return content, nil
	}

	registry := require.NewRegistry(require.WithLoader(staticFileLoader))
	registry.Enable(compiler.vm)

	// NodeJS has this module both importable and globally available.
	registry.RegisterNativeModule("process", ProcessModule)
	compiler.vm.Set("process", require.Require(compiler.vm, "process"))

	return &compiler, nil
}

func newGojaVM() *goja.Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	vm.Set("print", func(call goja.FunctionCall) goja.Value {
		log.Info().Interface("args", call.Arguments).Msg("print")
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

type Job struct {
	ID       int64
	Name     string
	JobType  string
	Priority int8
	Settings JobSettings
	Metadata JobMetadata
}

type JobSettings map[string]interface{}
type JobMetadata map[string]string

func (c *GojaJobCompiler) Run(jobType string) error {
	program, ok := c.jobtypes[jobType]
	if !ok {
		return ErrJobTypeUnknown
	}

	job := Job{
		ID:       327,
		JobType:  "blender-render",
		Priority: 50,
		Name:     "190_0030_A.lighting",
		Settings: JobSettings{
			"blender_cmd":           "{blender}",
			"chunk_size":            5,
			"frames":                "101-172",
			"render_output":         "{render}/sprites/farm_output/shots/190_credits/190_0030_A/190_0030_A.lighting/######",
			"fps":                   24.0,
			"extract_audio":         false,
			"images_or_video":       "images",
			"format":                "OPEN_EXR_MULTILAYER",
			"output_file_extension": ".exr",
			"filepath":              "{shaman}/65/61672427b63a96392cd72d65/pro/shots/190_credits/190_0030_A/190_0030_A.lighting.flamenco.blend",
		},
		Metadata: JobMetadata{
			"user":    "Sybren A. Stüvel <sybren@blender.org>",
			"project": "Sprøte Frøte",
		},
	}
	c.vm.Set("job", &job)

	_, err := c.vm.RunProgram(program)
	return err
}

func (j *Job) NewTask(call goja.ConstructorCall) goja.Value {
	log.Debug().Interface("args", call.Arguments).Msg("job.NewTask")
	return goja.Undefined()
}
