// Package job_compilers contains functionality to convert a Flamenco job
// definition into concrete tasks and commands to execute by Workers.
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
	"context"
	"errors"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

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
	compiler := Service{
		compilers:   map[string]Compiler{},
		timeService: ts,
	}

	if err := compiler.loadScripts(); err != nil {
		return nil, err
	}

	staticFileLoader := func(path string) ([]byte, error) {
		content, err := compiler.loadScriptBytes(path)
		if err != nil {
			// The 'require' module uses this to try different variations of the path
			// in order to find it (without .js, with .js, etc.), so don't log any of
			// such errors.
			return nil, require.ModuleFileDoesNotExistError
		}
		return content, nil
	}

	compiler.registry = require.NewRegistry(require.WithLoader(staticFileLoader))
	compiler.registry.RegisterNativeModule("author", AuthorModule)
	compiler.registry.RegisterNativeModule("path", PathModule)
	compiler.registry.RegisterNativeModule("process", ProcessModule)

	return &compiler, nil
}

func (s *Service) Compile(ctx context.Context, sj api.SubmittedJob) (*AuthoredJob, error) {
	vm, err := s.compilerForJobType(sj.Type)
	if err != nil {
		return nil, err
	}

	// Create an AuthoredJob from this SubmittedJob.
	aj := AuthoredJob{
		JobID:    uuid.New().String(), // Ignore the submitted ID.
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

func (s *Service) Run(jobTypeName string) error {
	vm, err := s.compilerForJobType(jobTypeName)
	if err != nil {
		return err
	}

	compileJob, err := vm.getCompileJob()
	if err != nil {
		return err
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

	if err := compileJob(&job); err != nil {
		return err
	}

	log.Info().
		Int("tasks", len(job.Tasks)).
		Str("name", job.Name).
		Str("jobtype", job.JobType).
		Msg("job created")

	return nil
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
	return api.AvailableJobTypes{JobTypes: jobTypes}
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
