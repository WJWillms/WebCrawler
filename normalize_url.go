package main

import (
	"net/url"
	"strings"
)

// normalizeURL normalizes a URL string.
func normalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Convert scheme and host to lowercase
	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	parsedURL.Host = strings.ToLower(parsedURL.Host)

	// Remove trailing slash from path
	parsedURL.Path = strings.TrimRight(parsedURL.Path, "/")

	// Remove default ports
	if parsedURL.Scheme == "http" && strings.HasSuffix(parsedURL.Host, ":80") {
		parsedURL.Host = strings.TrimSuffix(parsedURL.Host, ":80")
	} else if parsedURL.Scheme == "https" && strings.HasSuffix(parsedURL.Host, ":443") {
		parsedURL.Host = strings.TrimSuffix(parsedURL.Host, ":443")
	}

	// Remove fragment
	parsedURL.Fragment = ""

	normalizedURL := parsedURL.Host + parsedURL.Path + parsedURL.RawQuery

	// Construct and return the normalized URL with the scheme
	return normalizedURL, nil
}
