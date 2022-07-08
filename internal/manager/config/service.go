package config

import "github.com/rs/zerolog/log"

// SPDX-License-Identifier: GPL-3.0-or-later

// Service provides access to Flamenco Manager configuration.
type Service struct {
	config Conf
}

func NewService() *Service {
	return &Service{
		config: DefaultConfig(),
	}
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
