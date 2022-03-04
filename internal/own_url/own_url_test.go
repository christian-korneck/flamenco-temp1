// Package own_url provides a way for a process to find a URL on which it can be reached.
package own_url

import (
	"context"
	"testing"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestAvailableURLs(t *testing.T) {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()

	_, err := AvailableURLs(ctx, "http", ":9999", true)
	if err != nil {
		t.Fatal(err)
	}
	// t.Fatalf("urls: %v", urls)
}
