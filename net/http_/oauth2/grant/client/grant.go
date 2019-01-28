package client

import (
	"github.com/searKing/golib/net/http_/oauth2/grant/accesstoken"
	"github.com/searKing/golib/net/http_/oauth2/grant/authorize"
)

// rfc6749 4.4.1
// Since the client authentication is used as the authorization grant,
// no additional authorization request is needed.
type AuthorizationRequest struct {
	authorize.AuthorizationRequest
}
type AuthorizationResponse struct {
	authorize.AuthorizationResponse
}

// rfc6749 4.4.2
// POST /token HTTP/1.1
// Host: server.example.com
// Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
// Content-Type: application/x-www-form-urlencoded
// grant_type=client_credentials
type AccessTokenRequest struct {
	GrantType string `json:"grant_type"`
	Scope     string `json:"scope,omitempty"`
}

// rfc6749 4.4.3
// HTTP/1.1 200 OK
// Content-Type: application/json;charset=UTF-8
// Cache-Control: no-store
// Pragma: no-cache
//
// {
//	"access_token":"2YotnFZFEjr1zCsicMWpAA",
//	"token_type":"example",
//	"expires_in":3600,
//	"example_parameter":"example_value"
// }
type AccessTokenResponse struct {
	accesstoken.SuccessfulResponse
}
type AccessTokenErrorResponse struct {
	accesstoken.ErrorResponse
}
