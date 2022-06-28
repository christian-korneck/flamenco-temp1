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

package shaman

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.blender.org/flamenco/pkg/shaman/config"
	"git.blender.org/flamenco/pkg/shaman/filestore"
	"git.blender.org/flamenco/pkg/shaman/jwtauth"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func createTestShaman() (*Server, func()) {
	conf, confCleanup := config.CreateTestConfig()
	shaman := NewServer(conf, jwtauth.AlwaysDeny{})
	return shaman, confCleanup
}

func makeOld(shaman *Server, expectOld mtimeMap, relPath string) {
	oldTime := time.Now().Add(-2 * shaman.config.GarbageCollect.MaxAge)
	absPath := filepath.Join(shaman.config.FileStorePath(), relPath)

	err := os.Chtimes(absPath, oldTime, oldTime)
	if err != nil {
		panic(err)
	}

	// Do a stat on the file to get the actual on-disk mtime (could be rounded/truncated).
	stat, err := os.Stat(absPath)
	if err != nil {
		panic(err)
	}
	expectOld[absPath] = stat.ModTime()
}

func TestGCCanary(t *testing.T) {
	server, cleanup := createTestShaman()
	defer cleanup()

	assert.True(t, server.config.GarbageCollect.MaxAge > 10*time.Minute,
		"config.GarbageCollect.MaxAge must be big enough for this test to be reliable, is %v",
		server.config.GarbageCollect.MaxAge)
}

func TestGCFindOldFiles(t *testing.T) {
	server, cleanup := createTestShaman()
	defer cleanup()

	filestore.LinkTestFileStore(server.config.FileStorePath())

	// Since all the links have just been created, nothing should be considered old.
	ageThreshold := server.gcAgeThreshold()
	old, err := server.gcFindOldFiles(ageThreshold, log.With().Str("test", "test").Logger())
	assert.NoError(t, err)
	assert.EqualValues(t, mtimeMap{}, old)

	// Make some files old, they should show up in a scan.
	expectOld := mtimeMap{}
	makeOld(server, expectOld, "stored/59/0c148428d5c35fab3ebad2f3365bb469ab9c531b60831f3e826c472027a0b9/3367.blob")
	makeOld(server, expectOld, "stored/80/b749c27b2fef7255e7e7b3c2029b03b31299c75ff1f1c72732081c70a713a3/7488.blob")
	makeOld(server, expectOld, "stored/dc/89f15de821ad1df3e78f8ef455e653a2d1862f2eb3f5ee78aa4ca68eb6fb35/781.blob")

	old, err = server.gcFindOldFiles(ageThreshold, log.With().Str("package", "shaman/test").Logger())
	assert.NoError(t, err)
	assert.EqualValues(t, expectOld, old)
}

// Test of the lower-level functions of the garbage collector.
func TestGCComponents(t *testing.T) {
	server, cleanup := createTestShaman()
	defer cleanup()

	extraCheckoutDir := filepath.Join(server.config.TestTempDir, "extra-checkout")
	server.config.GarbageCollect.ExtraCheckoutDirs = []string{extraCheckoutDir}

	filestore.LinkTestFileStore(server.config.FileStorePath())

	copymap := func(somemap mtimeMap) mtimeMap {
		theCopy := mtimeMap{}
		for key, value := range somemap {
			theCopy[key] = value
		}
		return theCopy
	}

	// Make some files old.
	expectOld := mtimeMap{}
	makeOld(server, expectOld, "stored/30/928ffced04c7008f3324fded86d133effea50828f5ad896196f2a2e190ac7e/6001.blob")
	makeOld(server, expectOld, "stored/59/0c148428d5c35fab3ebad2f3365bb469ab9c531b60831f3e826c472027a0b9/3367.blob")
	makeOld(server, expectOld, "stored/80/b749c27b2fef7255e7e7b3c2029b03b31299c75ff1f1c72732081c70a713a3/7488.blob")
	makeOld(server, expectOld, "stored/dc/89f15de821ad1df3e78f8ef455e653a2d1862f2eb3f5ee78aa4ca68eb6fb35/781.blob")

	// utility mapping to be able to find absolute paths more easily
	absPaths := map[string]string{}
	for absPath := range expectOld {
		absPaths[filepath.Base(absPath)] = absPath
	}

	// No symlinks created yet, so this should report all the files in oldFiles.
	oldFiles := copymap(expectOld)
	err := server.gcFilterLinkedFiles(server.config.CheckoutPath(), oldFiles, log.With().Str("package", "shaman/test").Logger(), nil)
	assert.NoError(t, err)
	assert.EqualValues(t, expectOld, oldFiles)

	// Create some symlinks
	checkoutInfo, err := server.checkoutMan.PrepareCheckout("checkoutID")
	assert.NoError(t, err)
	err = server.checkoutMan.SymlinkToCheckout(absPaths["3367.blob"], server.config.CheckoutPath(),
		filepath.Join(checkoutInfo.RelativePath, "use-of-3367.blob"))
	assert.NoError(t, err)
	err = server.checkoutMan.SymlinkToCheckout(absPaths["781.blob"], extraCheckoutDir,
		filepath.Join(checkoutInfo.RelativePath, "use-of-781.blob"))
	assert.NoError(t, err)

	// There should only be two old file reported now.
	expectRemovable := mtimeMap{
		absPaths["6001.blob"]: expectOld[absPaths["6001.blob"]],
		absPaths["7488.blob"]: expectOld[absPaths["7488.blob"]],
	}
	oldFiles = copymap(expectOld)
	stats := GCStats{}
	err = server.gcFilterLinkedFiles(server.config.CheckoutPath(), oldFiles, log.With().Str("package", "shaman/test").Logger(), &stats)
	assert.Equal(t, 1, stats.numSymlinksChecked) // 1 is in checkoutPath, the other in extraCheckoutDir
	assert.NoError(t, err)
	assert.Equal(t, len(expectRemovable)+1, len(oldFiles)) // one file is linked from the extra checkout dir
	err = server.gcFilterLinkedFiles(extraCheckoutDir, oldFiles, log.With().Str("package", "shaman/test").Logger(), &stats)
	assert.Equal(t, 2, stats.numSymlinksChecked) // 1 is in checkoutPath, the other in extraCheckoutDir
	assert.NoError(t, err)
	assert.EqualValues(t, expectRemovable, oldFiles)

	// Touching a file before requesting deletion should not delete it.
	now := time.Now()
	err = os.Chtimes(absPaths["6001.blob"], now, now)
	assert.NoError(t, err)

	// Running the garbage collector should only remove that one unused and untouched file.
	assert.FileExists(t, absPaths["6001.blob"], "file should exist before GC")
	assert.FileExists(t, absPaths["7488.blob"], "file should exist before GC")
	server.gcDeleteOldFiles(true, oldFiles, log.With().Str("package", "shaman/test").Logger())
	assert.FileExists(t, absPaths["6001.blob"], "file should exist after dry-run GC")
	assert.FileExists(t, absPaths["7488.blob"], "file should exist after dry-run GC")

	server.gcDeleteOldFiles(false, oldFiles, log.With().Str("package", "shaman/test").Logger())

	assert.FileExists(t, absPaths["3367.blob"], "file should exist after GC")
	assert.FileExists(t, absPaths["6001.blob"], "file should exist after GC")
	assert.FileExists(t, absPaths["781.blob"], "file should exist after GC")
	_, err = os.Stat(absPaths["7488.blob"])
	assert.True(t, errors.Is(err, fs.ErrNotExist), "file %s should NOT exist after GC", absPaths["7488.blob"])
}

