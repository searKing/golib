package implict

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

// rfc6749 4.2.1
// GET /authorize?response_type=token&client_id=s6BhdRkqt3&state=xyz
//		 &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
// Host: server.example.com
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
