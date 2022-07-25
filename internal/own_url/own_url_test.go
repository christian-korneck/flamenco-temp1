// Package own_url provides a way for a process to find a URL on which it can be reached.
package own_url

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"net"
	"testing"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestAvailableURLs(t *testing.T) {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	// This should run without errors. It's hard to predict the returned URLs
	// though, as they depend on the local network devices.
	urls, err := AvailableURLs("http", ":9999")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, urls, "expected at least one URL to be returned")
}

func TestSpecificHostURL(t *testing.T) {
	tests := []struct {
		name   string
		expect string // Empty string encodes "expect nil pointer"
		listen string
	}{
		{"Specific IPv4 with port", "http://192.168.0.1:8080/", "192.168.0.1:8080"},
		{"Specific IPv4 without port", "http://192.168.0.1/", "192.168.0.1"},
		{"Specific IPv6 with port", "http://[fe80::5054:ff:fede:2ad7]:8080/", "[fe80::5054:ff:fede:2ad7]:8080"},
		{"Specific IPv6 without port", "http://[fe80::5054:ff:fede:2ad7]/", "[fe80::5054:ff:fede:2ad7]"},

		{"Wildcard IPv4", "", "0.0.0.0:8080"},
		{"Wildcard IPv6", "", "[::0]:8080"},
		{"No host, just port", "", ":8080"},

		{"Invalid address", "http://this%20is%20not%20an%20address/", "this is not an address"},
		{"Invalid port", "", "192.168.0.1::too-many-colons"},
	}

	for _, test := range tests {
		actual := specificHostURL("http", test.listen)
		if test.expect == "" {
			assert.Nil(t, actual, "for input %q", test.listen)
			continue
		}
		if actual == nil {
			t.Errorf("returned URL is nil for input %q", test.listen)
			continue
		}
		assert.Equal(t, test.expect, actual.String(), "for input %q", test.listen)
	}
}

func TestURLsForNetworkInterfaces(t *testing.T) {
	addrs := []net.IP{linkLocalIPv6, lanIPv4}
	urls, err := urlsForNetworkInterfaces("http", ":9999", addrs)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, urls, 2)
	assert.Equal(t, "http://[fe80::5054:ff:fede:2ad7]:9999/", urls[0].String())
	assert.Equal(t, "http://192.168.0.1:9999/", urls[1].String())
}
