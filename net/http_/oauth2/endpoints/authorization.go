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
	"log"
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
type AuthorizeAuthorizationResult struct {
	Code string `json:"code"`
}
type ImplicitAuthorizationResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
}
type AuthorizeAccessTokenRequest struct {
	Code     string `json:"code"`
	ClientId string `json:"client_id"`
}
type AuthorizeAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}
type ResourceAccessTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Scope    string `json:"scope,omitempty"`
}
type ClientAccessTokenRequest struct {
	Scope    string `json:"scope,omitempty"`
	UserID   string `json:"-"`
	Password string `json:"-"`
}
type RefreshAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope,omitempty"`
	UserID       string `json:"-"`
	Password     string `json:"-"`
}
type AuthorizationEndpoint struct {
	// If nil, logging is done via the log package's standard logger.
	Logger *log.Logger

	AbortFunc func(ctx context.Context)

	AuthorizationCodeGrantAuthenticationFunc func(ctx context.Context, authReq *AuthorizationRequest) (res *AuthorizeAuthorizationResult, err authorize.ErrorText)
	ImplicitGrantAuthenticationFunc          func(ctx context.Context, authReq *AuthorizationRequest) (res *ImplicitAuthorizationResult, err implict.ErrorText)

	AuthorizationCodeGrantAccessTokenFunc                func(ctx context.Context, tokenReq *AuthorizeAccessTokenRequest) (tokenResp *AuthorizeAccessTokenResponse, err accesstoken.ErrorText)
	ResourceOwnerPasswordCredentialsGrantAccessTokenFunc func(ctx context.Context, tokenReq *ResourceAccessTokenRequest) (tokenResp *AccessTokenResponse, err accesstoken.ErrorText)
	ClientCredentialsGrantAccessTokenFunc                func(ctx context.Context, tokenReq *ClientAccessTokenRequest) (tokenResp *AccessTokenResponse, err accesstoken.ErrorText)
	RefreshTokenGrantAccessTokenFunc                     func(ctx context.Context, tokenReq *RefreshAccessTokenRequest) (tokenResp *AccessTokenResponse, err accesstoken.ErrorText)

	AuthorizateFunc func(ctx context.Context, token *accesstoken.AccessTokenType) (err accesstoken.ErrorText)
}