// Test of the high-level GCStorage() function.
func TestGarbageCollect(t *testing.T) {
	server, cleanup := createTestShaman()
	defer cleanup()

	extraCheckoutDir := filepath.Join(server.config.TestTempDir, "extra-checkout")
	server.config.GarbageCollect.ExtraCheckoutDirs = []string{extraCheckoutDir}

	filestore.LinkTestFileStore(server.config.FileStorePath())

	// Make some files old.
	expectOld := mtimeMap{}
	makeOld(server, expectOld, "stored/30/928ffced04c7008f3324fded86d133effea50828f5ad896196f2a2e190ac7e/6001.blob")
	makeOld(server, expectOld, "stored/59/0c148428d5c35fab3ebad2f3365bb469ab9c531b60831f3e826c472027a0b9/3367.blob")
	makeOld(server, expectOld, "stored/80/b749c27b2fef7255e7e7b3c2029b03b31299c75ff1f1c72732081c70a713a3/7488.blob")
	makeOld(server, expectOld, "stored/dc/89f15de821ad1df3e78f8ef455e653a2d1862f2eb3f5ee78aa4ca68eb6fb35/781.blob")

	// utility mapping to be able to find absolute paths more easily
	absPaths := map[string]string{}
	for absPath := range expectOld {
		absPaths[filepath.Base(absPath)] = absPath
	}

	// Create some symlinks
	checkoutInfo, err := server.checkoutMan.PrepareCheckout("checkoutID")
	assert.NoError(t, err)
	err = server.checkoutMan.SymlinkToCheckout(absPaths["3367.blob"], server.config.CheckoutPath(),
		filepath.Join(checkoutInfo.RelativePath, "use-of-3367.blob"))
	assert.NoError(t, err)
	err = server.checkoutMan.SymlinkToCheckout(absPaths["781.blob"], extraCheckoutDir,
		filepath.Join(checkoutInfo.RelativePath, "use-of-781.blob"))
	assert.NoError(t, err)

	// Running the garbage collector should only remove those two unused files.
	assert.FileExists(t, absPaths["6001.blob"], "file should exist before GC")
	assert.FileExists(t, absPaths["7488.blob"], "file should exist before GC")
	server.GCStorage(true)
	assert.FileExists(t, absPaths["6001.blob"], "file should exist after dry-run GC")
	assert.FileExists(t, absPaths["7488.blob"], "file should exist after dry-run GC")
	server.GCStorage(false)
	_, err = os.Stat(absPaths["6001.blob"])
	assert.True(t, errors.Is(err, fs.ErrNotExist), "file %s should NOT exist after GC", absPaths["6001.blob"])
	_, err = os.Stat(absPaths["7488.blob"])
	assert.True(t, errors.Is(err, fs.ErrNotExist), "file %s should NOT exist after GC", absPaths["7488.blob"])

	// Used files should still exist.
	assert.FileExists(t, absPaths["781.blob"])
	assert.FileExists(t, absPaths["3367.blob"])
}
