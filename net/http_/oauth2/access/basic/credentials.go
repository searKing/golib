package basic

import (
	"encoding/base64"
	"errors"
	"github.com/searKing/golib/net/http_/oauth2/access"
	"net/http"
	"strings"
)

// rfc1945 11.1
// WWW-Authenticate: Basic realm="WallyWorld"
// credentials    = basic-credentials | ( auth-scheme #auth-param )
// basic-credentials = "Basic" SP basic-cookie
// basic-cookie      = <base64 [5] encoding of userid-password,
// except not limited to 76 char/line>
// userid-password   = [ token ] ":" *TEXT
type Credentials struct {
	UserID   string
	Password string
}

func ParseCredentialsFromRequest(r *http.Request) (*Credentials, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, errors.New("missing Authorization in Header")
	}
	return ParseCredentials(auth)

}

func ParseCredentials(auth string) (*Credentials, error) {
	username, password, ok := parseBasicAuth(auth)
	if !ok {
		return nil, &access.BadStringError{"malformed basic-credentials", auth}
	}
	return &Credentials{
		UserID:   username,
		Password: password,
	}, nil
}

func (a *Credentials) String() string {

	var b strings.Builder
	b.WriteString("Basic")
	b.WriteRune(' ')
	b.WriteString(basicAuth(a.UserID, a.Password))
	return b.String()
}

// SetAuthHeader sets the Authorization header to r using the access
// token in t.
//
// This method is unnecessary when using Transport or an HTTP Client
// returned by this package.
// Authorization: Basic jwth.jwtb.jwts
func (a *Credentials) SetAuthHeader(r *http.Request) {
	r.Header.Set("Authorization", a.String())
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

// See 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
