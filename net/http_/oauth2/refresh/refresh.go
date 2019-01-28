package refresh

// rfc6749 6
// POST /token HTTP/1.1
// Host: server.example.com
// Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
// Content-Type: application/x-www-form-urlencoded
// grant_type=refresh_token&refresh_token=tGzv3JOkF0XG5Qx2TlKWIA
type AccessTokenRequest struct {
	GrantType    string `json:"grant_type" const:"refresh_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope,omitempty"`
	UserID       string `json:"-"`
	Password     string `json:"-"`
}

// rfc6749 6
//type AccessTokenResponse accesstoken.SuccessfulResponse
//type AccessTokenErrorResponse accesstoken.ErrorResponse
