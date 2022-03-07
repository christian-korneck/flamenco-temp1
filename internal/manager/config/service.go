package config

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

func (s *Service) ExpandVariables(valueToExpand string, audience VariableAudience, platform string) string {
	return s.config.ExpandVariables(valueToExpand, audience, platform)
}

func (s *Service) Get() *Conf {
	return &s.config
}
