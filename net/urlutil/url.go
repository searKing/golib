package urlutil

import (
	"net/url"
)

// parseURL is just urlutil.Parse. It exists only so that urlutil.Parse can be called
// in places where urlutil is shadowed for godoc. See https://golang.org/cl/49930.
var parseURL = url.Parse

func ParseURL(url string) (*url.URL, error) {

	u, err := parseURL(url) // Just urlutil.Parse (urlutil is shadowed for godoc).
	if err != nil {
		return nil, err
	}
	// The host's colon:port should be normalized. See Issue 14836.
	u.Host = removeEmptyPort(u.Host)

	return u, nil
}