func (e *AuthorizationEndpoint) logf(format string, v ...interface{}) {
	if e.Logger != nil {
		e.Logger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

func (e *AuthorizationEndpoint) AuthenticationHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authReq, err := grant.RetrieveAuthorizationRequest(ctx, r)
		if err != nil {
			e.logf("AuthenticationHandler: RetrieveAuthorizationRequest failed %s", err.Error())
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
		e.logf("AuthenticationHandler :response_type expect %s, actual %s", "code|token", authReq.ResponseType)
		e.unknownGrantAuthenticationHandler(ctx, authReq).ServeHTTP(w, r)
		return
	})
}
func (e *AuthorizationEndpoint) AccessTokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		grantType, err := grant.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			e.logf("AccessTokenHandler: RetrieveAccessTokenRequest failed %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if grantType == "authorization_code" {
			e.authorizationCodeGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		if grantType == "password" {
			e.resourceOwnerPasswordCredentialsGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		if grantType == "client_credentials" {
			e.clientCredentialsGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		if grantType == "refresh_token" {
			e.refreshTokenGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		e.logf("AccessTokenHandler :grant_type expect %s, actual %s",
			"authorization_code|password|client_credentials|refresh_token", grantType)
		e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
		return
	})
}
func (e *AuthorizationEndpoint) AuthorizateHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenType, err := accesstoken.RetrieveAccessTokenType(ctx, r)
		if err != nil {
			e.logf("AuthorizateHandler :RetrieveAccessTokenType failed %s", err.Error())
			e.authorizateRejected(ctx, accesstoken.ErrorTextUnauthorizedClient).ServeHTTP(w, r)
			return
		}

		errCode := e.authorizate(ctx, accessTokenType)
		if errCode != "" {
			e.logf("AuthorizateHandler :authorizate failed %s", errCode)
			e.authorizateRejected(ctx, errCode).ServeHTTP(w, r)
			return
		}
	})
}
func (e *AuthorizationEndpoint) authorizationCodeGrantAuthenticationHandler(ctx context.Context, authReq *grant.AuthorizationRequest) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authReq.ResponseType != "code" {
			e.logf("authorizationCodeGrantAuthenticationHandler :response_type expect %s, actual %s",
				"code", authReq.ResponseType)
			e.unknownGrantAuthenticationHandler(ctx, authReq).ServeHTTP(w, r)
			return
		}
		authRes, errCode := e.authorizationCodeGrantAuthentication(ctx, &AuthorizationRequest{
			ClientId: authReq.ClientId,
			Scope:    authReq.Scope,
		})
		if errCode != "" {
			e.logf("authorizationCodeGrantAuthentication failed %s", errCode)
			e.authenticationRejected(authReq, errCode.String(), errCode.Description()).ServeHTTP(w, r)
			return
		}
		if authRes == nil {
			e.logf("authorizationCodeGrantAuthentication authRes is nil")
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
			e.logf("implictGrantAuthenticationHandler :response_type expect %s, actual %s",
				"token", authReq.ResponseType)
			e.unknownGrantAuthenticationHandler(ctx, authReq).ServeHTTP(w, r)
			return
		}
		authRes, errCode := e.implicitGrantAuthentication(ctx, &AuthorizationRequest{
			ClientId: authReq.ClientId,
			Scope:    authReq.Scope,
		})
		if errCode != "" {
			e.logf("implicitGrantAuthentication failed %s", errCode)
			e.authenticationRejected(authReq, errCode.String(), errCode.Description()).ServeHTTP(w, r)
			return
		}
		if authRes == nil {
			e.logf("implicitGrantAuthentication authRes is nil")
			e.authenticationRejected(authReq, implict.ErrorTextServerError.String(), implict.ErrorTextServerError.Description()).ServeHTTP(w, r)
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
func (e *AuthorizationEndpoint) unknownGrantAuthenticationHandler(ctx context.Context, authReq *grant.AuthorizationRequest) http.Handler {
	return e.authenticationRejected(
		authReq,
		authorize.ErrorTextInvalidRequest.String(),
		authorize.ErrorTextInvalidRequest.Description())
}
func (e *AuthorizationEndpoint) authorizationCodeGrantAccessTokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := authorize.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			e.logf("authorizationCodeGrantAccessTokenHandler :RetrieveAccessTokenRequest failed %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.3
		if accessTokenReq.GrantType != "authorization_code" {
			e.logf("implictGrantAuthenticationHandler :grant_type expect %s, actual %s",
				"authorization_code", accessTokenReq.GrantType)
			e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		accessTokenResp, errCode := e.authorizationCodeGrantAccessToken(ctx, &AuthorizeAccessTokenRequest{
			Code:     accessTokenReq.Code,
			ClientId: accessTokenReq.ClientId,
		})
		if errCode != "" {
			e.accessTokenRejected(errCode).ServeHTTP(w, r)
			return
		}

		accessTokenRespBytes, err := json.Marshal(&accesstoken.SuccessfulIssueResponse{
			AccessToken:  accessTokenResp.AccessToken,
			TokenType:    accessTokenResp.TokenType,
			ExpiresIn:    accessTokenResp.ExpiresIn,
			RefreshToken: accessTokenResp.RefreshToken,
		})
		if err != nil {
			e.logf("implictGrantAuthenticationHandler :Marshal SuccessfulIssueResponse failed %s",
				err.Error())
			e.accessTokenRejected(accesstoken.ErrorTextUnauthorizedClient)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write(accessTokenRespBytes)
		return
	})
}
func (e *AuthorizationEndpoint) resourceOwnerPasswordCredentialsGrantAccessTokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := resource.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			e.logf("resourceOwnerPasswordCredentialsGrantAccessTokenHandler :RetrieveAccessTokenRequest failed %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.1.3
		if accessTokenReq.GrantType != "password" {
			e.logf("resourceOwnerPasswordCredentialsGrantAccessTokenHandler :grant_type expect %s, actual %s",
				"password", accessTokenReq.GrantType)
			e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		accessTokenResp, errCode := e.resourceOwnerPasswordCredentialsGrantAccessToken(ctx, &ResourceAccessTokenRequest{
			Username: accessTokenReq.Username,
			Password: accessTokenReq.Password,
			Scope:    accessTokenReq.Scope,
		})
		if errCode != "" {
			e.accessTokenRejected(errCode).ServeHTTP(w, r)
			return
		}

		accessTokenRespBytes, err := json.Marshal(&AccessTokenResponse{
			AccessToken:  accessTokenResp.AccessToken,
			TokenType:    accessTokenResp.TokenType,
			ExpiresIn:    accessTokenResp.ExpiresIn,
			RefreshToken: accessTokenResp.RefreshToken,
			Scope: func() string {
				// rfc6749 5.1
				// OPTIONAL, if identical to the scope requested by the client;
				// otherwise, REQUIRED.
				if accessTokenReq.Scope == accessTokenResp.Scope {
					return ""
				}
				return accessTokenResp.Scope
			}(),
		})
		if err != nil {
			e.logf("resourceOwnerPasswordCredentialsGrantAccessTokenHandler :Marshal AccessTokenResponse failed %s",
				err.Error())
			e.accessTokenRejected(accesstoken.ErrorTextUnauthorizedClient)
			return
		}
		// rfc6749 5.1
		// The authorization server MUST include the HTTP "Cache-Control"
		// response header field [RFC2616] with a value of "no-store" in any
		// response containing tokens, credentials, or other sensitive
		// information, as well as the "Pragma" response header field [RFC2616]
		// with a value of "no-cache".
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write(accessTokenRespBytes)
		return
	})
}
func (e *AuthorizationEndpoint) clientCredentialsGrantAccessTokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := client.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			e.logf("clientCredentialsGrantAccessTokenHandler :RetrieveAccessTokenRequest failed %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.4.2
		if accessTokenReq.GrantType != "client_credentials" {
			e.logf("clientCredentialsGrantAccessTokenHandler :grant_type expect %s, actual %s",
				"client_credentials", accessTokenReq.GrantType)
			e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		accessTokenResp, errCode := e.clientCredentialsGrantAccessToken(ctx, &ClientAccessTokenRequest{
			Scope:    accessTokenReq.Scope,
			UserID:   accessTokenReq.UserID,
			Password: accessTokenReq.Password,
		})
		if errCode != "" {
			e.logf("clientCredentialsGrantAccessTokenHandler :accessTokenRejected %",
				errCode)
			e.accessTokenRejected(errCode).ServeHTTP(w, r)
			return
		}
		accessTokenRespBytes, err := json.Marshal(&AccessTokenResponse{
			AccessToken:  accessTokenResp.AccessToken,
			TokenType:    accessTokenResp.TokenType,
			ExpiresIn:    accessTokenResp.ExpiresIn,
			RefreshToken: accessTokenResp.RefreshToken,
			Scope: func() string {
				if accessTokenReq.Scope == accessTokenResp.Scope {
					return ""
				}
				return accessTokenResp.Scope
			}(),
		})
		if err != nil {
			e.logf("clientCredentialsGrantAccessTokenHandler :Marshal AccessTokenResponse failed %s",
				err.Error())
			e.accessTokenRejected(accesstoken.ErrorTextUnauthorizedClient)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write(accessTokenRespBytes)
		return
	})
}
func (e *AuthorizationEndpoint) refreshTokenGrantAccessTokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenReq, err := refresh.RetrieveAccessTokenRequest(ctx, r)
		if err != nil {
			e.logf("refreshTokenGrantAccessTokenHandler :RetrieveAccessTokenRequest failed %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// rfc6749 4.4.2
		if accessTokenReq.GrantType != "refresh_token" {
			e.unknownGrantAccessTokenHandler(ctx).ServeHTTP(w, r)
			return
		}
		accessTokenResp, errCode := e.refreshTokenGrantAccessToken(ctx, &RefreshAccessTokenRequest{
			RefreshToken: accessTokenReq.RefreshToken,
			Scope:        accessTokenReq.Scope,
			UserID:       accessTokenReq.UserID,
			Password:     accessTokenReq.Password,
		})
		if errCode != "" {
			e.accessTokenRejected(errCode).ServeHTTP(w, r)
			return
		}
		accessTokenRespBytes, err := json.Marshal(&AccessTokenResponse{
			AccessToken:  accessTokenResp.AccessToken,
			TokenType:    accessTokenResp.TokenType,
			ExpiresIn:    accessTokenResp.ExpiresIn,
			RefreshToken: accessTokenResp.RefreshToken,
			Scope: func() string {
				if accessTokenReq.Scope == accessTokenResp.Scope {
					return ""
				}
				return accessTokenResp.Scope
			}(),
		})
		if err != nil {
			e.logf("refreshTokenGrantAccessTokenHandler :Marshal AccessTokenResponse failed %s",
				err.Error())
			e.accessTokenRejected(accesstoken.ErrorTextUnauthorizedClient)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
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
			e.logf("accessTokenRejected :Marshal ErrorIssueResponse failed %s",
				err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write([]byte(authRespBytes))
		return
	})
}
func (e *AuthorizationEndpoint) authorizateRejected(ctx context.Context, err accesstoken.ErrorText) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer e.abort(ctx)

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
			e.logf("authorizateRejected :Marshal AccessTokenTypeErrorResponse failed %s",
				err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write([]byte(authRespBytes))
		return
	})
}

