package implict

import "net/url"

// rfc6749 4.2.1
//	GET /authorize?response_type=token&client_id=s6BhdRkqt3&state=xyz
//	&redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
//	Host: server.example.com
//type AuthorizationRequest grant.AuthorizationRequest

// rfc6749 4.2.2
//	HTTP/1.1 302 Found
//	Location: http://example.com/cb#access_token=2YotnFZFEjr1zCsicMWpAA
//	&state=xyz&token_type=example&expires_in=3600
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
	State       string `json:"state,omitempty" options:"pair"`
}

func (resp *AccessTokenResponse) UrlEncode() string {
	val := url.Values{
		"access_token": []string{resp.AccessToken},
		"token_type":   []string{resp.TokenType},
	}

	if resp.ExpiresIn != "" {
		val["expires_in"] = []string{resp.ExpiresIn}
	}
	if resp.Scope != "" {
		val["scope"] = []string{resp.Scope}
	}
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
		return `The client is not authorized to request an access token
 using this method.`
	case ErrorTextAccessDenied:
		return `The resource owner or authorization server denied the
 request.`
	case ErrorTextUnsupportedResponseType:
		return `The authorization server does not support obtaining an
 access token using this method.`
	case ErrorTextInvalidScope:
		return `The requested scope is invalid, unknown, or malformed.`
	case ErrorTextServerError:
		return `The authorization server encountered an unexpected
 condition that prevented it from fulfilling the request.
 (This error code is needed because a 500 Internal Server
 Error HTTP status code cannot be returned to the client
 via an HTTP redirect.)`
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
