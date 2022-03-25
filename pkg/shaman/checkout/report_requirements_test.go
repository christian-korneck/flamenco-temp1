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
		{Sha: spec1.Sha, Size: spec1.Size, Path: spec1.Path, Status: api.ShamanFileStatusUnknown},
		{Sha: spec2.Sha, Size: spec2.Size, Path: spec2.Path, Status: api.ShamanFileStatusUnknown},
		{Sha: spec3.Sha, Size: spec3.Size, Path: spec3.Path, Status: api.ShamanFileStatusUnknown},
	}, response.Files)
}
