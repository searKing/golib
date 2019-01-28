package refresh

import "github.com/searKing/golib/net/http_/oauth2/grant/accesstoken"

// rfc6749 6
// POST /token HTTP/1.1
// Host: server.example.com
// Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
// Content-Type: application/x-www-form-urlencoded
// grant_type=refresh_token&refresh_token=tGzv3JOkF0XG5Qx2TlKWIA
type RefreshRequest struct {
	GrantType    string `json:"grant_type" const:"refresh_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope,omitempty"`
}

// rfc6749 6
type AccessTokenResponse struct {
	accesstoken.SuccessfulResponse
}
type AccessTokenErrorResponse struct {
	accesstoken.ErrorResponse
}
