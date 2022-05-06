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
	"errors"
	"fmt"
	"net"

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
				logger.Trace().Msg("    - usable")
				ifaceAddresses = append(ifaceAddresses, ip)
			}
		}

		usableAddresses = append(usableAddresses, filterAddresses(ifaceAddresses)...)
	}

	if len(usableAddresses) == 0 {
		return usableAddresses, ErrNoInterface
	}

	return usableAddresses, nil
}

// filterAddresses reduces the number of IPv6 addresses.
// It prefers link-local addresses; if these are in the list, all the other IPv6
// addresses will be removed. Link-local addresses are stable and meant for
// same-network connections, which is exactly what Flamenco needs.
// Loopback addresses (localhost) are always filtered out, unless they're the only addresses available.
func filterAddresses(addrs []net.IP) []net.IP {
	keepAddrs := make([]net.IP, 0)

	if hasOnlyLoopback(addrs) {
		return addrs
	}

	var keepLinkLocalv6 = hasLinkLocalv6(addrs)
	var keepLinkLocalv4 = hasLinkLocalv4(addrs)

	var keep bool
	for _, addr := range addrs {
		if addr.IsLoopback() {
			continue
		}

		isv4 := isIPv4(addr)
		if isv4 {
			keep = keepLinkLocalv4 == addr.IsLinkLocalUnicast()
		} else {
			keep = keepLinkLocalv6 == addr.IsLinkLocalUnicast()
		}

		if keep {
			keepAddrs = append(keepAddrs, addr)
		}
	}

	return keepAddrs
}

func isIPv4(addr net.IP) bool {
	return addr.To4() != nil
}

func hasLinkLocalv6(addrs []net.IP) bool {
	for _, addr := range addrs {
		if !isIPv4(addr) && addr.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

func hasLinkLocalv4(addrs []net.IP) bool {
	for _, addr := range addrs {
		if isIPv4(addr) && addr.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

func hasOnlyLoopback(addrs []net.IP) bool {
	for _, addr := range addrs {
		if !addr.IsLoopback() {
			return false
		}
	}
	return true
}
