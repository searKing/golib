package endpoints

import (
	"context"
	"encoding/json"
	"github.com/searKing/golib/net/http_/oauth2/grant"
	"github.com/searKing/golib/net/http_/oauth2/grant/accesstoken"
	"github.com/searKing/golib/net/http_/oauth2/grant/authorize"
	"github.com/searKing/golib/net/http_/oauth2/grant/implict"
	"net/http"
)

// rfc6749 3.1
// Authorization endpoint - used by the client to obtain
// authorization from the resource owner via user-agent redirection.
type AuthorizationCodeGrantAuthorizationResult struct {
	Code string `json:"code"`
}
type ImplicitGrantAuthorizationResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
}
type AuthorizationEndpoint struct {
	AuthorizationCodeGrantAuthorizationFunc func(ctx context.Context, authReq *grant.AuthorizationRequest) (res *AuthorizationCodeGrantAuthorizationResult, err authorize.ErrorText)
	ImplicitGrantAuthorizationFunc          func(ctx context.Context, authReq *grant.AuthorizationRequest) (res *ImplicitGrantAuthorizationResult, err implict.ErrorText)

	GetAccessTokenFunc func(ctx context.Context, tokenReq *authorize.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulResponse, err accesstoken.ErrorText)
}

func (e *AuthorizationEndpoint) AuthorizationHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authReq, err := grant.RetrieveAuthorizationRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.1
		if authReq.ResponseType == "code" {
			e.AuthorizationCodeGrantAuthorizationHandler(ctx, authReq).ServeHTTP(w, r)
			return
		}
		if authReq.ResponseType == "token" {
			e.ImplictGrantAuthorizationHandler(ctx, authReq).ServeHTTP(w, r)
			return
		}
		e.UnknownGrantAuthorizationHandler(ctx, authReq).ServeHTTP(w, r)
		return
	})
}
func (e *AuthorizationEndpoint) AuthorizationCodeGrantAuthorizationHandler(ctx context.Context, authReq *grant.AuthorizationRequest) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authReq.ResponseType != "code" {
			return
		}
		authRes, errCode := e.AuthorizationCodeGrantAuthorization(ctx, authReq)
		if errCode == "" {
			e.AuthorizationRejected(authReq, errCode.String(), errCode.Description()).ServeHTTP(w, r)
			return
		}
		if authRes == nil {
			e.AuthorizationRejected(authReq, authorize.ErrorTextServerError.String(), authorize.ErrorTextServerError.Description()).ServeHTTP(w, r)
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
func (e *AuthorizationEndpoint) ImplictGrantAuthorizationHandler(ctx context.Context, authReq *grant.AuthorizationRequest) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authReq.ResponseType != "token" {
			return
		}
		authRes, errCode := e.ImplicitGrantAuthorization(ctx, authReq)
		if errCode == "" {
			e.AuthorizationRejected(authReq, errCode.String(), errCode.Description()).ServeHTTP(w, r)
			return
		}
		if authRes == nil {
			e.AuthorizationRejected(authReq, implict.ErrorTextServerError.String(), implict.ErrorTextServerError.Description()).ServeHTTP(w, r)
			return
		}
		accessTokenResp := &implict.AccessTokenResponse{
			AccessToken: authRes.AccessToken,
			TokenType:   authRes.TokenType,
			ExpiresIn:   authRes.ExpiresIn,
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
func (e *AuthorizationEndpoint) UnknownGrantAuthorizationHandler(ctx context.Context, authReq *grant.AuthorizationRequest) http.Handler {
	return e.AuthorizationRejected(
		authReq,
		authorize.ErrorTextInvalidRequest.String(),
		authorize.ErrorTextInvalidRequest.Description())
}
func (e *AuthorizationEndpoint) GetAccessTokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		grantType, body, err := grant.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if grantType == "authorization_code" {
			e.AuthorizationCodeGrantAccessTokenHandler(ctx, body).ServeHTTP(w, r)
			return
		}
		if grantType == "password" {
			e.ImplicitGrantAccessTokenHandler(ctx, body).ServeHTTP(w, r)
			return
		}
		e.UnknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
		return
	})
}
func (e *AuthorizationEndpoint) AuthorizationCodeGrantAccessTokenHandler(ctx context.Context, body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := authorize.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.3
		if accessTokenReq.GrantType != "authorization_code" {
			authResp := accesstoken.ErrorResponse{
				Error:            accesstoken.ErrorTextInvalidGrant,
				ErrorDescription: accesstoken.ErrorTextInvalidGrant.Description(),
				ErrorUri:         "https://tools.ietf.org/pdf/rfc6749.pdf",
			}
			w.WriteHeader(http.StatusBadRequest)
			authRespBytes, err := json.Marshal(&authResp)
			if err != nil {
				return
			}
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			w.Write([]byte(authRespBytes))
			return
		}
		accessTokenResp, errCode := e.GetAccessToken(ctx, accessTokenReq)
		if errCode != "" {
			accessTokenErrResp := accesstoken.ErrorResponse{
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
func (e *AuthorizationEndpoint) ImplicitGrantAccessTokenHandler(ctx context.Context, body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := authorize.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.3
		if accessTokenReq.GrantType != "authorization_code" {
			authResp := accesstoken.ErrorResponse{
				Error:            accesstoken.ErrorTextInvalidGrant,
				ErrorDescription: accesstoken.ErrorTextInvalidGrant.Description(),
				ErrorUri:         "https://tools.ietf.org/pdf/rfc6749.pdf",
			}
			w.WriteHeader(http.StatusBadRequest)
			authRespBytes, err := json.Marshal(&authResp)
			if err != nil {
				return
			}
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			w.Write([]byte(authRespBytes))
			return
		}
		accessTokenResp, errCode := e.GetAccessToken(ctx, accessTokenReq)
		if errCode != "" {
			accessTokenErrResp := accesstoken.ErrorResponse{
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
func (e *AuthorizationEndpoint) UnknownGrantAccessTokenHandler(ctx context.Context) http.Handler {
	return e.AccessTokenRejected(accesstoken.ErrorTextInvalidRequest)
}
func (e *AuthorizationEndpoint) AuthorizationRejected(authReq *grant.AuthorizationRequest, err, errDescription string) http.Handler {
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
func (e *AuthorizationEndpoint) AccessTokenRejected(err accesstoken.ErrorText) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authResp := accesstoken.ErrorResponse{
			Error:            accesstoken.ErrorTextInvalidGrant,
			ErrorDescription: accesstoken.ErrorTextInvalidGrant.Description(),
			ErrorUri:         "https://tools.ietf.org/pdf/rfc6749.pdf",
		}
		w.WriteHeader(http.StatusBadRequest)
		authRespBytes, err := json.Marshal(&authResp)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write([]byte(authRespBytes))
		return
	})
}

func (e *AuthorizationEndpoint) AuthorizationCodeGrantAuthorization(ctx context.Context, authReq *grant.AuthorizationRequest) (res *AuthorizationCodeGrantAuthorizationResult, err authorize.ErrorText) {
	if e.AuthorizationCodeGrantAuthorizationFunc != nil {
		return e.AuthorizationCodeGrantAuthorizationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, authorize.ErrorTextUnsupportedResponseType
}
func (e *AuthorizationEndpoint) ImplicitGrantAuthorization(ctx context.Context, authReq *grant.AuthorizationRequest) (res *ImplicitGrantAuthorizationResult, err implict.ErrorText) {
	if e.ImplicitGrantAuthorizationFunc != nil {
		return e.ImplicitGrantAuthorizationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, implict.ErrorTextUnsupportedResponseType
}
func (e *AuthorizationEndpoint) GetAccessToken(ctx context.Context, tokenReq *authorize.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulResponse, err accesstoken.ErrorText) {
	if e.GetAccessTokenFunc != nil {
		return e.GetAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
