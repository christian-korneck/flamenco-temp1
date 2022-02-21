package config

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

// Service provides access to Flamenco Manager configuration.
type Service struct {
	config Conf
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Load() error {
	config, err := getConf()
	if err != nil {
		return err
	}
	s.config = config
	return nil
}

func (s *Service) ExpandVariables(valueToExpand, audience, platform string) string {
	return s.config.ExpandVariables(valueToExpand, audience, platform)
}

func (s *Service) Get() *Conf {
	return &s.config
}
