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
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// PathModule provides file path manipulation functions by wrapping Go's `path`.
func PathModule(r *goja.Runtime, module *goja.Object) {
	obj := module.Get("exports").(*goja.Object)

	mustExport := func(name string, value interface{}) {
		err := obj.Set(name, value)
		if err != nil {
			log.Panic().Err(err).Msgf("unable to register '%s' in Goja 'path' module", name)
		}
	}

	mustExport("basename", filepath.Base)
	mustExport("dirname", filepath.Dir)
	mustExport("join", filepath.Join)
	mustExport("stem", Stem)
}

func Stem(fpath string) string {
	base := filepath.Base(fpath)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
