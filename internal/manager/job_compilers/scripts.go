package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rs/zerolog/log"
)

//go:embed scripts
var scriptsFS embed.FS

// loadScripts iterates over all JavaScript files, compiles them, and stores the
// result into `s.compilers`.
func (s *Service) loadScripts() error {
	scriptsSubFS, err := fs.Sub(scriptsFS, "scripts")
	if err != nil {
		return fmt.Errorf("failed to find embedded 'scripts' directory: %w", err)
	}

	return s.loadScriptsFrom(scriptsSubFS)
}

// loadScriptsFrom iterates over all given directory entries, compiles the
// files, and stores the result into `s.compilers`.
func (s *Service) loadScriptsFrom(filesystem fs.FS) error {
	dirEntries, err := fs.ReadDir(filesystem, ".")
	if err != nil {
		return fmt.Errorf("failed to find scripts in %v: %w", filesystem, err)
	}

	for _, dirEntry := range dirEntries {
		filename := dirEntry.Name()
		if !strings.HasSuffix(filename, ".js") {
			continue
		}

		script_bytes, err := s.loadScriptBytes(filesystem, filename)
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("failed to read script")
			continue
		}

		if len(script_bytes) < 8 {
			log.Debug().
				Str("script", filename).
				Int("fileSizeBytes", len(script_bytes)).
				Msg("ignoring tiny JS file, it is unlikely to be a job compiler script")
			continue
		}

		program, err := goja.Compile(filename, string(script_bytes), true)
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("failed to compile script")
			continue
		}

		jobTypeName := filenameToJobType(filename)
		s.compilers[jobTypeName] = Compiler{
			jobType:  jobTypeName,
			program:  program,
			filename: filename,
		}

		log.Debug().
			Str("script", filename).
			Str("jobType", jobTypeName).
			Msg("loaded script")
	}

	return nil
}

func (s *Service) loadScriptBytes(filesystem fs.FS, path string) ([]byte, error) {
	file, err := filesystem.Open(path)
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

	// Set some global functions.
	mustSet("print", jsPrint)
	mustSet("alert", jsAlert)
	mustSet("frameChunker", jsFrameChunker)
	mustSet("formatTimestampLocal", jsFormatTimestampLocal)

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
