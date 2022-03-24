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

package checkout

import (
	"context"
	"testing"

	"git.blender.org/flamenco/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestReportRequirements(t *testing.T) {
	manager, cleanup := createTestManager()
	defer cleanup()

	spec1 := api.ShamanFileSpec{Sha: "63b72c63b9424fd13b9370fb60069080c3a15717cf3ad442635b187c6a895079", Size: 127, Path: "file1.txt"}
	spec2 := api.ShamanFileSpec{Sha: "9f1470441beb98dbb66e3339e7da697d9c2312999a6a5610c461cbf55040e210", Size: 795, Path: "file2.txt"}
	spec3 := api.ShamanFileSpec{Sha: "59c6bd72af62aa860343adcafd46e3998934a9db2997ce08514b4361f099fa58", Size: 1134, Path: "file3.txt"}
	spec4 := api.ShamanFileSpec{Sha: "59c6bd72af62aa860343adcafd46e3998934a9db2997ce08514b4361f099fa58", Size: 1134, Path: "file4.txt"} // duplicate of the above

	required := api.ShamanRequirementsRequest{
		Files: []api.ShamanFileSpec{spec1, spec2, spec3, spec4},
	}

	response, err := manager.ReportRequirements(context.Background(), required)
	assert.NoError(t, err)

	// We should not be required to upload the same file twice, so the duplicate
	// should not be in the response.
	assert.Equal(t, []api.ShamanFileSpecWithStatus{
		{ShamanFileSpec: spec1, Status: api.ShamanFileStatusUnknown},
		{ShamanFileSpec: spec2, Status: api.ShamanFileStatusUnknown},
		{ShamanFileSpec: spec3, Status: api.ShamanFileStatusUnknown},
	}, response.Files)
}

// func TestCreateCheckout(t *testing.T) {
// 	manager, cleanup := createTestManager()
// 	defer cleanup()

// 	filestore.LinkTestFileStore(manager.fileStore.BasePath())

// 	defFile, err := ioutil.ReadFile("../_test_file_store/checkout_definition.txt")
// 	assert.Nil(t, err)
// 	compressedDefFile := httpserver.CompressBuffer(defFile)

// 	respRec := httptest.NewRecorder()
// 	req := httptest.NewRequest("POST", "/checkout/create/{checkoutID}", compressedDefFile)
// 	req = mux.SetURLVars(req, map[string]string{
// 		"checkoutID": "jemoeder",
// 	})
// 	req.Header.Set("Content-Type", "text/plain")
// 	req.Header.Set("Content-Encoding", "gzip")
// 	logrus.SetLevel(logrus.DebugLevel)
// 	manager.createCheckout(respRec, req)

// 	bodyBytes, err := ioutil.ReadAll(respRec.Body)
// 	assert.Nil(t, err)
// 	body := string(bodyBytes)
// 	assert.Equal(t, http.StatusOK, respRec.Code, body)

// 	// Check the symlinks of the checkout
// 	coPath := path.Join(manager.checkoutBasePath, "er", "jemoeder")
// 	assert.FileExists(t, path.Join(coPath, "subdir", "replacer.py"))
// 	assert.FileExists(t, path.Join(coPath, "feed.py"))
// 	assert.FileExists(t, path.Join(coPath, "httpstuff.py"))
// 	assert.FileExists(t, path.Join(coPath, "filesystemstuff.py"))

// 	storePath := manager.fileStore.StoragePath()
// 	assertLinksTo(t, path.Join(coPath, "subdir", "replacer.py"),
// 		path.Join(storePath, "59", "0c148428d5c35fab3ebad2f3365bb469ab9c531b60831f3e826c472027a0b9", "3367.blob"))
// 	assertLinksTo(t, path.Join(coPath, "feed.py"),
// 		path.Join(storePath, "80", "b749c27b2fef7255e7e7b3c2029b03b31299c75ff1f1c72732081c70a713a3", "7488.blob"))
// 	assertLinksTo(t, path.Join(coPath, "httpstuff.py"),
// 		path.Join(storePath, "91", "4853599dd2c351ab7b82b219aae6e527e51518a667f0ff32244b0c94c75688", "486.blob"))
// 	assertLinksTo(t, path.Join(coPath, "filesystemstuff.py"),
// 		path.Join(storePath, "d6", "fc7289b5196cc96748ea72f882a22c39b8833b457fe854ef4c03a01f5db0d3", "7217.blob"))
// }

// func assertLinksTo(t *testing.T, linkPath, expectedTarget string) {
// 	actualTarget, err := os.Readlink(linkPath)
// 	assert.Nil(t, err)
// 	assert.Equal(t, expectedTarget, actualTarget)
// }
