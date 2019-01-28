package endpoints

import (
	"context"
	"encoding/json"
	"github.com/searKing/golib/net/http_/oauth2/grant/accesstoken"
	"github.com/searKing/golib/net/http_/oauth2/grant/authorize"
	"net/http"
)

// rfc6749 3.1
// Authorization endpoint - used by the client to obtain
// authorization from the resource owner via user-agent redirection.
type AuthorizationEndpoint struct {
	AuthorizationFunc  func(ctx context.Context, authReq *authorize.AuthorizationRequest) (code string, err authorize.ErrorText)
	GetAccessTokenFunc func(ctx context.Context, tokenReq *authorize.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulResponse, err accesstoken.ErrorText)
}

func (e *AuthorizationEndpoint) AuthorizationHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authReq, err := authorize.RetrieveAuthorizationRequest(ctx, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.1
		if authReq.ResponseType != "code" {
			authResp := authorize.AuthorizationErrorResponse{
				Error:            authorize.ErrorTextInvalidRequest,
				ErrorDescription: authorize.ErrorTextInvalidRequest.Description(),
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
			return
		}
		code, errCode := e.Authorization(ctx, authReq)
		if errCode == "" {
			authResp := authorize.AuthorizationErrorResponse{
				Error:            errCode,
				ErrorDescription: errCode.Description(),
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
			return
		}
		authResp := &authorize.AuthorizationResponse{
			Code:  code,
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
func (e *AuthorizationEndpoint) GetAccessTokenHandler(ctx context.Context) http.Handler {
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
func (e *AuthorizationEndpoint) Authorization(ctx context.Context, authReq *authorize.AuthorizationRequest) (code string, err authorize.ErrorText) {
	if e.AuthorizationFunc != nil {
		return e.AuthorizationFunc(ctx, authReq)
	}
	// UnImplemented
	return "", authorize.ErrorTextUnsupportedResponseType
}
func (e *AuthorizationEndpoint) GetAccessToken(ctx context.Context, tokenReq *authorize.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulResponse, err accesstoken.ErrorText) {
	if e.GetAccessTokenFunc != nil {
		return e.GetAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
