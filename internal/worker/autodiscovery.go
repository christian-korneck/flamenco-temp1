package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"git.blender.org/flamenco/internal/upnp_ssdp"
	"git.blender.org/flamenco/pkg/api"
	"github.com/rs/zerolog/log"
)

// maybeAutodiscoverManager starts Manager auto-discovery if there is no Manager URL configured yet.
func MaybeAutodiscoverManager(ctx context.Context, configWrangler *FileConfigWrangler) error {
	cfg, err := configWrangler.WorkerConfig()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	if cfg.ManagerURL != "" {
		// Manager URL is already known, don't bother with auto-discovery.
		return nil
	}

	foundManager, err := autodiscoverManager(ctx)
	if err != nil {
		return err
	}

	configWrangler.SetManagerURL(foundManager)
	return nil
}

// autodiscoverManager uses UPnP/SSDP to find a Manager, and returns its URL if found.
func autodiscoverManager(ctx context.Context) (string, error) {
	c, err := upnp_ssdp.NewClient(log.Logger)
	if err != nil {
		return "", fmt.Errorf("unable to create UPnP/SSDP client: %w", err)
	}

	logger := log.Logger
	if deadline, ok := ctx.Deadline(); ok {
		timeout := deadline.Sub(time.Now()).Round(1 * time.Second)
		logger = logger.With().Str("timeout", timeout.String()).Logger()
	}
	logger.Info().Msg("auto-discovering Manager via UPnP/SSDP")

	urls, err := c.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to find Manager: %w", err)
	}

	if len(urls) == 0 {
		return "", errors.New("no Manager could be found")
	}

	// Try out the URLs to see which one responds.
	usableURLs := pingManagers(ctx, urls)

	switch len(usableURLs) {
	case 0:
		return "", fmt.Errorf("autodetected %d URLs, but none were usable", len(urls))
	case 1:
		log.Info().Str("url", usableURLs[0]).Msg("found Manager")
	default:
		log.Info().
			Strs("urls", urls).
			Str("url", usableURLs[0]).
			Msg("found multiple usable URLs, using the first one")
	}

	return usableURLs[0], nil
}

// pingManager connects to a Manager and returns true if it responds.
func pingManager(ctx context.Context, url string) bool {
	logger := log.With().Str("manager", url).Logger()

	client, err := api.NewClientWithResponses(url)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to create API client with this URL")
		return false
	}

	resp, err := client.GetVersionWithResponse(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to get Flamenco version from Manager")
		return false
	}

	if resp.JSON200 == nil {
		logger.Warn().
			Int("httpStatus", resp.StatusCode()).
			Msg("unable to get Flamenco version, unexpected reply")
		return false
	}

	logger.Info().
		Str("version", resp.JSON200.Version).
		Str("name", resp.JSON200.Name).
		Msg("found Flamenco Manager")
	return true
}

// pingManagers pings all URLs in parallel, returning only those that responded.
func pingManagers(ctx context.Context, urls []string) []string {
	startTime := time.Now()

	wg := new(sync.WaitGroup)
	wg.Add(len(urls))
	mutex := new(sync.Mutex)

	pingURL := func(idx int, url string) {
		defer wg.Done()
		ok := pingManager(ctx, url)

		mutex.Lock()
		defer mutex.Unlock()

		if !ok {
			// Erase the URL from the usable list.
			// Modifying the original slice instead of appending to a new one ensures
			// the original order is maintained.
			urls[idx] = ""
		}
	}

	for idx, url := range urls {
		go pingURL(idx, url)
	}

	wg.Wait()
	log.Debug().Str("pingTime", time.Since(startTime).String()).Msg("pinging all Manager URLs done")

	// Find the usable URLs.
	usableURLs := make([]string, 0)
	for _, url := range urls {
		if url != "" {
			usableURLs = append(usableURLs, url)
		}
	}

	return usableURLs
}
