package worker

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
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

// Generate the mock for the client interface.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/client.gen.go -package mocks gitlab.com/blender/flamenco-ng-poc/internal/worker FlamencoClient

// FlamencoClient is a wrapper for api.ClientWithResponsesInterface so that locally mocks can be created.
type FlamencoClient interface {
	api.ClientWithResponsesInterface
}