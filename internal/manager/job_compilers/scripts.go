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
	"embed"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rs/zerolog/log"
)

//go:embed scripts
var scriptsFS embed.FS

func (s *Service) loadScripts() error {
	scripts, err := scriptsFS.ReadDir("scripts")
	if err != nil {
		return fmt.Errorf("failed to find scripts: %w", err)
	}

	for _, script := range scripts {
		if !strings.HasSuffix(script.Name(), ".js") {
			continue
		}
		filename := path.Join("scripts", script.Name())

		script_bytes, err := s.loadScriptBytes(filename)
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("failed to read script")
			continue
		}

		program, err := goja.Compile(filename, string(script_bytes), true)
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("failed to compile script")
			continue
		}

		jobTypeName := filenameToJobType(script.Name())
		s.compilers[jobTypeName] = Compiler{
			program:  program,
			filename: script.Name(),
		}

		log.Debug().Str("script", script.Name()).Str("jobType", jobTypeName).Msg("loaded script")
	}

	return nil
}

func (s *Service) loadScriptBytes(path string) ([]byte, error) {
	file, err := scriptsFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open embedded script: %w", err)
	}
	return io.ReadAll(file)
}

func filenameToJobType(filename string) string {
	extension := path.Ext(filename)
	stem := filename[:len(filename)-len(extension)]
	return strings.ReplaceAll(stem, "_", "-")
}

func (s *Service) newGojaVM() *goja.Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	mustSet := func(name string, value interface{}) {
		err := vm.Set(name, value)
		if err != nil {
			log.Panic().Err(err).Msgf("unable to register '%s' in Goja VM", name)
		}
	}

	// Set some global functions for script debugging purposes.
	mustSet("print", jsPrint)
	mustSet("alert", jsAlert)

	// Pre-import some useful modules.
	s.registry.Enable(vm)
	mustSet("author", require.Require(vm, "author"))
	mustSet("path", require.Require(vm, "path"))
	mustSet("process", require.Require(vm, "process"))

	return vm
}

// compilerForJobType returns a Goja *Runtime that has the job compiler script for
// the given job type loaded up.
func (s *Service) compilerForJobType(jobTypeName string) (*VM, error) {
	program, ok := s.compilers[jobTypeName]
	if !ok {
		return nil, ErrJobTypeUnknown
	}

	vm := s.newGojaVM()
	if _, err := vm.RunProgram(program.program); err != nil {
		return nil, err
	}

	return &VM{
		runtime:  vm,
		compiler: program,
	}, nil
}
