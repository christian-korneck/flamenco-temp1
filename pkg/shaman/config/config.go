/* (c) 2019, Blender Foundation - Sybren A. Stüvel
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 *
 * The above copyright notice and this permission notice shall be
 * included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package config

import (
	"path/filepath"
	"time"
)

const (
	// fileStoreSubdir is the sub-directory of the configured storage path, used
	// for the file store (i.e. the place where binary blobs are uploaded to).
	fileStoreSubdir = "file-store"

	// checkoutSubDir is the sub-directory of the configured storage path, used
	// for the checkouts of job files (f.e. the blend files used for render jobs,
	// symlinked from the file store dir).
	checkoutSubDir = "jobs"
)

// Config contains all the Shaman configuration
type Config struct {
	// Used only for unit tests, so that they know where the temporary
	// directory created for this test is located.
	TestTempDir string `yaml:"-"`

	Enabled        bool           `yaml:"enabled"`
	StoragePath    string         `yaml:"-"` // Needs to be set externally, not saved in config.
	GarbageCollect GarbageCollect `yaml:"garbageCollect"`
}

// GarbageCollect contains the config options for the GC.
type GarbageCollect struct {
	// How frequently garbage collection is performed on the file store:
	Period time.Duration `yaml:"period"`
	// How old files must be before they are GC'd:
	MaxAge time.Duration `yaml:"maxAge"`
	// Paths to check for symlinks before GC'ing files.
	ExtraCheckoutDirs []string `yaml:"extraCheckoutPaths"`

	// Used by the -gc CLI arg to silently disable the garbage collector
	// while we're performing a manual sweep.
	SilentlyDisable bool `yaml:"-"`
}

// FileStorePath returns the sub-directory of the configured storage path,
// used for the file store (i.e. the place where binary blobs are uploaded to).
func (c Config) FileStorePath() string {
	return filepath.Join(c.StoragePath, fileStoreSubdir)
}

// CheckoutPath returns the sub-directory of the configured storage path, used
// for the checkouts of job files (f.e. the blend files used for render jobs,
// symlinked from the file store dir).
func (c Config) CheckoutPath() string {
	return filepath.Join(c.StoragePath, checkoutSubDir)
}
