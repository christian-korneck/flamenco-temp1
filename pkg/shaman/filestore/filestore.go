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
	"path"
	"strconv"

	"git.blender.org/flamenco/pkg/shaman/config"
	"github.com/rs/zerolog/log"
)

// Store represents the default Shaman file store.
type Store struct {
	baseDir string

	uploading storageBin
	stored    storageBin
}

// New returns a new file store.
func New(conf config.Config) *Store {
	log.Info().Str("storageDir", conf.FileStorePath).Msg("shaman: opening file store")
	store := &Store{
		conf.FileStorePath,
		storageBin{conf.FileStorePath, "uploading", true, ".tmp"},
		storageBin{conf.FileStorePath, "stored", false, ".blob"},
	}
	store.createDirectoryStructure()
	return store
}

// Create the base directory structure for this store.
func (s *Store) createDirectoryStructure() {
	mkdir := func(subdir string) {
		path := path.Join(s.baseDir, subdir)

		logger := log.With().Str("path", path).Logger()
		logger.Debug().Msg("shaman: creating directory")

		if err := os.MkdirAll(path, 0777); err != nil {
			if os.IsExist(err) {
				logger.Trace().Msg("shaman: directory exists")
				return
			}
			logger.Error().Err(err).Msg("shaman: unable to create directory")
		}
	}

	mkdir(s.uploading.dirName)
	mkdir(s.stored.dirName)
}

// StoragePath returns the directory path of the 'stored' storage bin.
func (s *Store) StoragePath() string {
	return path.Join(s.stored.basePath, s.stored.dirName)
}

// BasePath returns the directory path of the storage.
func (s *Store) BasePath() string {
	return s.baseDir
}

// Returns the checksum/filesize dependent parts of the file's path.
// To be combined with a base directory, status directory, and status-dependent suffix.
func (s *Store) partialFilePath(checksum string, filesize int64) string {
	return path.Join(checksum[0:2], checksum[2:], strconv.FormatInt(filesize, 10))
}

// ResolveFile checks the status of the file in the store.
func (s *Store) ResolveFile(checksum string, filesize int64, storedOnly StoredOnly) (path string, status FileStatus) {
	partial := s.partialFilePath(checksum, filesize)

	logger := log.With().
		Str("checksum", checksum).
		Int64("filesize", filesize).
		Str("partialPath", partial).
		Str("storagePath", s.baseDir).
		Logger()

	if path = s.stored.resolve(partial); path != "" {
		logger.Trace().Str("path", path).Msg("shaman: found stored file")
		return path, StatusStored
	}

	if storedOnly != ResolveEverything {
		logger.Trace().Msg("shaman: file does not exist in 'stored' state")
		return "", StatusDoesNotExist
	}

	if path = s.uploading.resolve(partial); path != "" {
		logger.Debug().Str("path", path).Msg("shaman: found currently uploading file")
		return path, StatusUploading
	}

	logger.Trace().Msg("shaman: file does not exist")
	return "", StatusDoesNotExist
}

// OpenForUpload returns a file pointer suitable to stream an uploaded file to.
func (s *Store) OpenForUpload(checksum string, filesize int64) (*os.File, error) {
	partial := s.partialFilePath(checksum, filesize)
	return s.uploading.openForWriting(partial)
}

// MoveToStored moves a file from 'uploading' to 'stored' storage.
// It is assumed that the checksum and filesize have been verified.
func (s *Store) MoveToStored(checksum string, filesize int64, uploadedFilePath string) error {
	// Check that the uploaded file path is actually in the 'uploading' storage.
	partial := s.partialFilePath(checksum, filesize)
	if !s.uploading.contains(partial, uploadedFilePath) {
		return ErrNotInUploading
	}

	// Move to the other storage bin.
	targetPath := s.stored.pathFor(partial)
	targetDir, _ := path.Split(targetPath)
	if err := os.MkdirAll(targetDir, 0777); err != nil {
		return err
	}
	log.Debug().
		Str("uploadedPath", uploadedFilePath).
		Str("storagePath", targetPath).
		Msg("shaman: moving uploaded file to storage")
	if err := os.Rename(uploadedFilePath, targetPath); err != nil {
		return err
	}

	s.RemoveUploadedFile(uploadedFilePath)
	return nil
}

func (s *Store) removeFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Debug().Err(err).Msg("shaman: unable to delete file; ignoring")
		}
	}

	// Clean up directory structure, but ignore any errors (dirs may not be empty)
	directory := path.Dir(filePath)
	os.Remove(directory)
	os.Remove(path.Dir(directory))

	return err
}

// RemoveUploadedFile removes a file from the 'uploading' storage bin.
// Errors are ignored.
func (s *Store) RemoveUploadedFile(filePath string) {
	// Check that the file path is actually in the 'uploading' storage.
	if !s.uploading.contains("", filePath) {
		log.Error().Str("file", filePath).
			Msg("shaman: RemoveUploadedFile called with file not in 'uploading' storage bin")
		return
	}
	s.removeFile(filePath)
}

// RemoveStoredFile removes a file from the 'stored' storage bin.
func (s *Store) RemoveStoredFile(filePath string) error {
	// Check that the file path is actually in the 'stored' storage.
	if !s.stored.contains("", filePath) {
		log.Error().Str("file", filePath).
			Msg("shaman: RemoveStoredFile called with file not in 'stored' storage bin")
		return os.ErrNotExist
	}
	return s.removeFile(filePath)
}
