package worker

import (
	"context"
	"errors"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

const (
	requestRetry        = 5 * time.Second
	credentialsFilename = "flamenco-worker-credentials.yaml"
	configFilename      = "flamenco-worker.yaml"
)

var (
	errRequestAborted = errors.New("request to Manager aborted")
)

// Worker performs regular Flamenco Worker operations.
type Worker struct {
	doneChan chan struct{}
	doneWg   *sync.WaitGroup

	manager *url.URL
	client  api.ClientWithResponsesInterface
	creds   *workerCredentials

	state         api.WorkerStatus
	stateStarters map[string]func() // gotoStateXXX functions
	stateMutex    *sync.Mutex

	taskRunner TaskRunner

	configWrangler ConfigWrangler
	config         WorkerConfig
	managerFinder  ManagerFinder
}

type ManagerFinder interface {
	FindFlamencoManager() <-chan *url.URL
}

type TaskRunner interface{}

// NewWorker constructs and returns a new Worker.
func NewWorker(
	flamenco api.ClientWithResponsesInterface,
	configWrangler ConfigWrangler,
	managerFinder ManagerFinder,
	taskRunner TaskRunner,
) *Worker {

	worker := &Worker{
		doneChan: make(chan struct{}),
		doneWg:   new(sync.WaitGroup),

		client: flamenco,

		state:         api.WorkerStatusStarting,
		stateStarters: make(map[string]func()),
		stateMutex:    new(sync.Mutex),

		// taskRunner: taskRunner,

		configWrangler: configWrangler,
		managerFinder:  managerFinder,
	}
	// worker.setupStateMachine()
	worker.loadConfig()
	return worker
}

func (w *Worker) start(ctx context.Context, register bool) {
	w.doneWg.Add(1)
	defer w.doneWg.Done()

	w.loadCredentials()

	if w.creds == nil || register {
		w.register(ctx)
	}

	startState := w.signOn(ctx)
	log.Error().Str("state", string(startState)).Msg("here the road ends, nothing else is implemented")
	// w.changeState(startState)
}

func (w *Worker) loadCredentials() {
	log.Debug().Msg("loading credentials")

	w.creds = &workerCredentials{}
	err := w.configWrangler.LoadConfig(credentialsFilename, w.creds)
	if err != nil {
		log.Warn().Err(err).Str("file", credentialsFilename).
			Msg("unable to load credentials configuration file")
		w.creds = nil
		return
	}
}

func (w *Worker) loadConfig() {
	logger := log.With().Str("filename", configFilename).Logger()
	err := w.configWrangler.LoadConfig(configFilename, &w.config)
	if os.IsNotExist(err) {
		logger.Info().Msg("writing default configuration file")
		w.config = w.configWrangler.DefaultConfig()
		w.saveConfig()
		err = w.configWrangler.LoadConfig(configFilename, &w.config)
	}
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to load config file")
	}

	if w.config.Manager != "" {
		w.manager, err = ParseURL(w.config.Manager)
		if err != nil {
			logger.Fatal().Err(err).Str("url", w.config.Manager).
				Msg("unable to parse manager URL")
		}
		logger.Debug().Str("url", w.config.Manager).Msg("parsed manager URL")
	}

}

func (w *Worker) saveConfig() {
	err := w.configWrangler.WriteConfig(configFilename, "Configuration", w.config)
	if err != nil {
		log.Warn().Err(err).Str("filename", configFilename).
			Msg("unable to write configuration file")
	}
}

// Close gracefully shuts down the Worker.
func (w *Worker) Close() {
	log.Debug().Msg("worker gracefully shutting down")
	close(w.doneChan)
	w.doneWg.Wait()
	log.Debug().Msg("worker shut down")
}
