package endpoints

import (
	"context"
	"encoding/json"
	"github.com/searKing/golib/net/http_/oauth2/access"
	"github.com/searKing/golib/net/http_/oauth2/grant"
	"github.com/searKing/golib/net/http_/oauth2/grant/accesstoken"
	"github.com/searKing/golib/net/http_/oauth2/grant/authorize"
	"github.com/searKing/golib/net/http_/oauth2/grant/client"
	"github.com/searKing/golib/net/http_/oauth2/grant/implict"
	"github.com/searKing/golib/net/http_/oauth2/grant/resource"
	"github.com/searKing/golib/net/http_/oauth2/refresh"
	"net/http"
	"time"
)

type AuthorizationRequest struct {
	ClientId string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
}

const (
	defaultTimeFormat = time.RFC3339
)

// rfc6749 3.1
// Authorization endpoint - used by the client to obtain
// authorization from the resource owner via user-agent redirection.
type AuthorizationCodeGrantAuthorizationResult struct {
	Code string `json:"code"`
}
type ImplicitGrantAuthorizationResult struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   time.Time `json:"expires_in,omitempty"`
	Scope       string    `json:"scope,omitempty"`
}
type AuthorizationEndpoint struct {
	AuthorizationCodeGrantAuthenticationFunc func(ctx context.Context, authReq *AuthorizationRequest) (res *AuthorizationCodeGrantAuthorizationResult, err authorize.ErrorText)
	ImplicitGrantAuthenticationFunc          func(ctx context.Context, authReq *AuthorizationRequest) (res *ImplicitGrantAuthorizationResult, err implict.ErrorText)

	AuthorizationCodeGrantAccessTokenFunc                func(ctx context.Context, tokenReq *authorize.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)
	ResourceOwnerPasswordCredentialsGrantAccessTokenFunc func(ctx context.Context, tokenReq *resource.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)
	ClientCredentialsGrantAccessTokenFunc                func(ctx context.Context, tokenReq *client.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)
	RefreshTokenGrantAccessTokenFunc                     func(ctx context.Context, tokenReq *refresh.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)

	AuthorizateFunc func(ctx context.Context, token *accesstoken.AccessTokenType) (err accesstoken.ErrorText)
}

