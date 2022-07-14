package config

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/rs/zerolog/log"
)

// Service provides access to Flamenco Manager configuration.
type Service struct {
	config        Conf
	forceFirstRun bool
}

func NewService() *Service {
	return &Service{
		config: DefaultConfig(),
	}
}

// IsFirstRun returns true if this is likely to be the first run of Flamenco.
func (s *Service) IsFirstRun() (bool, error) {
	if s.forceFirstRun {
		return true, nil
	}

	config, err := getConf()
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// No configuration means first run.
		return true, nil
	case err != nil:
		return false, fmt.Errorf("loading %s: %w", configFilename, err)
	}

	// No shared storage configured means first run.
	return config.SharedStoragePath == "", nil
}

func (s *Service) ForceFirstRun() {
	s.forceFirstRun = true
}

func (s *Service) Load() error {
	config, err := getConf()
	if err != nil {
		return err
	}
	s.config = config
	return nil
}

func (s *Service) Get() *Conf {
	return &s.config
}

// Save writes the in-memory configuration to the config file.
func (s *Service) Save() error {
	err := s.config.Write(configFilename)
	if err != nil {
		return err
	}

	// Do the logging here, as our caller doesn't know `configFilename``.
	log.Info().Str("filename", configFilename).Msg("configuration file written")
	return nil
}

// Expose some functions on Conf here, for easier mocking of functionality via interfaces.
//
func (s *Service) ExpandVariables(valueToExpand string, audience VariableAudience, platform VariablePlatform) string {
	return s.config.ExpandVariables(valueToExpand, audience, platform)
}
func (s *Service) ResolveVariables(audience VariableAudience, platform VariablePlatform) map[string]ResolvedVariable {
	return s.config.ResolveVariables(audience, platform)
}
func (s *Service) EffectiveStoragePath() string {
	return s.config.EffectiveStoragePath()
}
