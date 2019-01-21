package jwt_

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/searKing/golib/net/http_"
	"github.com/searKing/golib/net/http_/auth/internal"
	"io"
	"net/http"
	"strings"
)

// for Auth check
// Authorization: Bearer x.y.z
type AuthenticationScheme struct {
	Key    *AuthKey   `options:"required"`
	Claims jwt.Claims `options:"optional"`

	bufferedToken string
}

func NewAuthenticationSchemeFromRandom(alg string) (*AuthenticationScheme, error) {
	authKey, err := NewAuthKeyFromRandowm(alg)
	if err != nil {
		return nil, err
	}
	return &AuthenticationScheme{
		Key: authKey,
	}, err
}

func NewAuthenticationScheme(alg string, privateKey []byte, publicKey []byte, password ...string) (*AuthenticationScheme, error) {
	authKey, err := NewAuthKey(alg, privateKey, publicKey, password...)
	if err != nil {
		return nil, err
	}
	return &AuthenticationScheme{
		Key: authKey,
	}, err
}

func NewAuthenticationSchemeFromFile(alg string, privateKeyFile string, publicKeyFile string, password ...string) (*AuthenticationScheme, error) {
	authKey, err := NewAuthKeyFromFile(alg, privateKeyFile, publicKeyFile, password...)
	if err != nil {
		return nil, err
	}
	return &AuthenticationScheme{
		Key: authKey,
	}, err
}

// basic-credentials = "Basic" SP basic-cookie
func (a *AuthenticationScheme) ReadString(basicCredentials string) error {
	// Authorization: Bearer x.y.z

	// basic-credentials = "Bearer" SP jwt-token
	s := strings.SplitN(basicCredentials, " ", 2)
	if len(s) != 2 || s[0] != AuthSchemaJWT {
		return errors.New("auth header is invalid")
	}
	jwtToken := s[1]
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return a.Key.GetVerifiedKey(token)
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token is invalid")
	}
	a.Claims = token.Claims

	return nil
}

// ParseBasicAuthenticationScheme parses userID and passwordf frome http's Header
// Reference : https://tools.ietf.org/html/rfc1945#section-11 11.1
func (a *AuthenticationScheme) ReadHTTP(r *http.Request) error {
	// Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==
	return a.ReadString(internal.ParseAuthenticationCredentials(r))
}

// if token is empty, generate; else reuse it
func (a *AuthenticationScheme) Write(w io.Writer) error {
	if a.bufferedToken == "" {
		_, err := a.RefreshToken()
		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(w, `%s %s`, AuthSchemaJWT, a.bufferedToken)
	return err
}

func (a *AuthenticationScheme) RefreshToken() (token string, err error) {
	jwtToken, err := a.signedString(a.Claims)
	if err != nil {
		return "", err
	}
	a.bufferedToken = jwtToken
	return jwtToken, nil
}

func (a *AuthenticationScheme) String() string {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	if err := a.Write(bw); err != nil {
		return ""
	}
	bw.Flush()
	return buf.String()
}

// Authorization: Basic jwth.jwtb.jwts
func (a *AuthenticationScheme) WriteHTTP(w http.ResponseWriter) {
	w.Header().Set(http_.HeaderFieldAuthorization, a.String())
}

// SetAuthHeader sets the Authorization header to r using the access
// token in t.
//
// This method is unnecessary when using Transport or an HTTP Client
// returned by this package.
// Authorization: Basic jwth.jwtb.jwts
func (a *AuthenticationScheme) SetAuthHeader(r *http.Request) {
	r.Header.Set("Authorization", a.String())
}

func (a *AuthenticationScheme) signedString(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(a.Key.GetSignedMethod(), claims)
	key, err := a.Key.GetSignedKey(nil)
	if err != nil {
		return "", err
	}
	return token.SignedString(key)
}
