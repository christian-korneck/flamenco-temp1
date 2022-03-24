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

package shaman

import (
	"context"
	"io"
	"sync"

	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/shaman/checkout"
	"git.blender.org/flamenco/pkg/shaman/config"
	"git.blender.org/flamenco/pkg/shaman/fileserver"
	"git.blender.org/flamenco/pkg/shaman/filestore"
	"git.blender.org/flamenco/pkg/shaman/jwtauth"
	"github.com/rs/zerolog/log"
)

// Server represents a Shaman Server.
type Server struct {
	config config.Config

	auther      jwtauth.Authenticator
	fileStore   *filestore.Store
	fileServer  *fileserver.FileServer
	checkoutMan *checkout.Manager

	shutdownChan chan struct{}
	wg           sync.WaitGroup
}

// NewServer creates a new Shaman server.
func NewServer(conf config.Config, auther jwtauth.Authenticator) *Server {
	if !conf.Enabled {
		log.Info().Msg("shaman server is disabled")
		return nil
	}

	if conf.CheckoutPath == "" {
		log.Error().Interface("config", conf).Msg("shaman: no checkout path configured, unable to start")
		return nil
	}

	fileStore := filestore.New(conf)
	checkoutMan := checkout.NewManager(conf, fileStore)
	fileServer := fileserver.New(fileStore)

	shamanServer := &Server{
		config:      conf,
		auther:      auther,
		fileStore:   fileStore,
		fileServer:  fileServer,
		checkoutMan: checkoutMan,

		shutdownChan: make(chan struct{}),
		wg:           sync.WaitGroup{},
	}

	return shamanServer
}

// Go starts goroutines for background operations.
// After Go() has been called, use Close() to stop those goroutines.
func (s *Server) Go() {
	log.Info().Msg("Shaman server starting")
	s.fileServer.Go()

	if s.config.GarbageCollect.Period == 0 {
		log.Warn().Msg("garbage collection disabled, set garbageCollect.period > 0 in configuration")
	} else if s.config.GarbageCollect.SilentlyDisable {
		log.Debug().Msg("not starting garbage collection")
	} else {
		s.wg.Add(1)
		go s.periodicCleanup()
	}
}

// Close shuts down the Shaman server.
func (s *Server) Close() {
	log.Info().Msg("shutting down Shaman server")

	close(s.shutdownChan)

	s.fileServer.Close()
	s.checkoutMan.Close()
	s.wg.Wait()
}

func (s *Server) IsEnabled() bool {
	return s != nil && s.config.Enabled
}

// Checkout creates a directory, and symlinks the required files into it. The
// files must all have been uploaded to Shaman before calling this.
func (s *Server) Checkout(ctx context.Context, checkout api.ShamanCheckout) error {
	return s.checkoutMan.Checkout(ctx, checkout)
}

// Requirements checks a Shaman Requirements file, and returns the subset
// containing the unknown files.
func (s *Server) Requirements(ctx context.Context, requirements api.ShamanRequirementsRequest) (api.ShamanRequirementsResponse, error) {
	return s.checkoutMan.ReportRequirements(ctx, requirements)
}

var fsStatusToApiStatus = map[filestore.FileStatus]api.ShamanFileStatus{
	filestore.StatusDoesNotExist: api.ShamanFileStatusUnknown,
	filestore.StatusUploading:    api.ShamanFileStatusUploading,
	filestore.StatusStored:       api.ShamanFileStatusStored,
}

// Check the status of a file on the Shaman server.
// status (stored, currently being uploaded, unknown).
func (s *Server) FileStoreCheck(ctx context.Context, checksum string, filesize int64) api.ShamanFileStatus {
	status := s.fileServer.CheckFile(checksum, filesize)
	apiStatus, ok := fsStatusToApiStatus[status]
	if !ok {
		log.Warn().
			Str("checksum", checksum).
			Int64("filesize", filesize).
			Int("fileserverStatus", int(status)).
			Msg("shaman: unknown status on fileserver")
		return api.ShamanFileStatusUnknown
	}
	return apiStatus
}

// Store a new file on the Shaman server. Note that the Shaman server can return
// early when another client finishes uploading the exact same file, to prevent
// double uploads.
func (s *Server) FileStore(ctx context.Context, file io.ReadCloser, checksum string, filesize int64, canDefer bool, originalFilename string) error {
	err := s.fileServer.ReceiveFile(ctx, file, checksum, filesize, canDefer, originalFilename)
	// TODO: Maybe translate this error into something that can be understood by
	// the caller without relying on types declared in the `fileserver` package?
	return err
}
