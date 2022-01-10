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
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/google/uuid"
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

	registry.RegisterNativeModule("author", AuthorModule)
	registry.RegisterNativeModule("path", PathModule)
	registry.RegisterNativeModule("process", ProcessModule)
	compiler.vm.Set("author", require.Require(compiler.vm, "author"))
	compiler.vm.Set("path", require.Require(compiler.vm, "path"))
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

	created, err := time.Parse(time.RFC3339, "2022-01-03T18:53:00+01:00")
	if err != nil {
		panic("hard-coded timestamp is wrong")
	}

	job := AuthoredJob{
		JobID:    uuid.New().String(),
		JobType:  "blender-render",
		Priority: 50,
		Name:     "190_0030_A.lighting",
		Created:  created,
		Settings: JobSettings{
			"blender_cmd":           "{blender}",
			"chunk_size":            5,
			"frames":                "101-172",
			"render_output":         "{render}/sprites/farm_output/shots/190_credits/190_0030_A/190_0030_A.lighting/######",
			"fps":                   24.0,
			"extract_audio":         false,
			"images_or_video":       "images",
			"format":                "JPG",
			"output_file_extension": ".jpg",
			"filepath":              "{shaman}/65/61672427b63a96392cd72d65/pro/shots/190_credits/190_0030_A/190_0030_A.lighting.flamenco.blend",
		},
		Metadata: JobMetadata{
			"user":    "Sybren A. Stüvel <sybren@blender.org>",
			"project": "Sprøte Frøte",
		},
	}
	c.vm.Set("job", &job)

	if _, err := c.vm.RunProgram(program); err != nil {
		return err
	}

	log.Info().
		Int("tasks", len(job.Tasks)).
		Str("name", job.Name).
		Str("jobtype", job.JobType).
		Msg("job created")

	return nil
}
