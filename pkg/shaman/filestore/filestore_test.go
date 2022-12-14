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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mustCreateFile creates an empty file.
// The containing directory structure is created as well, if necessary.
func mustCreateFile(file_path string) {
	err := os.MkdirAll(filepath.Dir(file_path), 0777)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(file_path)
	if err != nil {
		panic(err)
	}
	file.Close()
}

func TestCreateDirectories(t *testing.T) {
	store := CreateTestStore()
	defer CleanupTestStore(store)

	assert.Equal(t, filepath.Join(store.baseDir, "uploading", "x"), store.uploading.storagePrefix("x"))
	assert.Equal(t, filepath.Join(store.baseDir, "stored", "x"), store.stored.storagePrefix("x"))

	assert.DirExists(t, filepath.Join(store.baseDir, "uploading"))
	assert.DirExists(t, filepath.Join(store.baseDir, "stored"))
}

func TestResolveStoredFile(t *testing.T) {
	store := CreateTestStore()
	defer CleanupTestStore(store)

	foundPath, status := store.ResolveFile("abcdefxxx", 123, ResolveStoredOnly)
	assert.Equal(t, "", foundPath)
	assert.Equal(t, StatusDoesNotExist, status)

	fname := filepath.Join(store.baseDir, "stored", "ab", "cdefxxx", "123.blob")
	mustCreateFile(fname)

	foundPath, status = store.ResolveFile("abcdefxxx", 123, ResolveStoredOnly)
	assert.Equal(t, fname, foundPath)
	assert.Equal(t, StatusStored, status)

	foundPath, status = store.ResolveFile("abcdefxxx", 123, ResolveEverything)
	assert.Equal(t, fname, foundPath)
	assert.Equal(t, StatusStored, status)
}

func TestResolveUploadingFile(t *testing.T) {
	store := CreateTestStore()
	defer CleanupTestStore(store)

	foundPath, status := store.ResolveFile("abcdefxxx", 123, ResolveEverything)
	assert.Equal(t, "", foundPath)
	assert.Equal(t, StatusDoesNotExist, status)

	fname := filepath.Join(store.baseDir, "uploading", "ab", "cdefxxx", "123-unique-code.tmp")
	mustCreateFile(fname)

	foundPath, status = store.ResolveFile("abcdefxxx", 123, ResolveStoredOnly)
	assert.Equal(t, "", foundPath)
	assert.Equal(t, StatusDoesNotExist, status)

	foundPath, status = store.ResolveFile("abcdefxxx", 123, ResolveEverything)
	assert.Equal(t, fname, foundPath)
	assert.Equal(t, StatusUploading, status)
}

func TestOpenForUpload(t *testing.T) {
	store := CreateTestStore()
	defer CleanupTestStore(store)

	contents := []byte("je moešje")
	fileSize := int64(len(contents))

	file, err := store.OpenForUpload("abcdefxxx", fileSize)
	assert.NoError(t, err)
	_, err = file.Write(contents)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())

	foundPath, status := store.ResolveFile("abcdefxxx", fileSize, ResolveEverything)
	assert.Equal(t, file.Name(), foundPath)
	assert.Equal(t, StatusUploading, status)

	readContents, err := ioutil.ReadFile(foundPath)
	assert.NoError(t, err)
	assert.EqualValues(t, contents, readContents)
}

func TestMoveToStored(t *testing.T) {
	store := CreateTestStore()
	defer CleanupTestStore(store)

	contents := []byte("je moešje")
	fileSize := int64(len(contents))

	err := store.MoveToStored("abcdefxxx", fileSize, "/just/some/path")
	assert.Error(t, err)

	file, err := store.OpenForUpload("abcdefxxx", fileSize)
	assert.NoError(t, err)
	_, err = file.Write(contents)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())
	tempLocation := file.Name()

	err = store.MoveToStored("abcdefxxx", fileSize, file.Name())
	assert.NoError(t, err, "moving file %s", file.Name())

	foundPath, status := store.ResolveFile("abcdefxxx", fileSize, ResolveEverything)
	assert.NotEqual(t, file.Name(), foundPath)
	assert.Equal(t, StatusStored, status)

	assert.FileExists(t, foundPath)

	// The entire directory structure should be kept clean.
	assert.NoFileExists(t, tempLocation)
	assert.NoDirExists(t, filepath.Dir(tempLocation))
	assert.NoDirExists(t, filepath.Dir(filepath.Dir(tempLocation)))
}
