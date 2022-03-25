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
	"bufio"
	"context"
	"fmt"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

/* Checkout Definition files contain a line for each to-be-checked-out file.
 * Each line consists of three fields: checksum, file size, path in the checkout.
 */

// FileInvalidError is returned when there is an invalid line in a checkout definition file.
type FileInvalidError struct {
	lineNumber int // base-1 line number that's bad
	innerErr   error
	reason     string
}

func (cfie FileInvalidError) Error() string {
	return fmt.Sprintf("invalid line %d: %s", cfie.lineNumber, cfie.reason)
}

// DefinitionLine is a single line in a checkout definition file.
type DefinitionLine struct {
	Checksum string
	FileSize int64
	FilePath string
}

// DefinitionReader reads and parses a checkout definition
type DefinitionReader struct {
	ctx     context.Context
	channel chan *DefinitionLine
	reader  *bufio.Reader

	Err        error
	LineNumber int
}

var (
	// This is a wider range than used in SHA256 sums, but there is no harm in accepting a few more ASCII letters.
	validChecksumRegexp = regexp.MustCompile("^[a-zA-Z0-9]{16,}$")
)

// NewDefinitionReader creates a new DefinitionReader for the given reader.
func NewDefinitionReader(ctx context.Context, reader io.Reader) *DefinitionReader {
	return &DefinitionReader{
		ctx:     ctx,
		channel: make(chan *DefinitionLine),
		reader:  bufio.NewReader(reader),
	}
}

// Read spins up a new goroutine for parsing the checkout definition.
// The returned channel will receive definition lines.
func (fr *DefinitionReader) Read() <-chan *DefinitionLine {
	go func() {
		defer close(fr.channel)
		defer logrus.Debug("done reading request")

		for {
			line, err := fr.reader.ReadString('\n')
			if err != nil && err != io.EOF {
				fr.Err = FileInvalidError{
					lineNumber: fr.LineNumber,
					innerErr:   err,
					reason:     fmt.Sprintf("I/O error: %v", err),
				}
				return
			}
			if err == io.EOF && line == "" {
				return
			}

			if contextError := fr.ctx.Err(); contextError != nil {
				fr.Err = fr.ctx.Err()
				return
			}

			fr.LineNumber++
			logrus.WithFields(logrus.Fields{
				"line":   line,
				"number": fr.LineNumber,
			}).Debug("read line")

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			definitionLine, err := fr.parseLine(line)
			if err != nil {
				fr.Err = err
				return
			}

			fr.channel <- definitionLine
		}
	}()

	return fr.channel
}

func (fr *DefinitionReader) parseLine(line string) (*DefinitionLine, error) {

	parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
	if len(parts) != 3 {
		return nil, FileInvalidError{
			lineNumber: fr.LineNumber,
			reason: fmt.Sprintf("line should consist of three space-separated parts, not %d: %v",
				len(parts), line),
		}
	}

	checksum := parts[0]
	if !validChecksumRegexp.MatchString(checksum) {
		return nil, FileInvalidError{fr.LineNumber, nil, "invalid checksum"}
	}

	fileSize, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, FileInvalidError{fr.LineNumber, err, "invalid file size"}
	}

	filePath := strings.TrimSpace(parts[2])
	if path.IsAbs(filePath) {
		return nil, FileInvalidError{fr.LineNumber, err, "no absolute paths allowed"}
	}
	if filePath != path.Clean(filePath) || strings.Contains(filePath, "..") {
		return nil, FileInvalidError{fr.LineNumber, err, "paths must be clean and not have any .. in them."}
	}

	return &DefinitionLine{
		Checksum: parts[0],
		FileSize: fileSize,
		FilePath: filePath,
	}, nil
}
