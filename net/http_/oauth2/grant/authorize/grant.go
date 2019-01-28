package authorize

import (
	"net/url"
)

// rfc6749 4.1.1
// GET /authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz
//        &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
//    Host: server.example.com
type AuthorizationRequest struct {
	ResponseType string `json:"response_type" const:"code"`
	ClientId     string `json:"client_id"`
	RedirectUri  string `json:"redirect_uri,omitempty"`
	Scope        string `json:"scope,omitempty"`
	State        string `json:"state,omitempty"`
}

// rfc6749 4.1.2
//	HTTP/1.1 302 Found
//	Location: https://client.example.com/cb?code=SplxlOBeZQQYbYS6WxSbIA
//				&state=xyz
type AuthorizationResponse struct {
	Code  string `json:"code"`
	State string `json:"state,omitempty" options:"pair"`
}

func (resp *AuthorizationResponse) UrlEncode() string {
	val := url.Values{"code": []string{resp.Code}}
	if resp.State != "" {
		val["state"] = []string{resp.State}
	}
	return val.Encode()
}

type ErrorText string

const (
	ErrorTextInvalidRequest          ErrorText = "invalid_request"
	ErrorTextUnauthorizedClient      ErrorText = "unauthorized_client"
	ErrorTextAccessDenied            ErrorText = "access_denied"
	ErrorTextUnsupportedResponseType ErrorText = "unsupported_response_type"
	ErrorTextInvalidScope            ErrorText = "invalid_scope"
	ErrorTextServerError             ErrorText = "server_error"
	ErrorTextTemporarilyUnavailable  ErrorText = "temporarily_unavailable"
)

func (e ErrorText) String() string {
	return string(e)
}

func (e ErrorText) Description() string {
	switch e {
	case ErrorTextInvalidRequest:
		return `The request is missing a required parameter, includes an
 invalid parameter value, includes a parameter more than
 once, or is otherwise malformed.`
	case ErrorTextUnauthorizedClient:
		return `The client is not authorized to request an authorization
 code using this method.`
	case ErrorTextAccessDenied:
		return `The resource owner or authorization server denied the
 request.`
	case ErrorTextUnsupportedResponseType:
		return `The authorization server does not support obtaining an
 authorization code using this method.`
	case ErrorTextInvalidScope:
		return `The requested scope is invalid, unknown, or malformed.`
	case ErrorTextServerError:
		return `The authorization server encountered an unexpected
 condition that prevented it from fulfilling the request.
 (This error code is needed because a 500 Internal Server
 Error HTTP status code cannot be returned to the client
 via an HTTP redirect.`
	case ErrorTextTemporarilyUnavailable:
		return `The authorization server is currently unable to handle
 the request due to a temporary overloading or maintenance
 of the server. (This error code is needed because a 503
 Service Unavailable HTTP status code cannot be returned
 to the client via an HTTP redirect.)`
	default:
		return e.String()
	}
}

// rfc6749 4.1.2.1
//	HTTP/1.1 302 Found
//	Location: https://client.example.com/cb?error=access_denied&state=xyz
type AuthorizationErrorResponse struct {
	Error            ErrorText `json:"error"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ErrorUri         string    `json:"error_uri,omitempty"`
	State            string    `json:"state,omitempty" options:"pair"`
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

// rfc6749 4.1.3
// POST /token HTTP/1.1
// Host: server.example.com
// Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
// Content-Type: application/x-www-form-urlencoded
//
// grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
// &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
type AccessTokenRequest struct {
	GrantType   string `json:"grant_type"`
	Code        string `json:"code"`
	RedirectUri string `json:"redirect_uri"`
	ClientId    string `json:"client_id"`
}

// rfc6749 4.1.4
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
//type AccessTokenResponse accesstoken.SuccessfulResponse
//type AccessTokenErrorResponse accesstoken.ErrorResponse
