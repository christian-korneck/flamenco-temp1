package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"git.blender.org/flamenco/internal/own_url"
	"git.blender.org/flamenco/internal/upnp_ssdp"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	c, err := upnp_ssdp.NewServer(log.Logger)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Handle Ctrl+C
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		for signum := range signals {
			log.Info().Str("signal", signum.String()).Msg("signal received, shutting down")
			cancel()
		}
	}()

	urls, err := own_url.AvailableURLs("http", ":8080")
	if err != nil {
		log.Fatal().Err(err).Msg("unable to construct list of URLs")
	}
	urlStrings := []string{}
	for _, url := range urls {
		urlStrings = append(urlStrings, url.String())
	}
	log.Info().Strs("urls", urlStrings).Msg("URLs to try")

	location := strings.Join(urlStrings, ";")
	c.AddAdvertisement(location)

	c.Run(ctx)
}
