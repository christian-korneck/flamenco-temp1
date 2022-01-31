package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker"
)

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

const (
	credentialsFilename = "flamenco-worker-credentials.yaml"
	configFilename      = "flamenco-worker.yaml"
)

func loadConfig(configWrangler worker.FileConfigWrangler) (worker.WorkerConfig, error) {
	logger := log.With().Str("filename", configFilename).Logger()

	var cfg worker.WorkerConfig

	err := configWrangler.LoadConfig(configFilename, &cfg)

	// If the configuration file doesn't exist, write the defaults & retry loading them.
	if os.IsNotExist(err) {
		logger.Info().Msg("writing default configuration file")
		cfg = configWrangler.DefaultConfig()
		err = configWrangler.WriteConfig(configFilename, "Configuration", cfg)
		if err != nil {
			return cfg, fmt.Errorf("error writing default config: %w", err)
		}
		err = configWrangler.LoadConfig(configFilename, &cfg)
	}
	if err != nil {
		return cfg, fmt.Errorf("error loading config from %s: %w", configFilename, err)
	}

	// Validate the manager URL.
	if cfg.Manager != "" {
		_, err := worker.ParseURL(cfg.Manager)
		if err != nil {
			return cfg, fmt.Errorf("error parsing manager URL %s: %w", cfg.Manager, err)
		}
		logger.Debug().Str("url", cfg.Manager).Msg("parsed manager URL")
	}

	return cfg, nil
}

func loadCredentials(configWrangler worker.FileConfigWrangler) (worker.WorkerCredentials, error) {
	logger := log.With().Str("filename", configFilename).Logger()
	logger.Info().Msg("loading credentials")

	var creds worker.WorkerCredentials

	err := configWrangler.LoadConfig(credentialsFilename, &creds)
	if err != nil {
		return worker.WorkerCredentials{}, err
	}

	return creds, nil
}
