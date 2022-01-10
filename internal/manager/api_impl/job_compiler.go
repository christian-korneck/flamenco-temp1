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
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/blender/flamenco-goja-test/pkg/api"
)

func (f *Flamenco) GetJobTypes(e echo.Context) error {
	// Some helper functions because Go doesn't allow taking the address of a literal.
	defaultString := func(s string) *interface{} {
		var iValue interface{}
		iValue = s
		return &iValue
	}
	defaultInt32 := func(i int32) *interface{} {
		var iValue interface{}
		iValue = i
		return &iValue
	}
	defaultBool := func(b bool) *interface{} {
		var iValue interface{}
		iValue = b
		return &iValue
	}
	boolPtr := func(b bool) *bool {
		return &b
	}
	choicesStr := func(choices ...string) *[]string {
		return &choices
	}

	// TODO: dynamically build based on the actually registered job types.
	types := api.AvailableJobTypes{
		JobTypes: []api.AvailableJobType{{
			Name: "simple-blender-render",
			Settings: []api.AvailableJobSetting{
				{Key: "blender_cmd", Type: "string", Default: defaultString("{blender}")},
				{Key: "chunk_size", Type: "int32", Default: defaultInt32(1)},
				{Key: "frames", Type: "string", Required: boolPtr(true)},
				{Key: "render_output", Type: "string", Required: boolPtr(true)},
				{Key: "fps", Type: "int32"},
				{Key: "extract_audio", Type: "bool", Default: defaultBool(true)},
				{Key: "images_or_video",
					Type:     "string",
					Required: boolPtr(true),
					Choices:  choicesStr("images", "video"),
					Visible:  boolPtr(false),
				},
				{Key: "format", Type: "string", Required: boolPtr(true)},
				{Key: "output_file_extension", Type: "string", Required: boolPtr(true)},
			},
		}},
	}

	return e.JSON(http.StatusOK, &types)
}
