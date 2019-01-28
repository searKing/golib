package extension

import "github.com/searKing/golib/net/http_/oauth2/grant/accesstoken"

// rfc6749 4.5
// POST /token HTTP/1.1
// Host: server.example.com
// Content-Type: application/x-www-form-urlencoded
// grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Asaml2-
// bearer&assertion=PEFzc2VydGlvbiBJc3N1ZUluc3RhbnQ9IjIwMTEtMDU
// [...omitted for brevity...]aG5TdGF0ZW1lbnQ-PC9Bc3NlcnRpb24-
type AuthorizationRequest struct {
	GrantType string `json:"grant_type"`
}

// rfc6749 4.5
type AccessTokenResponse struct {
	accesstoken.SuccessfulResponse
}
type AccessTokenErrorResponse struct {
	accesstoken.ErrorResponse
}
