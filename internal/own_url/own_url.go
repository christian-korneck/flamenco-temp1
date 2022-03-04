// Package own_url provides a way for a process to find a URL on which it can be reached.
package own_url

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/rs/zerolog/log"
)

func AvailableURLs(ctx context.Context, schema, listen string, includeLocal bool) ([]*url.URL, error) {
	var (
		host, port string
		portnum    int
		err        error
	)

	if listen == "" {
		panic("empty 'listen' parameter")
	}

	// Figure out which port we're supposted to listen on.
	if host, port, err = net.SplitHostPort(listen); err != nil {
		return nil, fmt.Errorf("unable to split host and port in address '%s': %w", listen, err)
	}
	if portnum, err = net.DefaultResolver.LookupPort(ctx, "listen", port); err != nil {
		return nil, fmt.Errorf("unable to look up port '%s': %w", port, err)
	}

	// If the host is empty or ::0/0.0.0.0, show a list of URLs to connect to.
	listenSpecificHost := false
	var ip net.IP
	if host != "" {
		ip = net.ParseIP(host)
		if ip == nil {
			addrs, erresolve := net.DefaultResolver.LookupHost(ctx, host)
			if erresolve != nil {
				return nil, fmt.Errorf("unable to resolve listen host '%v': %w", host, erresolve)
			}
			if len(addrs) > 0 {
				ip = net.ParseIP(addrs[0])
			}
		}
		if ip != nil && !ip.IsUnspecified() {
			listenSpecificHost = true
		}
	}

	if listenSpecificHost {
		// We can just construct a URL here, since we know it's a specific host anyway.
		log.Debug().Str("host", ip.String()).Msg("listening on host")

		link := fmt.Sprintf("%s://%s:%d/", schema, host, portnum)
		myURL, errparse := url.Parse(link)
		if errparse != nil {
			return nil, fmt.Errorf("unable to parse listen URL %s: %w", link, errparse)
		}
		return []*url.URL{myURL}, nil
	}

	log.Debug().Str("host", host).Msg("not listening on any specific host")

	addrs, err := networkInterfaces(false, includeLocal)
	if err == ErrNoInterface {
		addrs, err = networkInterfaces(true, includeLocal)
	}
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("iterating network interfaces to find possible URLs for Flamenco Manager.")

	links := make([]*url.URL, 0)
	for _, addr := range addrs {
		var strAddr string
		if ipv4 := addr.To4(); ipv4 != nil {
			strAddr = ipv4.String()
		} else {
			strAddr = fmt.Sprintf("[%s]", addr)
		}

		constructedURL := fmt.Sprintf("%s://%s:%d/", schema, strAddr, portnum)
		parsedURL, err := url.Parse(constructedURL)
		if err != nil {
			log.Warn().
				Str("address", strAddr).
				Str("url", constructedURL).
				Err(err).
				Msg("skipping address, as it results in an unparseable URL")
			continue
		}
		links = append(links, parsedURL)
	}

	return links, nil
}
