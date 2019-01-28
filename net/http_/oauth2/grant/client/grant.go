package client

// rfc6749 4.4.1
// Since the client authentication is used as the authorization grant,
// no additional authorization request is needed.

// rfc6749 4.4.2
// POST /token HTTP/1.1
// Host: server.example.com
// Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
// Content-Type: application/x-www-form-urlencoded
// grant_type=client_credentials
type AccessTokenRequest struct {
	GrantType string `json:"grant_type"`
	Scope     string `json:"scope,omitempty"`
	UserID    string `json:"-"`
	Password  string `json:"-"`
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
//type AccessTokenResponse accesstoken.SuccessfulResponse
//type AccessTokenErrorResponse accesstoken.ErrorResponse
