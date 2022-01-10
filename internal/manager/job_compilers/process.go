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
	"os"
	"runtime"

	"github.com/dop251/goja"
)

// Process implements a subset of the built-in NodeJS process object.
// It purely exists to make a few NPM-installed packages work.
// See https://nodejs.org/api/process.html
type Process struct {
	runtime *goja.Runtime
}

func (p *Process) cwd() (string, error) {
	return os.Getwd()
}

func ProcessModule(r *goja.Runtime, module *goja.Object) {
	p := &Process{
		runtime: r,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("cwd", p.cwd)

	// To get a list of possible values of runtime.GOOS, run `go tool dist list`.
	// The NodeJS values are documented on https://nodejs.org/api/process.html#processplatform
	// Both lists are equal enough to just use runtime.GOOS here.
	obj.Set("platform", runtime.GOOS)
}
