// Package api_impl implements the OpenAPI API from pkg/api/flamenco-manager.yaml.
package api_impl

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

	"github.com/labstack/echo/v4"
	"gitlab.com/blender/flamenco-goja-test/internal/manager/job_compilers"
	"gitlab.com/blender/flamenco-goja-test/pkg/api"
)

type Flamenco struct {
	jobCompiler JobCompiler
	persist     JobPersistenceService
}

type JobPersistenceService interface {
	// StoreJob stores a job in the persistence layer.
	StoreJob(ctx context.Context, authoredJob job_compilers.AuthoredJob) error
}

type JobCompiler interface {
	ListJobTypes() api.AvailableJobTypes
	Compile(ctx context.Context, job api.SubmittedJob) (*job_compilers.AuthoredJob, error)
}

var _ api.ServerInterface = (*Flamenco)(nil)

// NewFlamenco creates a new Flamenco service, using the given JobCompiler.
func NewFlamenco(jc JobCompiler, jps JobPersistenceService) *Flamenco {
	return &Flamenco{
		jobCompiler: jc,
		persist:     jps,
	}
}

// sendPetstoreError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendAPIError(e echo.Context, code int, message string) error {
	petErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	return e.JSON(code, petErr)
}
