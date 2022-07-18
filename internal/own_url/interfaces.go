package own_url

/* (c) 2019, Blender Foundation - Sybren A. Stüvel
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 *
 * The above copyright notice and this permission notice shall be
 * included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sort"

	"github.com/rs/zerolog/log"
)

var (
	// ErrNoInterface is returned when no network interfaces with a real IP-address were found.
	ErrNoInterface = errors.New("no network interface found")
)

// networkInterfaces returns a list of interface addresses.
// Only those addresses that can be reached by a unicast TCP/IP connection are returned.
func networkInterfaces() ([]net.IP, error) {
	log.Trace().Msg("iterating over all network interfaces")

	interfaces, err := net.Interfaces()
	if err != nil {
		return []net.IP{}, err
	}

	usableAddresses := make([]net.IP, 0)
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			log.Trace().Str("interface", iface.Name).Msg("skipping down interface")
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		ifaceAddresses := make([]net.IP, 0)
		for k := range addrs {
			var ip net.IP
			switch a := addrs[k].(type) {
			case *net.IPAddr:
				ip = a.IP
			case *net.IPNet:
				ip = a.IP
			default:
				log.Warn().
					Interface("addr", addrs[k]).
					Str("type", fmt.Sprintf("%T", addrs[k])).
					Msg("    - skipping unknown interface type")
				continue
			}

			logger := log.With().
				Interface("ip", ip).
				Str("iface", iface.Name).
				Logger()
			switch {
			case ip.IsMulticast():
				logger.Trace().Msg("    - skipping multicast")
			case ip.IsUnspecified():
				logger.Trace().Msg("    - skipping unspecified")
			default:
				logger.Trace().Msg("    - potentially usable")
				ifaceAddresses = append(ifaceAddresses, ip)
			}
		}

		usableAddresses = append(usableAddresses, filterAddresses(ifaceAddresses)...)
	}

	if len(usableAddresses) == 0 {
		return usableAddresses, ErrNoInterface
	}

	sort.Slice(usableAddresses, func(i, j int) bool {
		// Sort loopback addresses after others.
		if usableAddresses[i].IsLoopback() != usableAddresses[j].IsLoopback() {
			return usableAddresses[j].IsLoopback()
		}
		// Sort IPv4 before IPv6, because people are likely to be more familiar with
		// them.
		if isIPv4(usableAddresses[i]) != isIPv4(usableAddresses[j]) {
			return isIPv4(usableAddresses[i])
		}
		// Otherwise just order lexicographically.
		return bytes.Compare(usableAddresses[i], usableAddresses[j]) < 0
	})

	return usableAddresses, nil
}

// filterAddresses reduces the number of IP addresses.
//
// The function prefers non-link-local addresses over link-local ones.
// Link-local addresses are stable and meant for same-network connections, but
// they require a "zone index", typically the interface name, so something like
// `[fe80::cafe:f00d%eth0]`. This is not supported by webbrowsers. Furthermore,
// they require the interface name of the side initiating the connection,
// whereas this code is used to answer the question "how can this machine be
// reached?".
//
// Source: https://stackoverflow.com/a/52972417/875379
//
// Loopback addresses (localhost) are always filtered out, unless they're the
// only addresses available.
func filterAddresses(addrs []net.IP) []net.IP {
	keepAddrs := make([]net.IP, 0)
	keepLinkLocalv4 := hasOnlyLinkLocalv4(addrs)

	for _, addr := range addrs {
		if addr.IsLoopback() {
			continue
		}

		isv4 := isIPv4(addr)

		var keep bool
		if isv4 {
			keep = keepLinkLocalv4 == addr.IsLinkLocalUnicast()
		} else {
			// Never keep IPv6 link-local addresses. They need a "zone index" to work,
			// and those can only be determined on the connecting side. Furthermore,
			// they're incompatible with most webbrowsers.
			keep = !addr.IsLinkLocalUnicast()
		}

		if keep {
			keepAddrs = append(keepAddrs, addr)
		}
	}

	// Only when after the filtering there is nothing left, add the loopback
	// addresses. This is likely a bit of a strange test, because either this is a
	// loopback device (and should only have loopback addresses) or it is not (and
	// should only have non-loopback addresses). It does make the code reliable
	// even when things are mixed, which is nice.
	if len(keepAddrs) == 0 {
		for _, addr := range addrs {
			if addr.IsLoopback() {
				keepAddrs = append(keepAddrs, addr)
			}
		}
	}

	return keepAddrs
}

func isIPv4(addr net.IP) bool {
	return addr.To4() != nil
}

func hasOnlyLinkLocalv4(addrs []net.IP) bool {
	hasLinkLocalv4 := false
	for _, addr := range addrs {
		// Only consider non-loopback IPv4 addresses.
		if addr.IsLoopback() || !isIPv4(addr) {
			continue
		}
		if !addr.IsLinkLocalUnicast() {
			return false
		}
		hasLinkLocalv4 = true
	}
	return hasLinkLocalv4
}