func (e *AuthorizationEndpoint) AuthenticationHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authReq, err := grant.RetrieveAuthorizationRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.1
		if authReq.ResponseType == "code" {
			e.authorizationCodeGrantAuthenticationHandler(ctx, authReq).ServeHTTP(w, r)
			return
		}
		if authReq.ResponseType == "token" {
			e.implictGrantAuthenticationHandler(ctx, authReq).ServeHTTP(w, r)
			return
		}
		e.unknownGrantAuthenticationHandler(ctx, authReq).ServeHTTP(w, r)
		return
	})
}
func (e *AuthorizationEndpoint) AccessTokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		grantType, body, err := grant.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if grantType == "authorization_code" {
			e.authorizationCodeGrantAccessTokenHandler(ctx, body).ServeHTTP(w, r)
			return
		}
		if grantType == "password" {
			e.resourceOwnerPasswordCredentialsGrantAccessTokenHandler(ctx, body).ServeHTTP(w, r)
			return
		}
		if grantType == "client_credentials" {
			e.clientCredentialsGrantAccessTokenHandler(ctx, body).ServeHTTP(w, r)
			return
		}
		if grantType == "refresh_token" {
			e.refreshTokenGrantAccessTokenHandler(ctx, body).ServeHTTP(w, r)
		}
		e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
		return
	})
}
func (e *AuthorizationEndpoint) AuthorizateHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenType, err := accesstoken.RetrieveAccessTokenType(ctx, r)
		if err != nil {
			e.authorizateRejected(accesstoken.ErrorTextUnauthorizedClient)
		}

		errCode := e.authorizate(ctx, accessTokenType)
		if errCode != "" {
			e.authorizateRejected(errCode)
			return
		}
	})
}
func (e *AuthorizationEndpoint) authorizationCodeGrantAuthenticationHandler(ctx context.Context, authReq *grant.AuthorizationRequest) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authReq.ResponseType != "code" {
			e.unknownGrantAuthenticationHandler(ctx, authReq).ServeHTTP(w, r)
			return
		}
		authRes, errCode := e.authorizationCodeGrantAuthentication(ctx, &AuthorizationRequest{
			ClientId: authReq.ClientId,
			Scope:    authReq.Scope,
		})
		if errCode == "" {
			e.authenticationRejected(authReq, errCode.String(), errCode.Description()).ServeHTTP(w, r)
			return
		}
		if authRes == nil {
			e.authenticationRejected(authReq, authorize.ErrorTextServerError.String(), authorize.ErrorTextServerError.Description()).ServeHTTP(w, r)
			return
		}
		authResp := &authorize.AuthorizationResponse{
			Code:  authRes.Code,
			State: authReq.State,
		}
		if authReq.RedirectUri != "" {
			w.WriteHeader(http.StatusFound)
			w.Header().Set("Location", authReq.RedirectUri+"?"+authResp.UrlEncode())
			return
		}
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write([]byte(authResp.UrlEncode()))
	})
}
func (e *AuthorizationEndpoint) implictGrantAuthenticationHandler(ctx context.Context, authReq *grant.AuthorizationRequest) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authReq.ResponseType != "token" {
			e.unknownGrantAuthenticationHandler(ctx, authReq).ServeHTTP(w, r)
			return
		}
		authRes, errCode := e.implicitGrantAuthentication(ctx, &AuthorizationRequest{
			ClientId: authReq.ClientId,
			Scope:    authReq.Scope,
		})
		if errCode == "" {
			e.authenticationRejected(authReq, errCode.String(), errCode.Description()).ServeHTTP(w, r)
			return
		}
		if authRes == nil {
			e.authenticationRejected(authReq, implict.ErrorTextServerError.String(), implict.ErrorTextServerError.Description()).ServeHTTP(w, r)
			return
		}
		accessTokenResp := &implict.AccessTokenResponse{
			AccessToken: authRes.AccessToken,
			TokenType:   authRes.TokenType,
			ExpiresIn:   authRes.ExpiresIn.Format(defaultTimeFormat),
			Scope: func() string {
				// rfc6749 4.2.2
				if authRes.Scope == authReq.Scope {
					return ""
				}
				return authRes.Scope
			}(),
			State: authReq.State,
		}
		if authReq.RedirectUri != "" {
			w.WriteHeader(http.StatusFound)
			w.Header().Set("Location", authReq.RedirectUri+"?"+accessTokenResp.UrlEncode())
			return
		}
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write([]byte(accessTokenResp.UrlEncode()))
	})
}
func (e *AuthorizationEndpoint) unknownGrantAuthenticationHandler(ctx context.Context, authReq *grant.AuthorizationRequest) http.Handler {
	return e.authenticationRejected(
		authReq,
		authorize.ErrorTextInvalidRequest.String(),
		authorize.ErrorTextInvalidRequest.Description())
}
func (e *AuthorizationEndpoint) authorizationCodeGrantAccessTokenHandler(ctx context.Context, body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := authorize.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.3
		if accessTokenReq.GrantType != "authorization_code" {
			e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		accessTokenResp, errCode := e.authorizationCodeGrantAccessToken(ctx, accessTokenReq)
		if errCode != "" {
			accessTokenErrResp := accesstoken.ErrorIssueResponse{
				Error:            errCode,
				ErrorDescription: errCode.Description(),
				ErrorUri:         "https://tools.ietf.org/pdf/rfc6749.pdf",
			}
			statusCode := http.StatusBadRequest
			if accessTokenErrResp.Error == accesstoken.ErrorTextInvalidClient {
				statusCode = http.StatusUnauthorized
			}
			w.WriteHeader(statusCode)

			accessTokenErrRespBytes, err := json.Marshal(&accessTokenErrResp)
			if err != nil {
				return
			}
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			w.Write(accessTokenErrRespBytes)
			return
		}

		accessTokenRespBytes, err := json.Marshal(&accessTokenResp)
		if err != nil {
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write(accessTokenRespBytes)
		return
	})
}
func (e *AuthorizationEndpoint) resourceOwnerPasswordCredentialsGrantAccessTokenHandler(ctx context.Context, body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := resource.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.3
		if accessTokenReq.GrantType != "password" {
			e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		accessTokenResp, errCode := e.resourceOwnerPasswordCredentialsGrantAccessToken(ctx, accessTokenReq)
		if errCode != "" {
			accessTokenErrResp := accesstoken.ErrorIssueResponse{
				Error:            errCode,
				ErrorDescription: errCode.Description(),
				ErrorUri:         "https://tools.ietf.org/pdf/rfc6749.pdf",
			}
			statusCode := http.StatusBadRequest
			if accessTokenErrResp.Error == accesstoken.ErrorTextInvalidClient {
				statusCode = http.StatusUnauthorized
			}
			w.WriteHeader(statusCode)

			accessTokenErrRespBytes, err := json.Marshal(&accessTokenErrResp)
			if err != nil {
				return
			}
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			w.Write(accessTokenErrRespBytes)
			return
		}

		accessTokenRespBytes, err := json.Marshal(&accessTokenResp)
		if err != nil {
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write(accessTokenRespBytes)
		return
	})
}
func (e *AuthorizationEndpoint) clientCredentialsGrantAccessTokenHandler(ctx context.Context, body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := client.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.4.2
		if accessTokenReq.GrantType != "client_credentials" {
			e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		accessTokenResp, errCode := e.clientCredentialsGrantAccessToken(ctx, accessTokenReq)
		if errCode != "" {
			e.accessTokenRejected(errCode)
			return
		}
		accessTokenRespBytes, err := json.Marshal(&accessTokenResp)
		if err != nil {
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write(accessTokenRespBytes)
		return
	})
}
func (e *AuthorizationEndpoint) refreshTokenGrantAccessTokenHandler(ctx context.Context, body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := refresh.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.4.2
		if accessTokenReq.GrantType != "refresh_token" {
			e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		accessTokenResp, errCode := e.refreshTokenGrantAccessToken(ctx, accessTokenReq)
		if errCode != "" {
			e.accessTokenRejected(errCode)
			return
		}
		accessTokenRespBytes, err := json.Marshal(&accessTokenResp)
		if err != nil {
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write(accessTokenRespBytes)
		return
	})
}

