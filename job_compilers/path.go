package job_compilers

import (
	"path/filepath"

	"github.com/dop251/goja"
)

// PathModule provides file path manipulation functions by wrapping Go's `path`.
func PathModule(r *goja.Runtime, module *goja.Object) {
	obj := module.Get("exports").(*goja.Object)
	obj.Set("basename", filepath.Base)
	obj.Set("dirname", filepath.Dir)
	obj.Set("join", filepath.Join)
	obj.Set("stem", Stem)
}

func Stem(fpath string) string {
	base := filepath.Base(fpath)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
