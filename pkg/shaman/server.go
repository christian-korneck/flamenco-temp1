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
	"sync"

	"git.blender.org/flamenco/pkg/shaman/checkout"
	"git.blender.org/flamenco/pkg/shaman/config"
	"git.blender.org/flamenco/pkg/shaman/fileserver"
	"git.blender.org/flamenco/pkg/shaman/filestore"
	"git.blender.org/flamenco/pkg/shaman/httpserver"
	"git.blender.org/flamenco/pkg/shaman/jwtauth"
	"github.com/gorilla/mux"
)

// Server represents a Shaman Server.
type Server struct {
	config config.Config

	auther      jwtauth.Authenticator
	fileStore   filestore.Storage
	fileServer  *fileserver.FileServer
	checkoutMan *checkout.Manager

	shutdownChan chan struct{}
	wg           sync.WaitGroup
}

// NewServer creates a new Shaman server.
func NewServer(conf config.Config, auther jwtauth.Authenticator) *Server {

	if !conf.Enabled {
		packageLogger.Warning("Shaman server is disabled")
		return nil
	}

	fileStore := filestore.New(conf)
	checkoutMan := checkout.NewManager(conf, fileStore)
	fileServer := fileserver.New(fileStore)

	shamanServer := &Server{
		conf,
		auther,
		fileStore,
		fileServer,
		checkoutMan,

		make(chan struct{}),
		sync.WaitGroup{},
	}

	return shamanServer
}

// Go starts goroutines for background operations.
// After Go() has been called, use Close() to stop those goroutines.
func (s *Server) Go() {
	packageLogger.Info("Shaman server starting")
	s.fileServer.Go()

	if s.config.GarbageCollect.Period == 0 {
		packageLogger.Warning("garbage collection disabled, set garbageCollect.period > 0 in configuration")
	} else if s.config.GarbageCollect.SilentlyDisable {
		packageLogger.Debug("not starting garbage collection")
	} else {
		s.wg.Add(1)
		go s.periodicCleanup()
	}
}

// AddRoutes adds the Shaman server endpoints to the given router.
func (s *Server) AddRoutes(router *mux.Router) {
	s.checkoutMan.AddRoutes(router, s.auther)
	s.fileServer.AddRoutes(router, s.auther)

	httpserver.RegisterTestRoutes(router, s.auther)
}

// Close shuts down the Shaman server.
func (s *Server) Close() {
	packageLogger.Info("shutting down Shaman server")
	close(s.shutdownChan)
	s.fileServer.Close()
	s.checkoutMan.Close()
	s.wg.Wait()
}
