// Package own_url provides a way for a process to find a URL on which it can be reached.
package own_url

import (
	"fmt"
	"net"
	"net/url"

	"github.com/rs/zerolog/log"
)

func AvailableURLs(schema, listen string) ([]url.URL, error) {
	if listen == "" {
		panic("empty 'listen' parameter")
	}

	hostURL := specificHostURL(schema, listen)
	if hostURL != nil {
		return []url.URL{*hostURL}, nil
	}
	log.Debug().Str("listen", listen).Msg("not listening on any specific host")

	addrs, err := networkInterfaces()
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("iterating network interfaces to find possible URLs for Flamenco Manager.")
	return urlsForNetworkInterfaces(schema, listen, addrs)
}

// ToStringers converts an array of URLs to an array of `fmt.Stringer`.
func ToStringers(urls []url.URL) []fmt.Stringer {
	stringers := make([]fmt.Stringer, len(urls))
	for idx := range urls {
		stringers[idx] = &urls[idx]
	}
	return stringers
}

// specificHostURL returns the hosts's URL if the "listen" string is specific enough, otherwise nil.
// Examples: "192.168.0.1:8080" is specific enough, "0.0.0.0:8080" and ":8080" are not.
func specificHostURL(scheme, listen string) *url.URL {
	var (
		host string
		err  error
	)

	// Figure out which port we're supposted to listen on.
	if host, _, err = net.SplitHostPort(listen); err != nil {
		// This is annoying. SplitHostPort() doesn't return specific errors, so we
		// have to test on the error message to see what's the problem.
		// A missing port is fine, but other errors are not.
		addrErr := err.(*net.AddrError)
		if addrErr.Err != "missing port in address" {
			log.Warn().Str("address", listen).Err(err).Msg("unable to split host and port in address")
			return nil
		}

		// 'listen' doesn't have a port number, so it's just the host.
		host = listen
	}
	if host == "" {
		// An empty host is never specific enough.
		return nil
	}

	ip := net.ParseIP(host)
	if ip != nil && ip.IsUnspecified() {
		// The host is "::0" or "0.0.0.0"; not specific.
		return nil
	}

	// We can just construct a URL here, since we know it's a specific host anyway.
	return &url.URL{
		Scheme: scheme,
		Host:   listen,
		Path:   "/",
	}
}

func urlsForNetworkInterfaces(scheme, listen string, addrs []net.IP) ([]url.URL, error) {
	// Find the port number in the 'listen' string.
	var (
		port string
		err  error
	)
	// Get the port number as integer.
	if _, port, err = net.SplitHostPort(listen); err != nil {
		return nil, fmt.Errorf("unable to split host and port in address %q", listen)
	}

	links := make([]url.URL, 0)
	for _, addr := range addrs {
		var strAddr string
		if ipv4 := addr.To4(); ipv4 != nil {
			strAddr = ipv4.String()
		} else {
			strAddr = fmt.Sprintf("[%s]", addr)
		}

		urlForAddr := url.URL{
			Scheme: scheme,
			Host:   fmt.Sprintf("%s:%s", strAddr, port),
			Path:   "/",
		}
		links = append(links, urlForAddr)
	}

	return links, nil
}
