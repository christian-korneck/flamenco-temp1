package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rs/zerolog/log"
)

// loadScripts iterates over all JavaScript files, compiles them, and stores the
// result into `s.compilers`.
func (s *Service) loadScripts() error {
	compilers := map[string]Compiler{}

	// Collect all job compilers.
	for _, fs := range getAvailableFilesystems() {
		compilersfromFS, err := loadScriptsFrom(fs)
		if err != nil {
			log.Error().Err(err).Interface("fs", fs).Msg("job compiler: error loading scripts")
			continue
		}
		if len(compilersfromFS) == 0 {
			continue
		}

		log.Debug().Interface("fs", fs).
			Int("numScripts", len(compilersfromFS)).
			Msg("job compiler: found job compiler scripts")

		// Merge the returned compilers into the big map, skipping ones that were
		// already there.
		for name := range compilersfromFS {
			_, found := compilers[name]
			if found {
				continue
			}

			compilers[name] = compilersfromFS[name]
		}
	}

	// Assign the new set of compilers in a thread-safe way.
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.compilers = compilers

	return nil
}

// loadScriptsFrom iterates over files in the root of the given filesystem,
// compiles the files, and returns the "name -> compiler" mapping.
func loadScriptsFrom(filesystem fs.FS) (map[string]Compiler, error) {
	dirEntries, err := fs.ReadDir(filesystem, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to find scripts in %v: %w", filesystem, err)
	}

	compilers := map[string]Compiler{}

	for _, dirEntry := range dirEntries {
		if !dirEntry.Type().IsRegular() {
			continue
		}

		filename := dirEntry.Name()
		if !strings.HasSuffix(filename, ".js") {
			continue
		}

		script_bytes, err := loadFileFromFS(filesystem, filename)
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
		compilers[jobTypeName] = Compiler{
			jobType:  jobTypeName,
			program:  program,
			filename: filename,
		}

		log.Debug().
			Str("script", filename).
			Str("jobType", jobTypeName).
			Msg("job compiler: loaded script")
	}

	return compilers, nil
}

func loadFileFromFS(filesystem fs.FS, path string) ([]byte, error) {
	file, err := filesystem.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s on filesystem %s: %w", path, filesystem, err)
	}
	return io.ReadAll(file)
}

func filenameToJobType(filename string) string {
	extension := path.Ext(filename)
	stem := filename[:len(filename)-len(extension)]
	return strings.ReplaceAll(stem, "_", "-")
}

func newGojaVM(registry *require.Registry) *goja.Runtime {
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
	registry.Enable(vm)
	mustSet("author", require.Require(vm, "author"))
	mustSet("path", require.Require(vm, "path"))
	mustSet("process", require.Require(vm, "process"))

	return vm
}

// compilerVMForJobType returns a Goja *Runtime that has the job compiler script
// for the given job type loaded up.
func (s *Service) compilerVMForJobType(jobTypeName string) (*VM, error) {
	program, ok := s.compilers[jobTypeName]
	if !ok {
		return nil, ErrJobTypeUnknown
	}

	vm := newGojaVM(s.registry)
	if _, err := vm.RunProgram(program.program); err != nil {
		return nil, err
	}

	return &VM{
		runtime:  vm,
		compiler: program,
	}, nil
}
