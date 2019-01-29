package accesstoken

// rfc6749 5.1
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
type SuccessfulIssueResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
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

func (e ErrorText) String() string {
	return string(e)
}

func (e ErrorText) Description() string {
	switch e {
	case ErrorTextInvalidRequest:
		return `The request is missing a required parameter, includes an
 unsupported parameter value (other than grant type),
 repeats a parameter, includes multiple credentials,
 utilizes more than one mechanism for authenticating the
 client, or is otherwise malformed.`
	case ErrorTextInvalidClient:
		return `Client authentication failed (e.g., unknown client, no
 client authentication included, or unsupported
 authentication method). The authorization server MAY
 return an HTTP 401 (Unauthorized) status code to indicate
 which HTTP authentication schemes are supported. If the
 client attempted to authenticate via the "Authorization"
 request header field, the authorization server MUST
 respond with an HTTP 401 (Unauthorized) status code and
 include the "WWW-Authenticate" response header field
 matching the authentication scheme used by the client.`
	case ErrorTextInvalidGrant:
		return `The provided authorization grant (e.g., authorization
 code, resource owner credentials) or refresh token is
 invalid, expired, revoked, does not match the redirection
 URI used in the authorization request, or was issued to
 another client.`
	case ErrorTextUnauthorizedClient:
		return `The authenticated client is not authorized to use this
 authorization grant type.`
	case ErrorTextUnsupportedGrantType:
		return `The authorization grant type is not supported by the
 authorization server.`
	case ErrorTextInvalidScope:
		return `The requested scope is invalid, unknown, malformed, or
 exceeds the scope granted by the resource owner.`
	default:
		return e.String()
	}
}

// rfc6749 5.2
//	HTTP/1.1 400 Bad Request
//	Content-Type: application/json;charset=UTF-8
//	Cache-Control: no-store
//	Pragma: no-cache
//
//	{
//	"error":"invalid_request"
//	}
type ErrorIssueResponse struct {
	Error            ErrorText `json:"error"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ErrorUri         string    `json:"error_uri,omitempty"`
}
