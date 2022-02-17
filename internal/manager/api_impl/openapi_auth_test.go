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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceURLPathVariables(t *testing.T) {
	// Test the regexp first.
	assert.True(t, urlVariablesReplacer.Match([]byte("/:var")))
	assert.True(t, urlVariablesReplacer.Match([]byte("/:var/")))

	assert.Equal(t, "", replaceURLPathVariables(""))
	assert.Equal(t, "/just/some/path", replaceURLPathVariables("/just/some/path"))
	assert.Equal(t, "/variable/at/{end}", replaceURLPathVariables("/variable/at/:end"))
	assert.Equal(t, "/mid/{var}/end", replaceURLPathVariables("/mid/:var/end"))
}
