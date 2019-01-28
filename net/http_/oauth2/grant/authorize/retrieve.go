package authorize

import (
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

// rfc6749 4.1.1
// GET /authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz
//        &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
//    Host: server.example.com
func RetrieveAuthorizationRequest(ctx context.Context, r *http.Request) (*AuthorizationRequest, error) {
	defer r.Body.Close()

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
		return &AuthorizationRequest{
			ResponseType: vals.Get("response_type"),
			ClientId:     vals.Get("client_id"),
			RedirectUri:  vals.Get("redirect_uri"),
			Scope:        vals.Get("scope"),
			State:        vals.Get("state"),
		}, nil
	case "application/json":
		var req AuthorizationRequest
		if err = json.Unmarshal(body, &req); err != nil {

			return nil, err
		}
		return &req, nil
	default:
		vars := r.URL.Query()
		responseTypes, ok := vars["response_type"]
		if !ok || len(responseTypes) == 0 {
			return nil, errors.New("missing response_type")
		}
		clientIds, ok := vars["client_id"]
		if !ok || len(clientIds) == 0 {
			return nil, errors.New("missing client_id")

		}
		getVal := func(key string) string {
			vals, ok := vars[key]
			if !ok || len(vals) == 0 {
				return ""
			}
			return vals[0]
		}
		return &AuthorizationRequest{
			ResponseType: responseTypes[0],
			ClientId:     clientIds[0],
			RedirectUri:  getVal("redirect_uri"),
			Scope:        getVal("scope"),
			State:        getVal("state"),
		}, nil

	}
}

// rfc6749 4.1.3
// POST /token HTTP/1.1
// Host: server.example.com
// Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
// Content-Type: application/x-www-form-urlencoded
//
// grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
// &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
func RetrieveAccessTokenRequest(ctx context.Context, r *http.Request) (*AccessTokenRequest, error) {
	defer r.Body.Close()

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
			GrantType:   vals.Get("grant_type"),
			Code:        vals.Get("code"),
			RedirectUri: vals.Get("redirect_uri"),
			ClientId:    vals.Get("client_id"),
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
		codes, ok := vars["code"]
		if !ok || len(codes) == 0 {
			return nil, errors.New("missing code")
		}
		redirectUris, ok := vars["redirect_uri"]
		if !ok || len(redirectUris) == 0 {
			return nil, errors.New("missing redirect_uri")
		}
		clientIds, ok := vars["client_id"]
		if !ok || len(clientIds) == 0 {
			return nil, errors.New("missing client_id")
		}
		return &AccessTokenRequest{
			GrantType:   grantTypes[0],
			Code:        codes[0],
			RedirectUri: redirectUris[0],
			ClientId:    clientIds[0],
		}, nil
	}
}
