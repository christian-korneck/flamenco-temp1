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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fromkeith/gossdp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Client struct {
	ssdp       *gossdp.ClientSsdp
	log        *zerolog.Logger
	wrappedLog *ssdpLogger

	mutex *sync.Mutex
	urls  []string
}

func NewClient(logger zerolog.Logger) (*Client, error) {
	wrap := wrappedLogger(&logger)
	client := Client{
		log:        &logger,
		wrappedLog: wrap,

		mutex: new(sync.Mutex),
		urls:  make([]string, 0),
	}

	ssdp, err := gossdp.NewSsdpClientWithLogger(&client, wrap)
	if err != nil {
		return nil, fmt.Errorf("create UPnP/SSDP client: %w", err)
	}

	client.ssdp = ssdp
	return &client, nil
}

func (c *Client) Run(ctx context.Context) ([]string, error) {
	defer c.stopCleanly()

	log.Debug().Msg("waiting for UPnP/SSDP answer")
	go c.ssdp.Start()

	var waitTime time.Duration
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(waitTime):
			if err := c.ssdp.ListenFor(FlamencoServiceType); err != nil {
				return nil, fmt.Errorf("unable to find Manager: %w", err)
			}
			waitTime = 1 * time.Second

			urls := c.receivedURLs()
			if len(urls) > 0 {
				return urls, nil
			}
		}
	}
}

// Response is called by the gossdp library on M-SEARCH responses.
func (c *Client) Response(message gossdp.ResponseMessage) {
	logger := c.log.With().
		Int("maxAge", message.MaxAge).
		Str("searchType", message.SearchType).
		Str("deviceID", message.DeviceId).
		Str("usn", message.Usn).
		Str("location", message.Location).
		Str("server", message.Server).
		Str("urn", message.Urn).
		Logger()
	if message.DeviceId != FlamencoUUID {
		logger.Debug().Msg("ignoring message from unknown device")
		return
	}

	logger.Debug().Msg("UPnP/SSDP message received")
	c.appendURLs(message.Location)
}

func (c *Client) appendURLs(location string) {
	urls := strings.Split(location, LocationSeparator)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.urls = append(c.urls, urls...)
	c.log.Debug().
		Int("new", len(urls)).
		Int("total", len(c.urls)).
		Msg("new URLs received")
}

// receivedURLs takes a thread-safe copy of the URLs received so far.
func (c *Client) receivedURLs() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	urls := make([]string, len(c.urls))
	copy(urls, c.urls)
	return urls
}

// stopCleanly tries to stop the SSDP client cleanly, without spurious logging.
func (c *Client) stopCleanly() {

	c.log.Trace().Msg("UPnP/SSDP client stopping")

	// Sneakily disable warnings when shutting down, otherwise the read operation
	// from the UDP socket will cause a warning.
	tempLog := c.log.Level(zerolog.ErrorLevel)
	c.wrappedLog.zlog = &tempLog
	c.ssdp.Stop()
	c.wrappedLog.zlog = c.log

	c.log.Debug().Msg("UPnP/SSDP client stopped")
}
