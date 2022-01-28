package ssdp

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
	"fmt"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender-institute/gossdp"
)

var _ gossdp.LoggerInterface = ZeroLogWrapper{}

type ZeroLogWrapper struct{}

func (l ZeroLogWrapper) Debugf(msg string, args ...interface{}) {
	log.Debug().Msg(fmt.Sprintf(msg, args...))
}
func (l ZeroLogWrapper) Infof(msg string, args ...interface{}) {
	log.Info().Msg(fmt.Sprintf(msg, args...))
}
func (l ZeroLogWrapper) Warnf(msg string, args ...interface{}) {
	log.Warn().Msg(fmt.Sprintf(msg, args...))
}
func (l ZeroLogWrapper) Errorf(msg string, args ...interface{}) {
	log.Error().Msg(fmt.Sprintf(msg, args...))
}
