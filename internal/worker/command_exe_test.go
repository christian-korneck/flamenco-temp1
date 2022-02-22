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
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

func TestCmdSettingAsStrings(t *testing.T) {
	cmd := api.Command{
		Name: "test",
		Parameters: map[string]interface{}{
			"strings": []string{"a", "b"},
			"ints":    []int{3, 4},
			"floats":  []float64{0.47, 0.327},
			"mixed":   []interface{}{"a", 47, 0.327},
		},
	}

	{
		slice, ok := cmdSettingAsStrings(cmd, "strings")
		if ok {
			assert.Equal(t, []string{"a", "b"}, slice)
		} else {
			t.Error("not ok")
		}
	}
	{
		_, ok := cmdSettingAsStrings(cmd, "ints")
		assert.False(t, ok, "only []string or []interface{} are expected to work")
	}
	{
		_, ok := cmdSettingAsStrings(cmd, "floats")
		assert.False(t, ok, "only []string or []interface{} are expected to work")
	}
	{
		slice, ok := cmdSettingAsStrings(cmd, "mixed")
		if ok {
			assert.Equal(t, []string{"a", "47", "0.327"}, slice)
		} else {
			t.Error("not ok")
		}
	}
}