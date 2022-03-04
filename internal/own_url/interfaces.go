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
// Only those addresses that can be eached by a unicast TCP/IP connection are returned.
func networkInterfaces(includeLinkLocal, includeLocalhost bool) ([]net.IP, error) {
	log.Debug().Msg("iterating over all network interfaces")

	interfaces, err := net.Interfaces()
	if err != nil {
		return []net.IP{}, err
	}

	usableAddresses := make([]net.IP, 0)
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			log.Debug().Str("interface", iface.Name).Msg("skipping down interface")
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
				logger.Debug().Msg("    - skipping multicast")
			case ip.IsMulticast():
				logger.Debug().Msg("    - skipping multicast")
			case ip.IsUnspecified():
				logger.Debug().Msg("    - skipping unspecified")
			case !includeLinkLocal && ip.IsLinkLocalUnicast():
				logger.Debug().Msg("    - skipping link-local")
			case !includeLocalhost && ip.IsLoopback():
				logger.Debug().Msg("    - skipping localhost")
			default:
				logger.Debug().Msg("    - usable")
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

// filterAddresses removes "privacy extension" addresses.
// It assumes the list of addresses belong to the same network interface, and
// that the OS reports preferred (i.e. private/random) addresses before
// non-random ones.
func filterAddresses(addrs []net.IP) []net.IP {
	keep := make([]net.IP, 0)

	var lastSeenIP net.IP
	for _, addr := range addrs {
		if addr.To4() != nil {
			// IPv4 addresses are always kept.
			keep = append(keep, addr)
			continue
		}

		lastSeenIP = addr
	}
	if len(lastSeenIP) > 0 {
		keep = append(keep, lastSeenIP)
	}

	return keep
}
