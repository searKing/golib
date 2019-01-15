package basic

import (
	"net/http"
)

// Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==

// HTTP/1.0  401 UnauthorizedFunc
// Content-type: text/html
// WWW-Authenticate: Basic realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
const (
	AuthSchemaBasic = "Basic"
	// https://tools.ietf.org/html/rfc7235#section-2.2
	defaultRealm = "Authorization Required"
)

// Basic is the http basic auth
func NewBasicAuthenticatorHandler(userID string, password string) http.Handler {
	authenticator := func(user, pass string) bool {
		return user == userID && pass == password
	}
	return NewBasicAuthenticatorAdvanceHandler(authenticator, defaultRealm)
}

// NewBasicAuthenticator return the BasicAuth
func NewBasicAuthenticatorAdvanceHandler(authenticator Authenticator, realm string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := &BasicAuth{Authenticator: authenticator, Realm: realm}
		if a.CheckAuth(r) {
			return
		}
		a.RequireAuth(w, r)
	})
}

// Callback function that should perform the authentication of the user based on userid and
// password. Must return true on success, false on failure. Required.
type Authenticator func(userid string, password string) (pass bool)

// BasicAuth store the SecretProvider and Realm
type BasicAuth struct {
	// Verify
	Authenticator Authenticator

	// Protection Space to display to the user. Required.
	// https://tools.ietf.org/html/rfc7235#section-2.2
	Realm string
}

// CheckAuth Checks the user-ID/password combination from the request. Returns
// either an empty string (authentication failed) or the name of the
// authenticated user.
// Supports MD5 and SHA1 password entries

func (a *BasicAuth) CheckAuth(r *http.Request) (pass bool) {
	if a.Authenticator == nil {
		return false
	}
	auth := NewAuthenticationScheme()
	auth.ReadHTTP(r)
	if a.Authenticator(auth.UserID, auth.Password) {
		return true
	}
	return false
}

// RequireAuth http.Handler for BasicAuth which initiates the authentication process
// (or requires reauthentication).
func (a *BasicAuth) RequireAuth(w http.ResponseWriter, r *http.Request) {
	auth := NewBasicAuthenticate(a.Realm)
	auth.WriteHTTP(w)
}
