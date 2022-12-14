/* (c) 2019, Blender Foundation
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

package touch

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTouch(t *testing.T) {
	testPath := "_touch_test.txt"

	// Create a file
	assert.Nil(t, ioutil.WriteFile(testPath, []byte("just a test"), 0644))
	defer os.Remove(testPath)

	// Make it old
	past := time.Now().Add(-5 * time.Hour)
	assert.Nil(t, os.Chtimes(testPath, past, past))

	// Touch & test
	assert.Nil(t, Touch(testPath))

	stat, err := os.Stat(testPath)
	assert.NoError(t, err)

	threshold := time.Now().Add(-5 * time.Second)
	assert.True(t, stat.ModTime().After(threshold),
		"mtime should be after %v but is %v", threshold, stat.ModTime())
}
