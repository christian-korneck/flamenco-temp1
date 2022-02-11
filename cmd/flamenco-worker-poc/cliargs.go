package main

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

import (
	"flag"
	"net/url"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker"
)

var cliArgs struct {
	version    bool
	verbose    bool
	debug      bool
	managerURL *url.URL
	manager    string
	register   bool
}

func parseCliArgs() {
	flag.BoolVar(&cliArgs.version, "version", false, "Shows the application version, then exits.")
	flag.BoolVar(&cliArgs.verbose, "verbose", false, "Enable info-level logging.")
	flag.BoolVar(&cliArgs.debug, "debug", false, "Enable debug-level logging.")

	flag.StringVar(&cliArgs.manager, "manager", "", "URL of the Flamenco Manager.")
	flag.BoolVar(&cliArgs.register, "register", false, "(Re-)register at the Manager.")

	flag.Parse()

	if cliArgs.manager != "" {
		var err error
		cliArgs.managerURL, err = worker.ParseURL(cliArgs.manager)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid manager URL")
		}
	}
}
