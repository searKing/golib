package ice

import (
	"fmt"
	"strings"
)

// Scheme indicates the type of server used in the ice.URL structure.
type Scheme string

const (
	// SchemeSTUN indicates the URL represents a STUN server.
	SchemeSTUN Scheme = "stun"

	// SchemeSTUNS indicates the URL represents a STUNS (secure) server.
	SchemeSTUNS Scheme = "stuns"

	// SchemeTURN indicates the URL represents a TURN server.
	SchemeTURN Scheme = "turn"

	// SchemeTURNS indicates the URL represents a TURNS (secure) server.
	SchemeTURNS Scheme = "turns"
)

func ParseSchemeType(s string) (Scheme, error) {
	scheme := Scheme(strings.ToLower(s))
	switch scheme {
	case SchemeSTUN, SchemeSTUNS, SchemeTURN, SchemeTURNS:
		return scheme, nil
	default:
		return "", fmt.Errorf("malformed scheme %s", s)
	}
}

func (t Scheme) String() string {
	switch t {
	case SchemeSTUN, SchemeSTUNS, SchemeTURN, SchemeTURNS:
		return string(t)
	default:
		return fmt.Errorf("malformed scheme %s", t).Error()
	}
}
