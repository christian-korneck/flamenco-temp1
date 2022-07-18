package own_url

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	globalIPv6    = net.ParseIP("2a10:3780:2:52:185:93:175:46")
	lanIPv6       = globalIPv6
	linkLocalIPv6 = net.ParseIP("fe80::5054:ff:fede:2ad7")
	localhostIPv6 = net.ParseIP("::1")

	globalIPv4    = net.ParseIP("8.8.8.8")
	lanIPv4       = net.ParseIP("192.168.0.1")
	linkLocalIPv4 = net.ParseIP("169.254.47.42")
	localhostIPv4 = net.ParseIP("127.0.0.1")
)

func Test_filterAddresses(t *testing.T) {
	tests := []struct {
		name   string
		expect []net.IP
		input  []net.IP
	}{
		// IPv6 tests:
		// Not a link-local address present, then use all but localhost
		{"IPv6 without link-local",
			[]net.IP{globalIPv6, lanIPv6},
			[]net.IP{globalIPv6, lanIPv6, localhostIPv6}},
		// In a mix, only the global address should be used.
		{"IPv6 with link-local",
			[]net.IP{globalIPv6},
			[]net.IP{linkLocalIPv6, globalIPv6, localhostIPv6}},
		// Only loopback and link-local.
		{"IPv6 with link-local + loopback",
			[]net.IP{localhostIPv6},
			[]net.IP{localhostIPv6, linkLocalIPv6}},
		// Only loopback
		{"IPv6 with only loopback",
			[]net.IP{localhostIPv6},
			[]net.IP{localhostIPv6}},

		// IPv4 tests:
		// Not a link-local address present, then use all but localhost
		{"IPv4 without link-local",
			[]net.IP{globalIPv4, lanIPv4},
			[]net.IP{globalIPv4, lanIPv4, localhostIPv4}},
		// In a mix, only the global and lan addresses should be used.
		{"IPv4 with link-local",
			[]net.IP{globalIPv4, lanIPv4},
			[]net.IP{globalIPv4, linkLocalIPv4, lanIPv4, localhostIPv4}},
		// Only loopback and link-local.
		{"IPv4 with link-local + loopback",
			[]net.IP{linkLocalIPv4},
			[]net.IP{localhostIPv4, linkLocalIPv4}},
		// Only loopback
		{"IPv4 with only loopback",
			[]net.IP{localhostIPv4},
			[]net.IP{localhostIPv4}},

		// Mixed IPv4/IPv6 tests:
		// IPv4 no link-local, but IPv6 with link-local:
		{"IPv4 w/o, IPv6 w/ link-local",
			[]net.IP{lanIPv4, lanIPv6},
			[]net.IP{lanIPv4, localhostIPv4, lanIPv6, linkLocalIPv6}},
		// IPv4 link-local, IPv6 without:
		{"IPv4 w/, IPv4 w/o link-local",
			[]net.IP{lanIPv4, lanIPv6},
			[]net.IP{linkLocalIPv4, lanIPv4, lanIPv6}},
		// Only loopback
		{"IPv4 + IPv6 with only loopback",
			[]net.IP{localhostIPv4, localhostIPv6},
			[]net.IP{localhostIPv4, localhostIPv6}},
	}
	for _, tt := range tests {
		got := filterAddresses(tt.input)
		assert.EqualValues(t, tt.expect, got, "for test %q", tt.name)
	}
}
