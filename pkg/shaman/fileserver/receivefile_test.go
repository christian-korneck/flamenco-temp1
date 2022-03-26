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

package fileserver

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"

	"git.blender.org/flamenco/pkg/shaman/config"
	"git.blender.org/flamenco/pkg/shaman/hasher"

	"git.blender.org/flamenco/pkg/shaman/filestore"
	"github.com/stretchr/testify/assert"
)

func TestStoreFile(t *testing.T) {
	server, cleanup := createTestServer()
	defer cleanup()

	payload := []byte("hähähä")
	// Just to double-check it's encoded as UTF-8:
	assert.EqualValues(t, []byte("h\xc3\xa4h\xc3\xa4h\xc3\xa4"), payload)

	filesize := int64(len(payload))
	correctChecksum := hasher.Checksum(payload)

	testWithChecksum := func(checksum string, reportSize int64) error {
		buffer := io.NopCloser(bytes.NewBuffer(payload))
		return server.ReceiveFile(context.Background(), buffer, checksum, reportSize, false, "testfile.txt")
	}

	var err error
	var path string
	var status filestore.FileStatus

	// A bad checksum should be rejected.
	badChecksum := "da-checksum-is-long-enough-like-this"
	err = testWithChecksum(badChecksum, filesize)
	assert.ErrorIs(t, err, ErrFileChecksumMismatch{
		DeclaredChecksum: badChecksum,
		ActualChecksum:   correctChecksum,
	})
	path, status = server.fileStore.ResolveFile(badChecksum, filesize, filestore.ResolveEverything)
	assert.Equal(t, filestore.StatusDoesNotExist, status)
	assert.Equal(t, "", path)

	// A bad file size should be rejected.
	err = testWithChecksum(correctChecksum, filesize+1)
	assert.ErrorIs(t, err, ErrFileSizeMismatch{
		DeclaredSize: filesize + 1,
		ActualSize:   filesize,
	})
	path, status = server.fileStore.ResolveFile(badChecksum, filesize, filestore.ResolveEverything)
	assert.Equal(t, filestore.StatusDoesNotExist, status)
	assert.Equal(t, "", path)

	// The correct checksum should be accepted.
	err = testWithChecksum(correctChecksum, filesize)
	assert.NoError(t, err)

	path, status = server.fileStore.ResolveFile(correctChecksum, filesize, filestore.ResolveEverything)
	assert.Equal(t, filestore.StatusStored, status)
	assert.FileExists(t, path)

	savedContent, err := ioutil.ReadFile(path)
	assert.NoError(t, err)
	assert.EqualValues(t, payload, savedContent, "The file should be saved uncompressed")
}

func createTestServer() (server *FileServer, cleanup func()) {
	config, configCleanup := config.CreateTestConfig()

	store := filestore.New(config)
	server = New(store)
	server.Go()

	cleanup = func() {
		server.Close()
		configCleanup()
	}
	return
}
