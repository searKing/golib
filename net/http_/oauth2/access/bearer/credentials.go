package bearer

import (
	"github.com/searKing/golib/net/http_/oauth2/access"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

// rfc6750 2.1
// Authorization: Bearer x.y.z
// credentials = "Bearer" 1*SP b64token
// b64token = 1*( ALPHA / DIGIT / "-" / "." / "_" / "~" / "+" / "/" ) *"="
type Credentials struct {
	B64token string
}

func parseCredentialsFromHeaderField(r *http.Request) string {
	// rfc6750 2.1
	// Authorization Request Header Field
	return r.Header.Get("Authorization")
}

func parseCredentialsFromFormEncodedBodyParameter(r *http.Request) string {
	defer r.Body.Close()
	// rfc6750 2.2
	// Authorization Request Header Field
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return ""
	}
	content, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch content {
	// rfc6750 2.2
	// Form-Encoded Body Parameter
	case "application/x-www-form-urlencoded":
		if r.Method == http.MethodGet {
			break
		}
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return ""
		}
		return vals.Get("access_token")
	}
	return ""
}

func parseCredentialsFromURIQueryParameter(r *http.Request) string {
	// rfc6750 2.3
	// URI Query Parameter
	vars := r.URL.Query()
	accessTokens, ok := vars["access_token"]
	if !ok || len(accessTokens) == 0 {
		return ""
	}
	return accessTokens[0]
}

func ParseCredentialsFromRequest(r *http.Request) (*Credentials, error) {
	auth := parseCredentialsFromHeaderField(r)
	if auth == "" {
		auth = parseCredentialsFromFormEncodedBodyParameter(r)
		if auth == "" {
			auth = parseCredentialsFromURIQueryParameter(r)
		}
	}

	return ParseAccessAuthentication(auth)
}
func ParseAccessAuthentication(auth string) (*Credentials, error) {
	// Authorization: Bearer x.y.z

	// basic-credentials = "Bearer" SP jwt-token
	s := strings.SplitN(auth, " ", 2)
	if len(s) != 2 || s[0] != "Bearer" {
		return nil, &access.BadStringError{"malformed credentials", auth}
	}
	b64token := strings.TrimLeftFunc(s[1], unicode.IsSpace)
	authentication := &Credentials{
		B64token: b64token,
	}
	if !authentication.Valid() {
		return nil, &access.BadStringError{"malformed b64token", b64token}
	}
	return authentication, nil
}

func isB64token(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) ||
		r == '-' || r == '.' || r == '_' || r == '~' || r == '+' || r == '/'
}

func (a *Credentials) Valid() bool {
	if a.B64token == "" {
		return false
	}
	b64token := a.B64token
	b64token = strings.TrimRight(b64token, "=")
	for _, token := range b64token {
		if isB64token(token) {
			continue
		}
		return false
	}
	return true
}

func (a *Credentials) String() string {
	var b strings.Builder
	b.WriteString("Bearer")
	b.WriteRune(' ')
	b.WriteString(a.B64token)
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
