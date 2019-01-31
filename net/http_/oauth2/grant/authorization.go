package grant

import (
	"net/url"
)

// rfc6749 4.1.1
// GET /authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz
//        &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
//    Host: server.example.com
type AuthorizationRequest struct {
	ResponseType string `json:"response_type" const:"code,token"`
	ClientId     string `json:"client_id"`
	RedirectUri  string `json:"redirect_uri,omitempty"`
	Scope        string `json:"scope,omitempty"`
	State        string `json:"state,omitempty"`
}

// rfc6749 4.1.2
// HTTP/1.1 302 Found
// Location: https://client.example.com/cb?code=SplxlOBeZQQYbYS6WxSbIA
// 	&state=xyz
// see authorize/grant/AuthorizationResponse

// rfc6749 4.1.2.1
//	HTTP/1.1 302 Found
//	Location: https://client.example.com/cb?error=access_denied&state=xyz
// rfc6749 4.2.2.1
// HTTP/1.1 302 Found
// Location: https://client.example.com/cb?code=SplxlOBeZQQYbYS6WxSbIA
// &state=xyz
type AuthorizationErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorUri         string `json:"error_uri,omitempty"`
	State            string `json:"state,optional" options:"pair"`
}

func (resp *AuthorizationErrorResponse) UrlEncode() string {
	val := url.Values{"error": []string{string(resp.Error)}}
	if resp.ErrorDescription != "" {
		val["error_description"] = []string{resp.ErrorDescription}
	}
	if resp.ErrorUri != "" {
		val["error_uri"] = []string{resp.ErrorUri}
	}
	if resp.State != "" {
		val["state"] = []string{resp.State}
	}
	return val.Encode()
}
