// Package job_compilers contains functionality to convert a Flamenco job
// definition into concrete tasks and commands to execute by Workers.
package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/uuid"
	"git.blender.org/flamenco/pkg/api"
)

var ErrJobTypeUnknown = errors.New("job type unknown")
var ErrScriptIncomplete = errors.New("job compiler script incomplete")

// Service contains job compilers defined in JavaScript.
type Service struct {
	compilers   map[string]Compiler // Mapping from job type name to the job compiler of that type.
	registry    *require.Registry   // Goja module registry.
	timeService TimeService
}

type Compiler struct {
	jobType  string
	program  *goja.Program // Compiled JavaScript file.
	filename string        // The filename of that JS file.
}

type VM struct {
	runtime  *goja.Runtime // Goja VM containing the job compiler script.
	compiler Compiler      // Program loaded into this VM.
}

// jobCompileFunc is a function that fills job.Tasks.
type jobCompileFunc func(job *AuthoredJob) error

// TimeService is a service that can tell the current time.
type TimeService interface {
	Now() time.Time
}

// Load returns a job compiler service with all JS files loaded.
func Load(ts TimeService) (*Service, error) {
	service := Service{
		compilers:   map[string]Compiler{},
		timeService: ts,
	}

	if err := service.loadScripts(); err != nil {
		return nil, err
	}

	staticFileLoader := func(path string) ([]byte, error) {
		// TODO: this should try different filesystems, once we allow loading from
		// disk as well.
		content, err := loadScriptBytes(scriptsFS, path)
		if err != nil {
			// The 'require' module uses this to try different variations of the path
			// in order to find it (without .js, with .js, etc.), so don't log any of
			// such errors.
			return nil, require.ModuleFileDoesNotExistError
		}
		return content, nil
	}

	service.registry = require.NewRegistry(require.WithLoader(staticFileLoader))
	service.registry.RegisterNativeModule("author", AuthorModule)
	service.registry.RegisterNativeModule("path", PathModule)
	service.registry.RegisterNativeModule("process", ProcessModule)

	return &service, nil
}

func (s *Service) Compile(ctx context.Context, sj api.SubmittedJob) (*AuthoredJob, error) {
	vm, err := s.compilerForJobType(sj.Type)
	if err != nil {
		return nil, err
	}

	// Create an AuthoredJob from this SubmittedJob.
	aj := AuthoredJob{
		JobID:    uuid.New(),
		Created:  s.timeService.Now(),
		Name:     sj.Name,
		JobType:  sj.Type,
		Priority: sj.Priority,
		Status:   api.JobStatusUnderConstruction,

		Settings: make(JobSettings),
		Metadata: make(JobMetadata),
	}
	if sj.Settings != nil {
		for key, value := range sj.Settings.AdditionalProperties {
			aj.Settings[key] = value
		}
	}
	if sj.Metadata != nil {
		for key, value := range sj.Metadata.AdditionalProperties {
			aj.Metadata[key] = value
		}
	}

	compiler, err := vm.getCompileJob()
	if err != nil {
		return nil, err
	}
	if err := compiler(&aj); err != nil {
		return nil, err
	}

	log.Info().
		Int("num_tasks", len(aj.Tasks)).
		Str("name", aj.Name).
		Str("jobtype", aj.JobType).
		Msg("job compiled")

	return &aj, nil
}

//  ListJobTypes returns the list of available job types.
func (s *Service) ListJobTypes() api.AvailableJobTypes {
	jobTypes := make([]api.AvailableJobType, 0)
	for typeName := range s.compilers {
		compiler, err := s.compilerForJobType(typeName)
		if err != nil {
			log.Warn().Err(err).Str("jobType", typeName).Msg("unable to determine job type settings")
			continue
		}

		jobType, err := compiler.getJobTypeInfo()
		if err != nil {
			log.Warn().Err(err).Str("jobType", typeName).Msg("unable to determine job type settings")
			continue
		}

		jobTypes = append(jobTypes, jobType)
	}

	sort.Slice(jobTypes, func(i, j int) bool { return jobTypes[i].Name < jobTypes[j].Name })

	return api.AvailableJobTypes{JobTypes: jobTypes}
}

// GetJobType returns information about the named job type.
// Returns ErrJobTypeUnknown when the name doesn't correspond with a known job type.
func (s *Service) GetJobType(typeName string) (api.AvailableJobType, error) {
	compiler, err := s.compilerForJobType(typeName)
	if err != nil {
		return api.AvailableJobType{}, err
	}
	return compiler.getJobTypeInfo()
}

func (vm *VM) getCompileJob() (jobCompileFunc, error) {
	compileJob, isCallable := goja.AssertFunction(vm.runtime.Get("compileJob"))
	if !isCallable {
		// TODO: construct a more elaborate Error type that contains this info, instead of logging here.
		log.Error().
			Str("jobType", vm.compiler.jobType).
			Str("script", vm.compiler.filename).
			Msg("script does not define a compileJob(job) function")
		return nil, ErrScriptIncomplete
	}

	// TODO: wrap this in a nicer way.
	return func(job *AuthoredJob) error {
		_, err := compileJob(nil, vm.runtime.ToValue(job))
		return err
	}, nil
}

func (vm *VM) getJobTypeInfo() (api.AvailableJobType, error) {
	jtValue := vm.runtime.Get("JOB_TYPE")

	var ajt api.AvailableJobType
	if err := vm.runtime.ExportTo(jtValue, &ajt); err != nil {
		// TODO: construct a more elaborate Error type that contains this info, instead of logging here.
		log.Error().
			Err(err).
			Str("jobType", vm.compiler.jobType).
			Str("script", vm.compiler.filename).
			Msg("script does not define a proper JOB_TYPE object")
		return api.AvailableJobType{}, ErrScriptIncomplete
	}

	ajt.Name = vm.compiler.jobType
	return ajt, nil
}
