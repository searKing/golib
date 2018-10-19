//https://tools.ietf.org/html/rfc7064
//https://tools.ietf.org/html/rfc7065
package ice

import (
	"fmt"
	"github.com/searKing/golib/net/urlutil"
	"net"
	"strconv"
)

// URL represents a STUN (rfc7064) or TURN (rfc7065) URL
type URL struct {
	Scheme Scheme
	Host   string
	Port   int
	Proto  Transport
}

// ParseURL parses a STUN or TURN urls following the ABNF syntax described in
// https://tools.ietf.org/html/rfc7064 and https://tools.ietf.org/html/rfc7065
// respectively.
func ParseURL(raw string) (*URL, error) {
	standardUrl, host, port, err := urlutil.ParseURL(raw, getDefaultPort)
	if err != nil {
		return nil, err
	}

	var u URL
	u.Scheme, err = ParseSchemeType(standardUrl.Scheme)
	if err != nil {
		return nil, err
	}

	if u.Host = host; u.Host == "" {
		return nil, fmt.Errorf("Missing host")
	}

	if u.Port = port; port == -1 {
		return nil, fmt.Errorf("Missing port")
	}

	proto, err := parseProto(u.Scheme, standardUrl.RawQuery)
	if err != nil {
		return nil, err
	}
	u.Proto = proto

	return &u, nil
}
func parseProto(scheme Scheme, rawQuery string) (Transport, error) {
	switch scheme {
	case SchemeSTUN, SchemeSTUNS:
		return parseStunProto(scheme, rawQuery)
	case SchemeTURN, SchemeTURNS:
		return parseTurnProto(scheme, rawQuery)
	default:
		return "", fmt.Errorf("malformed scheme:%s", scheme.String())
	}
}

func (u URL) String() string {
	rawURL := u.Scheme.String() + ":" + net.JoinHostPort(u.Host, strconv.Itoa(u.Port))
	if u.Scheme == SchemeTURN || u.Scheme == SchemeTURNS {
		rawURL += "?transport=" + u.Proto.String()
	}
	return rawURL
}

// IsSecure returns whether the this URL's scheme describes secure scheme or not.
func (u URL) IsSecure() bool {
	return u.Scheme == SchemeSTUNS || u.Scheme == SchemeTURNS
}
func (u URL) IsStunFamily() bool {
	return u.Scheme == SchemeSTUN || u.Scheme == SchemeSTUNS
}

func (u URL) IsTurnFamily() bool {
	return u.Scheme == SchemeTURN || u.Scheme == SchemeTURNS
}
