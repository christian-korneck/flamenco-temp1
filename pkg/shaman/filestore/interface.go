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
	"errors"
	"fmt"
)

// FileStatus represents the status of a file in the store.
type FileStatus int

// Valid statuses for files in the store.
const (
	StatusNotSet FileStatus = iota
	StatusDoesNotExist
	StatusUploading
	StatusStored
)

func (fs FileStatus) String() string {
	switch fs {
	case StatusDoesNotExist:
		return "DoesNotExist"
	case StatusUploading:
		return "Uploading"
	case StatusStored:
		return "Stored"
	default:
		return fmt.Sprintf("invalid(%d)", int(fs))
	}
}

// StoredOnly indicates whether to resolve only 'stored' files or also 'uploading' or 'checking'.
type StoredOnly bool

// For the ResolveFile() call. This is more explicit than just true/false values.
const (
	ResolveStoredOnly StoredOnly = true
	ResolveEverything StoredOnly = false
)

// Predefined errors
var (
	ErrFileDoesNotExist = errors.New("file does not exist")
	ErrNotInUploading   = errors.New("file not stored in 'uploading' storage")
)
