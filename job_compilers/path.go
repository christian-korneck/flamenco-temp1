package job_compilers

import (
	"path"

	"github.com/dop251/goja"
)

// PathModule provides file path manipulation functions by wrapping Go's `path`.
func PathModule(r *goja.Runtime, module *goja.Object) {
	obj := module.Get("exports").(*goja.Object)
	obj.Set("basename", path.Base)
	obj.Set("dirname", path.Dir)
	obj.Set("join", path.Join)
}
