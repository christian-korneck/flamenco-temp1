package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"os"
	"runtime"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
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

	mustExport := func(name string, value interface{}) {
		err := obj.Set(name, value)
		if err != nil {
			log.Panic().Err(err).Msgf("unable to register '%s' in Goja 'process' module", name)
		}
	}

	mustExport("cwd", p.cwd)

	// To get a list of possible values of runtime.GOOS, run `go tool dist list`.
	// The NodeJS values are documented on https://nodejs.org/api/process.html#processplatform
	// Both lists are equal enough to just use runtime.GOOS here.
	mustExport("platform", runtime.GOOS)
}
