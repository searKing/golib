package accesstoken

// rfc6749 7.1
//GET /resource/1 HTTP/1.1
//Host: example.com
//Authorization: Bearer mF_9.B5f-4.1JqM
// or
//GET /resource/1 HTTP/1.1
//Host: example.com
//Authorization: MAC id="h480djs93hd8",
//nonce="274312:dj83hs9s",
//mac="kDZvddkndxvhGRXZhvuDjEWhGeE="
type AccessTokenType struct {
	TokenType   string `json:"-"`
	AccessToken string `json:"access_token"`
}

// rfc6749 7.2
type AccessTokenTypeErrorResponse struct {
	Error            ErrorText `json:"error"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ErrorUri         string    `json:"error_uri,omitempty"`
}
