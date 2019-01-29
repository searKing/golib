package endpoints

import (
	"context"
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

type JWTAuthorizationEndpoint struct {
	// Duration that a jwt access-token is valid. Optional, defaults to one hour.
	AccessExpireIn time.Duration `json:"access_expire_in,omitempty"`
	// Duration that a jwt refresh-token is valid. Optional, defaults to seven days.
	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	RefreshExpireIn time.Duration `json:"refresh_expire_in,omitempty"`

	AuthorizationCodeGrantAuthorizationFunc func(ctx context.Context, authReq *grant.AuthorizationRequest) (res *AuthorizationCodeGrantAuthorizationResult, err authorize.ErrorText)
	ImplicitGrantAuthorizationFunc          func(ctx context.Context, authReq *grant.AuthorizationRequest) (res *ImplicitGrantAuthorizationResult, err implict.ErrorText)

	AuthorizationCodeGrantAccessTokenFunc                func(ctx context.Context, tokenReq *authorize.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)
	ResourceOwnerPasswordCredentialsGrantAccessTokenFunc func(ctx context.Context, tokenReq *resource.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)
	ClientCredentialsGrantAccessTokenFunc                func(ctx context.Context, tokenReq *client.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)
	RefreshTokenGrantAccessTokenFunc                     func(ctx context.Context, tokenReq *refresh.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)

	// TimeNowFunc provides the current time. You can override it to use another time value.
	// This is useful for testing or if your server uses a different time zone than your tokens.
	TimeNowFunc func(ctx context.Context) time.Time
	auth        *AuthorizationEndpoint
}

func (e *JWTAuthorizationEndpoint) lazyInit() {
	if e.auth != nil {
		return
	}
	e.auth = &AuthorizationEndpoint{
		AuthorizationCodeGrantAuthenticationFunc:             e.AuthorizationCodeGrantAuthorizationFunc,
		ImplicitGrantAuthenticationFunc:                      e.ImplicitGrantAuthorizationFunc,
		AuthorizationCodeGrantAccessTokenFunc:                e.AuthorizationCodeGrantAccessTokenFunc,
		ResourceOwnerPasswordCredentialsGrantAccessTokenFunc: e.ResourceOwnerPasswordCredentialsGrantAccessTokenFunc,
		ClientCredentialsGrantAccessTokenFunc:                e.ClientCredentialsGrantAccessTokenFunc,
		RefreshTokenGrantAccessTokenFunc:                     e.RefreshTokenGrantAccessTokenFunc,
	}
}

func (e *JWTAuthorizationEndpoint) AuthorizationHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e.lazyInit()
		e.auth.AuthenticationHandler(ctx).ServeHTTP(w, r)
		return
	})
}
func (e *JWTAuthorizationEndpoint) AccessTokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e.lazyInit()
		e.auth.AccessTokenHandler(ctx).ServeHTTP(w, r)
		return
	})
}

func (e *JWTAuthorizationEndpoint) authorizationCodeGrantAuthorization(ctx context.Context, authReq *grant.AuthorizationRequest) (res *AuthorizationCodeGrantAuthorizationResult, err authorize.ErrorText) {
	if e.AuthorizationCodeGrantAuthorizationFunc != nil {
		return e.AuthorizationCodeGrantAuthorizationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, authorize.ErrorTextUnsupportedResponseType
}
func (e *JWTAuthorizationEndpoint) implicitGrantAuthorization(ctx context.Context, authReq *grant.AuthorizationRequest) (res *ImplicitGrantAuthorizationResult, err implict.ErrorText) {
	if e.ImplicitGrantAuthorizationFunc != nil {
		return e.ImplicitGrantAuthorizationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, implict.ErrorTextUnsupportedResponseType
}
func (e *JWTAuthorizationEndpoint) authorizationCodeGrantAccessToken(ctx context.Context, tokenReq *authorize.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText) {
	if e.AuthorizationCodeGrantAccessTokenFunc != nil {
		return e.AuthorizationCodeGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *JWTAuthorizationEndpoint) resourceOwnerPasswordCredentialsGrantAccessToken(ctx context.Context, tokenReq *resource.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText) {
	if e.AuthorizationCodeGrantAccessTokenFunc != nil {
		return e.ResourceOwnerPasswordCredentialsGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *JWTAuthorizationEndpoint) clientCredentialsGrantAccessToken(ctx context.Context, tokenReq *client.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText) {
	if e.ClientCredentialsGrantAccessTokenFunc != nil {
		return e.ClientCredentialsGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}
func (e *JWTAuthorizationEndpoint) refreshTokenGrantAccessToken(ctx context.Context, tokenReq *refresh.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText) {
	if e.AuthorizationCodeGrantAccessTokenFunc != nil {
		return e.RefreshTokenGrantAccessTokenFunc(ctx, tokenReq)
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}

// TimeNowFunc provides the current time. You can override it to use another time value.
// This is useful for testing or if your server uses a different time zone than your tokens.
func (e *JWTAuthorizationEndpoint) TimeNow(ctx context.Context) time.Time {
	if e.TimeNowFunc != nil {
		return e.TimeNowFunc(ctx)
	}
	return time.Now()
}
