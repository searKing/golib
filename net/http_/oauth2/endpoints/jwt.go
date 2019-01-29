package endpoints

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/searKing/golib/net/http_/auth/jwt_"
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
	Key *jwt_.AuthKey `options:"required"`
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
	RefreshTokenGrantAccessTokenFunc                     func(ctx context.Context, tokenReq *RefreshToken) (tokenResp *accesstoken.SuccessfulIssueResponse, err accesstoken.ErrorText)

	AuthorizateFunc func(ctx context.Context, claims jwt.MapClaims) (err accesstoken.ErrorText)
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
		AuthorizationCodeGrantAuthenticationFunc:             e.authorizationCodeGrantAuthorization,
		ImplicitGrantAuthenticationFunc:                      e.implicitGrantAuthorization,
		AuthorizationCodeGrantAccessTokenFunc:                e.authorizationCodeGrantAccessToken,
		ResourceOwnerPasswordCredentialsGrantAccessTokenFunc: e.resourceOwnerPasswordCredentialsGrantAccessToken,
		ClientCredentialsGrantAccessTokenFunc:                e.clientCredentialsGrantAccessToken,
		RefreshTokenGrantAccessTokenFunc:                     e.refreshTokenGrantAccessToken,
		AuthorizateFunc:                                      e.authorizate,
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
func (e *JWTAuthorizationEndpoint) AuthorizateHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e.lazyInit()
		e.auth.AuthorizateHandler(ctx).ServeHTTP(w, r)
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
	if e.ResourceOwnerPasswordCredentialsGrantAccessTokenFunc != nil {
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

type RefreshToken struct {
	Claims   jwt.MapClaims
	Scope    string `json:"scope,omitempty"`
	UserID   string `json:"-"`
	Password string `json:"-"`
}

func (e *JWTAuthorizationEndpoint) refreshTokenGrantAccessToken(ctx context.Context, tokenReq *refresh.AccessTokenRequest) (tokenResp *accesstoken.SuccessfulIssueResponse, errText accesstoken.ErrorText) {
	if tokenReq == nil || tokenReq.GrantType != "refresh_token" {
		return nil, accesstoken.ErrorTextUnsupportedGrantType
	}
	jwtToken := tokenReq.RefreshToken
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if e.Key == nil {
			return nil, errors.New("missing key")
		}
		return e.Key.GetVerifiedKey(token)
	})
	if err != nil {
		return nil, accesstoken.ErrorTextUnauthorizedClient
	}

	if !token.Valid {
		return nil, accesstoken.ErrorTextUnauthorizedClient
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		return nil, accesstoken.ErrorTextUnauthorizedClient
	}
	if token_type, ok := claims["token_type"]; !ok || token_type != "refresh_token" {
		return nil, accesstoken.ErrorTextUnauthorizedClient
	}

	if e.RefreshTokenGrantAccessTokenFunc != nil {
		return e.RefreshTokenGrantAccessTokenFunc(ctx, &RefreshToken{
			Claims:   claims,
			Scope:    tokenReq.Scope,
			UserID:   tokenReq.UserID,
			Password: tokenReq.Password,
		})
	}
	// UnImplemented
	return nil, accesstoken.ErrorTextUnsupportedGrantType
}

func (e *JWTAuthorizationEndpoint) authorizate(ctx context.Context, accessTokenType *accesstoken.AccessTokenType) (errText accesstoken.ErrorText) {
	if accessTokenType == nil || accessTokenType.TokenType != "Bearer" {
		return accesstoken.ErrorTextUnauthorizedClient
	}
	jwtToken := accessTokenType.AccessToken
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if e.Key == nil {
			return nil, errors.New("missing key")
		}
		return e.Key.GetVerifiedKey(token)
	})
	if err != nil {
		return accesstoken.ErrorTextUnauthorizedClient
	}
	if !token.Valid {
		return accesstoken.ErrorTextUnauthorizedClient
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		return accesstoken.ErrorTextUnauthorizedClient
	}
	if token_type, ok := claims["token_type"]; !ok || token_type != "access_token" {
		return accesstoken.ErrorTextUnauthorizedClient
	}

	if e.AuthorizateFunc != nil {
		return e.AuthorizateFunc(ctx, claims)
	}

	// UnImplemented
	return accesstoken.ErrorTextUnauthorizedClient
}

// TimeNowFunc provides the current time. You can override it to use another time value.
// This is useful for testing or if your server uses a different time zone than your tokens.
func (e *JWTAuthorizationEndpoint) TimeNow(ctx context.Context) time.Time {
	if e.TimeNowFunc != nil {
		return e.TimeNowFunc(ctx)
	}
	return time.Now()
}
