package job_compilers

import (
	"embed"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

//go:embed scripts/*.js
var scriptsFS embed.FS

func (c *GojaJobCompiler) loadScripts() error {
	scripts, err := scriptsFS.ReadDir("scripts")
	if err != nil {
		return fmt.Errorf("failed to find scripts: %w", err)
	}

	for _, script := range scripts {
		filename := path.Join("scripts", script.Name())
		file, err := scriptsFS.Open(filename)
		if err != nil {
			return fmt.Errorf("failed to open embedded script: %w", err)
		}

		script_bytes, err := io.ReadAll(file)
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("failed to read script")
			continue
		}

		program, err := goja.Compile(filename, string(script_bytes), true)
		if err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("failed to compile script")
			continue
		}

		jobType := filenameToJobType(script.Name())
		c.jobtypes[jobType] = program

		log.Debug().Str("script", script.Name()).Str("jobType", jobType).Msg("loaded script")
	}

	return nil
}

func filenameToJobType(filename string) string {
	extension := path.Ext(filename)
	stem := filename[:len(filename)-len(extension)]
	return strings.ReplaceAll(stem, "_", "-")
}
