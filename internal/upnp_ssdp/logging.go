// package upnp_ssdp allows Workers to find their Manager on the LAN.
package upnp_ssdp

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
	"github.com/fromkeith/gossdp"
	"github.com/rs/zerolog"
)

type ssdpLogger struct {
	zlog *zerolog.Logger
}

var _ gossdp.LoggerInterface = (*ssdpLogger)(nil)

// wrappedLogger returns a gossdp.LoggerInterface-compatible wrapper around the given logger.
func wrappedLogger(logger *zerolog.Logger) *ssdpLogger {
	return &ssdpLogger{
		zlog: logger,
	}
}

func (sl *ssdpLogger) Tracef(fmt string, args ...interface{}) {
	sl.zlog.Debug().Msgf("SSDP: "+fmt, args...)
}

func (sl *ssdpLogger) Infof(fmt string, args ...interface{}) {
	sl.zlog.Info().Msgf("SSDP: "+fmt, args...)
}

func (sl *ssdpLogger) Warnf(fmt string, args ...interface{}) {
	sl.zlog.Warn().Msgf("SSDP: "+fmt, args...)
}

func (sl *ssdpLogger) Errorf(fmt string, args ...interface{}) {
	// Errors from the SSDP library are logged by that library AND returned as
	// error, which then triggers our own code to log the error as well. Since our
	// code can provide more context about what it's doing, demote SSDP errors to
	// the warning level.
	sl.zlog.Warn().Msgf("SSDP: "+fmt, args...)
}
