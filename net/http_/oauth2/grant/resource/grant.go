package resource

import "github.com/searKing/golib/net/http_/oauth2/grant/accesstoken"

// rfc6749 4.3.1
// The method through which the client obtains the resource owner
// credentials is beyond the scope of this specification. The client
// MUST discard the credentials once an access token has been obtained.

// rfc6749 4.3.2
// POST /token HTTP/1.1
// Host: server.example.com
// Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
// Content-Type: application/x-www-form-urlencoded
//
// grant_type=password&username=johndoe&password=A3ddj3w
type AccessTokenRequest struct {
	GrantType string `json:"grant_type"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Scope     string `json:"scope,omitempty"`
}

// rfc6749 4.3.3
// HTTP/1.1 200 OK
// Content-Type: application/json;charset=UTF-8
// Cache-Control: no-store
// Pragma: no-cache
//
// {
//	"access_token":"2YotnFZFEjr1zCsicMWpAA",
//	"token_type":"example",
//	"expires_in":3600,
//	"refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
//	"example_parameter":"example_value"
// }
type AccessTokenResponse struct {
	accesstoken.SuccessfulIssueResponse
}
type AccessTokenErrorResponse struct {
	accesstoken.ErrorIssueResponse
}

type ErrorText string

const (
	ErrorTextInvalidRequest       ErrorText = "invalid_request"
	ErrorTextInvalidClient        ErrorText = "invalid_client"
	ErrorTextInvalidGrant         ErrorText = "invalid_grant"
	ErrorTextUnauthorizedClient   ErrorText = "unauthorized_client"
	ErrorTextUnsupportedGrantType ErrorText = "unsupported_grant_type"
	ErrorTextInvalidScope         ErrorText = "invalid_scope"
)
