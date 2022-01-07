package job_compilers

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
