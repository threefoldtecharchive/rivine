package client

import (
	"errors"
	"regexp"
	"strings"
)

var (
	urlSchemeSplitter   = regexp.MustCompile(`^(https?://)?(.+)$`)
	urlLocalHostMatcher = regexp.MustCompile(`^(localhost|127\.0\.0\.1)?(\:[0-9]{1,5})?$`)
)

// look, a really bad validator! Hide it please :(
func sanitizeURL(url string) (string, error) {
	parts := urlSchemeSplitter.FindStringSubmatch(url)
	if len(parts) == 0 {
		return "", errors.New("invalid url format") // or perhaps our regexp just sucks >.<
	}
	if parts[1] == "" {
		if localParts := urlLocalHostMatcher.FindStringSubmatch(url); len(localParts) == 3 {
			parts[1] = "http://" // default to http for localhost
			if localParts[2] == "" {
				parts[2] += ":23110" // default to our default local daemon RPC port
			}
		} else {
			parts[1] = "https://" // default to https, you want insecure, though luck, be explicit!
		}
	} else if localParts := urlLocalHostMatcher.FindStringSubmatch(parts[2]); len(localParts) == 3 {
		if localParts[2] == "" {
			parts[2] += ":23110" // default to our default local daemon RPC port
		}
	}
	return strings.Join(parts[1:], ""), nil
}
