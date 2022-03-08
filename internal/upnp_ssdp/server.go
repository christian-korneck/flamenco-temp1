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
	"context"
	"net/url"
	"strings"

	"github.com/fromkeith/gossdp"
	"github.com/rs/zerolog"
)

// Server advertises services via UPnP/SSDP.
type Server struct {
	ssdp       *gossdp.Ssdp
	log        *zerolog.Logger
	wrappedLog *ssdpLogger
}

func NewServer(logger zerolog.Logger) (*Server, error) {
	wrap := wrappedLogger(&logger)
	ssdp, err := gossdp.NewSsdpWithLogger(nil, wrap)
	if err != nil {
		return nil, err
	}
	return &Server{ssdp, &logger, wrap}, nil
}

// AddAdvertisement adds a service advertisement for Flamenco Manager.
// Must be called before calling Run().
func (s *Server) AddAdvertisement(serviceLocation string) {
	// Define the service we want to advertise
	serverDef := gossdp.AdvertisableServer{
		ServiceType: FlamencoServiceType,
		DeviceUuid:  FlamencoUUID,
		Location:    serviceLocation,
		MaxAge:      3600, // Number of seconds this advertisement is valid for.
	}
	s.ssdp.AdvertiseServer(serverDef)
	s.log.Info().Str("location", serviceLocation).Msg("UPnP/SSDP location registered")
}

// AddAdvertisementURLs constructs a service location from the given URLs, and
// adds the advertisement for it.
func (s *Server) AddAdvertisementURLs(urls []url.URL) {
	urlStrings := make([]string, len(urls))
	for idx := range urls {
		urlStrings[idx] = urls[idx].String()
	}
	location := strings.Join(urlStrings, LocationSeparator)
	s.AddAdvertisement(location)
}

// Run starts the advertisement, and blocks until the context is closed.
func (s *Server) Run(ctx context.Context) {
	s.log.Info().Msg("UPnP/SSDP advertisement starting")

	isStopping := false

	go func() {
		// There is a bug in the SSDP library, where closing the server can cause a panic.
		defer func() {
			if isStopping {
				// Only capture a panic when we expect one.
				value := recover()
				s.log.Debug().Interface("value", value).Msg("recovered from panic in SSDP library")
			}
		}()

		s.ssdp.Start()
	}()

	<-ctx.Done()

	s.log.Debug().Msg("UPnP/SSDP advertisement stopping")

	// Sneakily disable warnings when shutting down, otherwise the read operation
	// from the UDP socket will cause a warning.
	tempLog := s.log.Level(zerolog.ErrorLevel)
	s.wrappedLog.zlog = &tempLog
	isStopping = true
	s.ssdp.Stop()
	s.wrappedLog.zlog = s.log

	s.log.Info().Msg("UPnP/SSDP advertisement stopped")
}
