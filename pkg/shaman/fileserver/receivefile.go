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
	"context"
	"errors"
	"fmt"
	"io"

	"git.blender.org/flamenco/pkg/shaman/filestore"
	"git.blender.org/flamenco/pkg/shaman/hasher"
	"github.com/rs/zerolog"
)

// ErrFileAlreadyExists indicates that a file already exists in the Shaman
// storage. It can also be returned during upload, when someone else succesfully
// uploaded the same file at the same time.
var ErrFileAlreadyExists = errors.New("uploaded file already exists")

type ErrFileSizeMismatch struct {
	DeclaredSize int64
	ActualSize   int64
}

func (e ErrFileSizeMismatch) Error() string {
	return fmt.Sprintf("file size mismatched, declared %d but received %d bytes", e.DeclaredSize, e.ActualSize)
}

type ErrFileChecksumMismatch struct {
	DeclaredChecksum string
	ActualChecksum   string
}

func (e ErrFileChecksumMismatch) Error() string {
	return fmt.Sprintf("file SHA256 mismatched, declared %s but received %s", e.DeclaredChecksum, e.ActualChecksum)
}

// ReceiveFile streams a file from a HTTP request to disk.
func (fs *FileServer) ReceiveFile(
	ctx context.Context, bodyReader io.ReadCloser,
	checksum string, filesize int64, canDefer bool,
) error {
	logger := *zerolog.Ctx(ctx)
	defer bodyReader.Close()

	localPath, status := fs.fileStore.ResolveFile(checksum, filesize, filestore.ResolveEverything)
	logger = logger.With().
		Str("path", localPath).
		Str("checksum", checksum).
		Int64("filesize", filesize).
		Str("status", status.String()).
		Logger()

	switch status {
	case filestore.StatusStored:
		logger.Info().Msg("shaman: uploaded file already exists")
		return ErrFileAlreadyExists
	case filestore.StatusUploading:
		if canDefer {
			logger.Info().Msg("shaman: someone is uploading this file and client can defer")
			return ErrFileAlreadyExists
		}
	}

	logger.Info().Msg("shaman: receiving file")

	streamTo, err := fs.fileStore.OpenForUpload(checksum, filesize)
	if err != nil {
		return fmt.Errorf("opening file for writing uploaded data: %w", err)
	}

	// Clean up temporary file if it still exists at function exit.
	defer func() {
		streamTo.Close()
		fs.fileStore.RemoveUploadedFile(streamTo.Name())
	}()

	// Abort this upload when the file has been finished by someone else.
	uploadDone := make(chan struct{})
	uploadAlreadyCompleted := false
	defer close(uploadDone)
	receiverChannel := fs.receiveListenerFor(checksum, filesize)
	go func() {
		select {
		case <-uploadDone:
			close(receiverChannel)
			return
		case <-receiverChannel:
		}
		logger.Info().Msg("file was completed during someone else's upload")

		uploadAlreadyCompleted = true
		err := bodyReader.Close()
		if err != nil {
			logger.Warn().Err(err).Msg("error closing connection")
		}
	}()

	// TODO: pass context to hasher.Copy()
	written, actualChecksum, err := hasher.Copy(streamTo, bodyReader)
	if err != nil {
		if closeErr := streamTo.Close(); closeErr != nil {
			logger.Error().
				AnErr("copyError", err).
				AnErr("closeError", closeErr).
				Msg("error closing local file after other I/O error occured")
		}

		logger = logger.With().Err(err).Logger()
		switch {
		case uploadAlreadyCompleted:
			logger.Debug().Msg("aborted upload")
			return ErrFileAlreadyExists
		case err == io.ErrUnexpectedEOF:
			logger.Debug().Msg("unexpected EOF, client probably just disconnected")
			return err
		default:
			return fmt.Errorf("unable to copy request body to file: %w", err)
		}
	}

	if err := streamTo.Close(); err != nil {
		return fmt.Errorf("closing local file: %w", err)
	}

	if written != filesize {
		logger.Warn().
			Int64("declaredSize", filesize).
			Int64("actualSize", written).
			Msg("mismatch between expected and actual size")
		return ErrFileSizeMismatch{
			DeclaredSize: filesize,
			ActualSize:   written,
		}
	}

	if actualChecksum != checksum {
		logger.Warn().
			Str("declaredChecksum", checksum).
			Str("actualChecksum", actualChecksum).
			Msg("mismatch between expected and actual checksum")
		return ErrFileChecksumMismatch{
			DeclaredChecksum: checksum,
			ActualChecksum:   actualChecksum,
		}
	}

	logger.Debug().
		Int64("receivedBytes", written).
		Str("checksum", actualChecksum).
		Str("tempFile", streamTo.Name()).
		Msg("File received")

	if err := fs.fileStore.MoveToStored(checksum, filesize, streamTo.Name()); err != nil {
		logger.Error().
			Err(err).
			Str("tempFile", streamTo.Name()).
			Msg("unable to move file from 'upload' to 'stored' storage")
		return err
	}

	return nil
}
