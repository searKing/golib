package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
)

// rfc6749 4.1.3
// POST /token HTTP/1.1
// Host: server.example.com
// Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
// Content-Type: application/x-www-form-urlencoded
//
// grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
// &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
func RetrieveAccessTokenRequest(ctx context.Context, r *http.Request) (*AccessTokenRequest, error) {
	var body []byte
	defer func() {
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
	}()

	// rfc6749 2.3.1
	// The client constructs the request URI by adding the following
	// parameters to the query component of the authorization endpoint URI
	// using the "application/x-www-form-urlencoded" format, per Appendix B:
	// Alternatively, the authorization server MAY support including the
	// client credentials in the request-body using the following
	// parameters
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	content, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch content {
	case "application/x-www-form-urlencoded", "text/plain":
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		return &AccessTokenRequest{
			GrantType: vals.Get("grant_type"),
			Username:  vals.Get("username"),
			Password:  vals.Get("password"),
			Scope:     vals.Get("scope"),
		}, nil
	case "application/json":
		var req AccessTokenRequest
		if err = json.Unmarshal(body, &req); err != nil {

			return nil, err
		}
		return &req, nil
	default:
		vars := r.URL.Query()
		grantTypes, ok := vars["grant_type"]
		if !ok || len(grantTypes) == 0 {
			return nil, errors.New("missing grant_type")
		}
		usernames, ok := vars["username"]
		if !ok || len(usernames) == 0 {
			return nil, errors.New("missing username")
		}
		passwords, ok := vars["password"]
		if !ok || len(passwords) == 0 {
			return nil, errors.New("missing password")
		}
		getValue := func(key string) string {
			vals, ok := vars[key]
			if !ok || len(vals) == 0 {
				return ""
			}
			return vals[0]
		}
		return &AccessTokenRequest{
			GrantType: grantTypes[0],
			Username:  usernames[0],
			Password:  passwords[0],
			Scope:     getValue("scope"),
		}, nil
	}
}
