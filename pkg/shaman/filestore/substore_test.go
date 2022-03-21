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

package filestore

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoragePrefix(t *testing.T) {
	bin := storageBin{
		basePath: "/base",
		dirName:  "testunit",
	}
	assert.Equal(t, "/base/testunit", bin.storagePrefix(""))
	assert.Equal(t, "/base/testunit", bin.storagePrefix("/"))
	assert.Equal(t, "/base/testunit/xxx", bin.storagePrefix("xxx"))
	assert.Equal(t, "/base/testunit/xxx", bin.storagePrefix("/xxx"))
}

func TestContains(t *testing.T) {
	bin := storageBin{
		basePath: "/base",
		dirName:  "testunit",
	}
	assert.True(t, bin.contains("", "/base/testunit/jemoeder.txt"))
	assert.True(t, bin.contains("jemoeder", "/base/testunit/jemoeder.txt"))
	assert.False(t, bin.contains("jemoeder", "/base/testunit/opjehoofd/jemoeder.txt"))
	assert.False(t, bin.contains("", "/etc/passwd"))
	assert.False(t, bin.contains("/", "/etc/passwd"))
	assert.False(t, bin.contains("/etc", "/etc/passwd"))
}

func TestFilePermissions(t *testing.T) {
	dirname, err := os.MkdirTemp("", "file-permission-test")
	assert.Nil(t, err)
	defer os.RemoveAll(dirname)

	bin := storageBin{
		basePath:      dirname,
		dirName:       "testunit",
		hasTempSuffix: true,
	}

	file, err := bin.openForWriting("testfilename.blend")
	assert.Nil(t, err)
	defer file.Close()

	filestat, err := file.Stat()
	assert.Nil(t, err)

	// The exact permissions depend on the current (unittest) process umask. This
	// umask is not easy to get, which is why we have a copy of `tempfile.go` in
	// the first place. The important part is that the permissions shouldn't be
	// the default 0600 created by ioutil.TempFile() but something more permissive
	// and dependent on the umask.
	fileMode := uint32(filestat.Mode())
	assert.True(t, fileMode > 0600,
		"Expecting more open permissions than 0o600, got %O", fileMode)

	groupWorldMode := fileMode & 0077
	assert.True(t, groupWorldMode < 0066,
		"Expecting tighter group+world permissions than wide-open 0o66, got %O. "+
			"Note that this test expects a non-zero umask.", groupWorldMode)
}