func (e *AuthorizationEndpoint) unknownGrantAccessTokenHandler(ctx context.Context) http.Handler {
	return e.accessTokenRejected(accesstoken.ErrorTextInvalidRequest)
}
func (e *AuthorizationEndpoint) authenticationRejected(authReq *grant.AuthorizationRequest, err, errDescription string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authResp := grant.AuthorizationErrorResponse{
			Error:            err,
			ErrorDescription: errDescription,
			ErrorUri:         "https://tools.ietf.org/pdf/rfc6749.pdf",
			State:            authReq.State,
		}
		if authReq.RedirectUri != "" {
			w.WriteHeader(http.StatusFound)
			w.Header().Set("Location", authReq.RedirectUri+"?"+authResp.UrlEncode())
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write([]byte(authResp.UrlEncode()))
	})
}
func (e *AuthorizationEndpoint) accessTokenRejected(err accesstoken.ErrorText) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authResp := accesstoken.ErrorIssueResponse{
			Error:            err,
			ErrorDescription: err.Description(),
			ErrorUri:         "https://tools.ietf.org/pdf/rfc6749.pdf",
		}
		statusCode := http.StatusBadRequest
		if authResp.Error == accesstoken.ErrorTextInvalidClient {
			statusCode = http.StatusUnauthorized
		}
		w.WriteHeader(statusCode)
		authRespBytes, err := json.Marshal(&authResp)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write([]byte(authRespBytes))
		return
	})
}
func (e *AuthorizationEndpoint) authorizateRejected(err accesstoken.ErrorText) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := access.NewWWWAuthentiate("Bearer", "")
		auth.SetAuthHeader(r)
		authResp := accesstoken.AccessTokenTypeErrorResponse{
			Error:            err,
			ErrorDescription: err.Description(),
			ErrorUri:         "https://tools.ietf.org/pdf/rfc6749.pdf",
		}
		w.WriteHeader(http.StatusUnauthorized)
		authRespBytes, err := json.Marshal(&authResp)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write([]byte(authRespBytes))
		return
	})
}

func (e *AuthorizationEndpoint) authorizationCodeGrantAuthentication(ctx context.Context, authReq *AuthorizationRequest) (res *AuthorizationCodeGrantAuthorizationResult, err authorize.ErrorText) {
	if e.AuthorizationCodeGrantAuthenticationFunc != nil {
		return e.AuthorizationCodeGrantAuthenticationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, authorize.ErrorTextUnsupportedResponseType
}
func (e *AuthorizationEndpoint) implicitGrantAuthentication(ctx context.Context, authReq *AuthorizationRequest) (res *ImplicitGrantAuthorizationResult, err implict.ErrorText) {
	if e.ImplicitGrantAuthenticationFunc != nil {
		return e.ImplicitGrantAuthenticationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, implict.ErrorTextUnsupportedResponseType
}
func (e *AuthorizationEndpoint) authorizationCodeGrantAccessToken(ctx context.Context, tokenReq *authorize.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText) {
	if e.AuthorizationCodeGrantAccessTokenFunc != nil {
		return e.AuthorizationCodeGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *AuthorizationEndpoint) resourceOwnerPasswordCredentialsGrantAccessToken(ctx context.Context, tokenReq *resource.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText) {
	if e.AuthorizationCodeGrantAccessTokenFunc != nil {
		return e.ResourceOwnerPasswordCredentialsGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *AuthorizationEndpoint) clientCredentialsGrantAccessToken(ctx context.Context, tokenReq *client.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText) {
	if e.ClientCredentialsGrantAccessTokenFunc != nil {
		return e.ClientCredentialsGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *AuthorizationEndpoint) refreshTokenGrantAccessToken(ctx context.Context, tokenReq *refresh.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText) {
	if e.AuthorizationCodeGrantAccessTokenFunc != nil {
		return e.RefreshTokenGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *AuthorizationEndpoint) authorizate(ctx context.Context, token *accesstoken.AccessTokenType) (err accesstoken.ErrorText) {
	if e.AuthorizateFunc != nil {
		return e.AuthorizateFunc(ctx, token)
	}
	// UnImplemented
	return accesstoken.ErrorTextUnauthorizedClient
}
