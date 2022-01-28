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
	"net/url"

	"github.com/rs/zerolog/log"
	"gitlab.com/blender-institute/gossdp"
)

// Finder is a uses UPnP/SSDP to find a Flamenco Manager on the local network.
type Finder struct {
	overrideURL *url.URL
}

type ssdpClient struct {
	response chan interface{}
}

// NewManagerFinder returns a default SSDP/UPnP based finder.
func NewManagerFinder(managerURL *url.URL) Finder {
	return Finder{
		overrideURL: managerURL,
	}
}

func (b *ssdpClient) NotifyAlive(message gossdp.AliveMessage) {
	log.Info().Interface("message", message).Msg("UPnP/SSDP NotifyAlive")
}
func (b *ssdpClient) NotifyBye(message gossdp.ByeMessage) {
	log.Info().Interface("message", message).Msg("UPnP/SSDP NotifyBye")
}
func (b *ssdpClient) Response(message gossdp.ResponseMessage) {
	log.Debug().Interface("message", message).Msg("UPnP/SSDP response")
	url, err := url.Parse(message.Location)
	if err != nil {
		b.response <- err
		return
	}
	b.response <- url
}

// FindFlamencoManager tries to find a Manager, sending its URL to the returned channel.
func (f Finder) FindFlamencoManager() <-chan *url.URL {
	reporter := make(chan *url.URL)

	go func() {
		defer close(reporter)

		if f.overrideURL != nil {
			log.Debug().Str("url", f.overrideURL.String()).Msg("Using configured Flamenco Manager URL")
			reporter <- f.overrideURL
			return
		}

		log.Info().Msg("finding Flamenco Manager via UPnP/SSDP")
		b := ssdpClient{make(chan interface{})}

		client, err := gossdp.NewSsdpClientWithLogger(&b, ZeroLogWrapper{})
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to create UPnP/SSDP client")
			return
		}

		log.Debug().Msg("Starting UPnP/SSDP client")
		go client.Start()
		defer client.Stop()

		if err := client.ListenFor("urn:flamenco:manager:0"); err != nil {
			log.Error().Err(err).Msg("unable to find Manager")
			return
		}

		log.Debug().Msg("Waiting for UPnP/SSDP answer")
		urlOrErr := <-b.response
		switch v := urlOrErr.(type) {
		case *url.URL:
			reporter <- v
		case error:
			log.Fatal().Err(v).Msg("Error waiting for UPnP/SSDP response from Manager")
		}
	}()

	return reporter
}
