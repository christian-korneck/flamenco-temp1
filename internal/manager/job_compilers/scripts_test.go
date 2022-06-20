package job_compilers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadScriptsFrom_skip_nonjs(t *testing.T) {
	thisDirFS := os.DirFS(".")
	compilers, err := loadScriptsFrom(thisDirFS)
	assert.NoError(t, err, "input without JS files should not cause errors")
	assert.Empty(t, compilers)
}

func TestLoadScriptsFrom_on_disk_js(t *testing.T) {
	scriptsFS := os.DirFS("scripts-for-unittest")
	compilers, err := loadScriptsFrom(scriptsFS)

	assert.NoError(t, err)
	expectKeys := map[string]bool{
		"echo-and-sleep":        true,
		"simple-blender-render": true,
		// Should NOT contain an entry for 'empty.js'.
	}
	assert.Equal(t, expectKeys, keys(compilers))
}

// keys returns the set of keys of the mapping.
func keys[K comparable, V any](mapping map[K]V) map[K]bool {
	keys := map[K]bool{}
	for k := range mapping {
		keys[k] = true
	}
	return keys
}
