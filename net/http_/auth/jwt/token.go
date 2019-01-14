package jwt

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/searKing/golib/net/http_/auth"
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

func NewAuthenticationScheme(alg string, keys ...[]byte) *AuthenticationScheme {
	return &AuthenticationScheme{
		Key: NewAuthKey(alg, keys...),
	}
}
func NewAuthenticationSchemeFromFile(alg string, keyFiles ...string) *AuthenticationScheme {
	return &AuthenticationScheme{
		Key: NewAuthKeyFromFile(alg, keyFiles...),
	}
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
		return a.Key.GetSignedKey(token)
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
	return a.ReadString(auth.ParseAuthenticationCredentials(r))
}

// if token is empty, generate; else reuse it
func (a *AuthenticationScheme) Write(w io.Writer) error {
	if a.bufferedToken == "" {
		_, err := a.RefreshToken()
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(w, `%s %s`, AuthSchemaJWT, a.bufferedToken)
	return nil
}

func (a *AuthenticationScheme) RefreshToken() (token string, err error) {
	jwtToken, err := a.signedString(a.Claims)
	if err == nil {
		return "", err
	}
	a.bufferedToken = jwtToken
	return jwtToken, nil
}

func (a *AuthenticationScheme) String() string {
	b := bytes.NewBuffer([]byte{})
	bw := bufio.NewWriter(b)
	if err := a.Write(bw); err != nil {
		return ""
	}
	return b.String()
}

// Authorization: Basic jwth.jwtb.jwts
func (a *AuthenticationScheme) WriteHTTP(w http.ResponseWriter) {
	w.Header().Set(auth.HeaderFieldAuthorization, a.String())
}

func (a *AuthenticationScheme) signedString(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(a.Key.GetSignedMethod(), claims)
	key, err := a.Key.GetSignedKey(nil)
	if err != nil {
		return "", err
	}
	return token.SignedString(key)
}
