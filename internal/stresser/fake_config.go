package stresser

import (
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/worker"
)

type FakeConfig struct {
	creds worker.WorkerCredentials
}

func NewFakeConfig(workerID, workerSecret string) *FakeConfig {
	return &FakeConfig{
		creds: worker.WorkerCredentials{
			WorkerID: workerID,
			Secret:   workerSecret,
		},
	}
}

func (fc *FakeConfig) WorkerConfig() (worker.WorkerConfig, error) {
	config := worker.NewConfigWrangler().DefaultConfig()
	config.ManagerURL = "http://localhost:8080/"
	return config, nil
}

func (fc *FakeConfig) WorkerCredentials() (worker.WorkerCredentials, error) {
	return fc.creds, nil
}

func (fc *FakeConfig) SaveCredentials(creds worker.WorkerCredentials) error {
	log.Info().
		Str("workerID", creds.WorkerID).
		Str("workerSecret", creds.Secret).
		Msg("remember these credentials for next time")
	return nil
}