func (e *AuthorizationEndpoint) authorizationCodeGrantAuthentication(ctx context.Context, authReq *AuthorizationRequest) (res *AuthorizeAuthorizationResult, err authorize.ErrorText) {
	if e.AuthorizationCodeGrantAuthenticationFunc != nil {
		return e.AuthorizationCodeGrantAuthenticationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, authorize.ErrorTextUnsupportedResponseType
}
func (e *AuthorizationEndpoint) implicitGrantAuthentication(ctx context.Context, authReq *AuthorizationRequest) (res *ImplicitAuthorizationResult, err implict.ErrorText) {
	if e.ImplicitGrantAuthenticationFunc != nil {
		return e.ImplicitGrantAuthenticationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, implict.ErrorTextUnsupportedResponseType
}
func (e *AuthorizationEndpoint) authorizationCodeGrantAccessToken(ctx context.Context, tokenReq *AuthorizeAccessTokenRequest) (tokenResp *AuthorizeAccessTokenResponse, err accesstoken.ErrorText) {
	if e.AuthorizationCodeGrantAccessTokenFunc != nil {
		return e.AuthorizationCodeGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *AuthorizationEndpoint) resourceOwnerPasswordCredentialsGrantAccessToken(ctx context.Context, tokenReq *ResourceAccessTokenRequest) (tokenResp *AccessTokenResponse, err accesstoken.ErrorText) {
	if e.ResourceOwnerPasswordCredentialsGrantAccessTokenFunc != nil {
		return e.ResourceOwnerPasswordCredentialsGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *AuthorizationEndpoint) clientCredentialsGrantAccessToken(ctx context.Context, tokenReq *ClientAccessTokenRequest) (tokenResp *AccessTokenResponse, err accesstoken.ErrorText) {
	if e.ClientCredentialsGrantAccessTokenFunc != nil {
		return e.ClientCredentialsGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *AuthorizationEndpoint) refreshTokenGrantAccessToken(ctx context.Context, tokenReq *RefreshAccessTokenRequest) (tokenResp *AccessTokenResponse, err accesstoken.ErrorText) {
	if e.RefreshTokenGrantAccessTokenFunc != nil {
		return e.RefreshTokenGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *AuthorizationEndpoint) authorizate(ctx context.Context, token *accesstoken.AccessTokenType) (err accesstoken.ErrorText) {
	if e.AuthorizateFunc == nil {
		// UnImplemented
		e.abort(ctx)
		return accesstoken.ErrorTextUnauthorizedClient
	}
	err = e.AuthorizateFunc(ctx, token)
	if err != "" {
		e.abort(ctx)
		return err
	}
	return ""
}
func (e *AuthorizationEndpoint) abort(ctx context.Context) {
	if e.AbortFunc != nil {
		e.AbortFunc(ctx)
	}
	// UnImplemented
	return
}
