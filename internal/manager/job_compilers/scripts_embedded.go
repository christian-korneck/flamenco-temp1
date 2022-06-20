package job_compilers

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"embed"
	"fmt"
	"io/fs"
)

// Scripts from the `./scripts` subdirectory are embedded into the executable
// here. Note that accessing these files still requires explicit use of the
// `scripts/` subdirectory, which is abstracted away by `getEmbeddedScriptFS()`.
// It is recommended to use that function to get the embedded scripts
// filesystem.

//go:embed scripts
var _embeddedScriptsFS embed.FS

// getEmbeddedScriptFS returns the `fs.FS` interface that allows access to the
// embedded job compiler scripts.
func getEmbeddedScriptFS() fs.FS {
	scriptsSubFS, err := fs.Sub(_embeddedScriptsFS, "scripts")
	if err != nil {
		panic(fmt.Sprintf("failed to find embedded 'scripts' directory: %v", err))
	}
	return scriptsSubFS
}
