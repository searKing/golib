package internal

import (
	"github.com/searKing/golib/net/http_"
	"net/http"
)

type Credentials string

// ParseAuthenticationScheme parses userID and passwordf frome http's Header
// https://tools.ietf.org/html/rfc1945#section-11 11.1
func ParseAuthenticationCredentials(r *http.Request) (credentials string) {
	// Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==
	return r.Header.Get(http_.HeaderFieldAuthorization)
}
