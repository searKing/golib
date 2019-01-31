package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/searKing/golib/crypto/auth"
	"github.com/searKing/golib/encoding/json_"
	"github.com/searKing/golib/net/http_/auth/jwt_"
	"github.com/searKing/golib/net/http_/oauth2/grant/accesstoken"
	"github.com/searKing/golib/net/http_/oauth2/grant/authorize"
	"github.com/searKing/golib/net/http_/oauth2/grant/implict"
	"net/http"
	"time"
)

type JWTAuthorizeAccessTokenResponse struct {
	CustomClaims jwt.MapClaims `json:"custom_claims"`
	// Duration that a jwt access-token is valid. Optional, defaults to one hour.
	AccessExpireIn time.Duration `json:"access_expire_in,omitempty"`
	// Duration that a jwt refresh-token is valid. Optional, defaults to seven days.
	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	RefreshExpireIn time.Duration `json:"refresh_expire_in,omitempty"`
}
type JWTAccessTokenResponse struct {
	CustomClaims jwt.MapClaims `json:"custom_claims"`
	// Duration that a jwt access-token is valid. Optional, defaults to one hour.
	AccessExpireIn time.Duration `json:"access_expire_in,omitempty"`
	// Duration that a jwt refresh-token is valid. Optional, defaults to seven days.
	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	RefreshExpireIn time.Duration `json:"refresh_expire_in,omitempty"`
	Scope           string        `json:"scope,omitempty"`
}

type JWTAuthorizationCodeGrantAuthorizationResult struct {
	Code string `json:"code"`
}

type JWTImplicitGrantAuthorizationResult struct {
	CustomClaims jwt.MapClaims `json:"custom_claims"`
	// Duration that a jwt access-token is valid. Optional, defaults to one hour.
	AccessExpireIn time.Duration `json:"access_expire_in,omitempty"`
	// Duration that a jwt refresh-token is valid. Optional, defaults to seven days.
	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	RefreshExpireIn time.Duration `json:"refresh_expire_in,omitempty"`
	Scope           string        `json:"scope,omitempty"`
}

type JWTRefreshTokenRequest struct {
	Claims   jwt.MapClaims
	Scope    string `json:"scope,omitempty"`
	UserID   string `json:"-"`
	Password string `json:"-"`
}

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

	AuthorizationCodeGrantAuthorizationFunc func(ctx context.Context, authReq *AuthorizationRequest) (res *AuthorizeAuthorizationResult, err authorize.ErrorText)
	ImplicitGrantAuthorizationFunc          func(ctx context.Context, authReq *AuthorizationRequest) (res *JWTImplicitGrantAuthorizationResult, err implict.ErrorText)

	AuthorizationCodeGrantAccessTokenFunc                func(ctx context.Context, tokenReq *AuthorizeAccessTokenRequest) (tokenResp *JWTAuthorizeAccessTokenResponse, err accesstoken.ErrorText)
	ResourceOwnerPasswordCredentialsGrantAccessTokenFunc func(ctx context.Context, tokenReq *ResourceAccessTokenRequest) (tokenResp *JWTAccessTokenResponse, err accesstoken.ErrorText)
	ClientCredentialsGrantAccessTokenFunc                func(ctx context.Context, tokenReq *ClientAccessTokenRequest) (tokenResp *JWTAccessTokenResponse, err accesstoken.ErrorText)
	RefreshTokenGrantAccessTokenFunc                     func(ctx context.Context, tokenReq *JWTRefreshTokenRequest) (tokenResp *JWTAccessTokenResponse, err accesstoken.ErrorText)

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

func (e *JWTAuthorizationEndpoint) authorizationCodeGrantAuthorization(ctx context.Context, authReq *AuthorizationRequest) (res *AuthorizeAuthorizationResult, err authorize.ErrorText) {
	if e.AuthorizationCodeGrantAuthorizationFunc != nil {
		return e.AuthorizationCodeGrantAuthorizationFunc(ctx, authReq)
	}
	// UnImplemented
	return nil, authorize.ErrorTextUnsupportedResponseType
}

func (e *JWTAuthorizationEndpoint) implicitGrantAuthorization(ctx context.Context, authReq *AuthorizationRequest) (res *ImplicitAuthorizationResult, errText implict.ErrorText) {
	if e.ImplicitGrantAuthorizationFunc == nil {
		// UnImplemented
		return nil, implict.ErrorTextUnsupportedResponseType
	}

	jwtAuthResp, errText := e.ImplicitGrantAuthorizationFunc(ctx, authReq)
	if errText != "" {
		return nil, errText
	}
	gen := &JWTGenerator{
		Key:             e.Key,
		AccessExpireIn:  jwtAuthResp.AccessExpireIn,
		RefreshExpireIn: jwtAuthResp.RefreshExpireIn,
	}
	accessToken, _, accessTokenExpireIn, err := gen.GenerateTokens(e.TimeNow(ctx), jwtAuthResp.CustomClaims, false)
	if err != nil {
		return nil, implict.ErrorTextServerError
	}

	// https://jwt.io/introduction/
	// Whenever the user wants to access a protected route or resource,
	// the user agent should send the JWT,
	// typically in the Authorization header using the Bearer schema.
	// The content of the header should look like the following:
	//
	//	Authorization: Bearer <token>
	return &ImplicitAuthorizationResult{
		AccessToken: accessToken,
		TokenType:   "bearer",
		ExpiresIn:   int64(accessTokenExpireIn.Seconds()),
		Scope:       jwtAuthResp.Scope,
	}, ""
}
func (e *JWTAuthorizationEndpoint) authorizationCodeGrantAccessToken(ctx context.Context, tokenReq *AuthorizeAccessTokenRequest) (tokenResp *AuthorizeAccessTokenResponse, errText accesstoken.ErrorText) {
	// UnImplemented
	if e.AuthorizationCodeGrantAccessTokenFunc == nil {
		// UnImplemented
		return nil, accesstoken.ErrorTextUnsupportedGrantType
	}

	jwtAuthResp, errText := e.AuthorizationCodeGrantAccessTokenFunc(ctx, tokenReq)
	if errText != "" {
		return nil, errText
	}
	gen := &JWTGenerator{
		Key:             e.Key,
		AccessExpireIn:  jwtAuthResp.AccessExpireIn,
		RefreshExpireIn: jwtAuthResp.RefreshExpireIn,
	}
	accessToken, refreshToken, accessTokenExpireIn, err := gen.GenerateTokens(e.TimeNow(ctx), jwtAuthResp.CustomClaims, true)
	if err != nil {
		return nil, accesstoken.ErrorTextUnauthorizedClient
	}
	// https://jwt.io/introduction/
	// Whenever the user wants to access a protected route or resource,
	// the user agent should send the JWT,
	// typically in the Authorization header using the Bearer schema.
	// The content of the header should look like the following:
	//
	//	Authorization: Bearer <token>
	return &AuthorizeAccessTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		ExpiresIn:    int64(accessTokenExpireIn.Seconds()),
		RefreshToken: refreshToken,
	}, ""
}
func (e *JWTAuthorizationEndpoint) resourceOwnerPasswordCredentialsGrantAccessToken(ctx context.Context, tokenReq *ResourceAccessTokenRequest) (tokenResp *AccessTokenResponse, errText accesstoken.ErrorText) {
	if e.ResourceOwnerPasswordCredentialsGrantAccessTokenFunc == nil {
		// UnImplemented
		return nil, accesstoken.ErrorTextUnsupportedGrantType
	}
	jwtTokenResp, errText := e.ResourceOwnerPasswordCredentialsGrantAccessTokenFunc(ctx, tokenReq)
	if errText != "" {
		return nil, errText
	}
	gen := &JWTGenerator{
		Key:             e.Key,
		AccessExpireIn:  jwtTokenResp.AccessExpireIn,
		RefreshExpireIn: jwtTokenResp.RefreshExpireIn,
	}
	accessToken, refreshToken, accessTokenExpireIn, err := gen.GenerateTokens(e.TimeNow(ctx), jwtTokenResp.CustomClaims, true)
	if err != nil {
		return nil, accesstoken.ErrorTextUnauthorizedClient
	}
	return &AccessTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		ExpiresIn:    int64(accessTokenExpireIn.Seconds()),
		RefreshToken: refreshToken,
		Scope:        jwtTokenResp.Scope,
	}, ""
}
func (e *JWTAuthorizationEndpoint) clientCredentialsGrantAccessToken(ctx context.Context, tokenReq *ClientAccessTokenRequest) (tokenResp *AccessTokenResponse, errText accesstoken.ErrorText) {
	if e.ClientCredentialsGrantAccessTokenFunc != nil {
		// UnImplemented
		return nil, accesstoken.ErrorTextUnsupportedGrantType
	}
	jwtTokenResp, errText := e.ClientCredentialsGrantAccessTokenFunc(ctx, tokenReq)
	if errText != "" {
		return nil, errText
	}
	gen := &JWTGenerator{
		Key:             e.Key,
		AccessExpireIn:  jwtTokenResp.AccessExpireIn,
		RefreshExpireIn: jwtTokenResp.RefreshExpireIn,
	}
	accessToken, refreshToken, accessTokenExpireIn, err := gen.GenerateTokens(e.TimeNow(ctx), jwtTokenResp.CustomClaims, true)
	if err != nil {
		return nil, accesstoken.ErrorTextUnauthorizedClient
	}
	return &AccessTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		ExpiresIn:    int64(accessTokenExpireIn.Seconds()),
		RefreshToken: refreshToken,
		Scope:        jwtTokenResp.Scope,
	}, ""
}

func (e *JWTAuthorizationEndpoint) refreshTokenGrantAccessToken(ctx context.Context, tokenReq *RefreshAccessTokenRequest) (tokenResp *AccessTokenResponse, errText accesstoken.ErrorText) {
	if e.RefreshTokenGrantAccessTokenFunc == nil {
		// UnImplemented
		return nil, accesstoken.ErrorTextUnsupportedGrantType
	}

	if tokenReq == nil {
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

	jwtTokenResp, errText := e.RefreshTokenGrantAccessTokenFunc(ctx, &JWTRefreshTokenRequest{
		Claims:   claims,
		Scope:    tokenReq.Scope,
		UserID:   tokenReq.UserID,
		Password: tokenReq.Password,
	})
	if errText != "" {
		return nil, errText
	}
	gen := &JWTGenerator{
		Key:             e.Key,
		AccessExpireIn:  jwtTokenResp.AccessExpireIn,
		RefreshExpireIn: jwtTokenResp.RefreshExpireIn,
	}
	accessToken, _, accessTokenExpireIn, err := gen.GenerateTokens(e.TimeNow(ctx), jwtTokenResp.CustomClaims, false)
	if err != nil {
		return nil, accesstoken.ErrorTextUnauthorizedClient
	}
	return &AccessTokenResponse{
		AccessToken: accessToken,
		TokenType:   "bearer",
		ExpiresIn:   int64(accessTokenExpireIn.Seconds()),
		Scope:       jwtTokenResp.Scope,
	}, ""
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

type Claims struct {
	// See http://tools.ietf.org/html/draft-jones-json-web-token-10#section-4.1
	jwt.StandardClaims
	// See http://tools.ietf.org/html/draft-jones-json-web-token-10#section-4.2
	PublicClaims jwt.MapClaims
	// See http://tools.ietf.org/html/draft-jones-json-web-token-10#section-4.3
	PrivateClaims jwt.MapClaims
}

func (c *Claims) MarshalJSON() ([]byte, error) {
	var vs []interface{}

	if len(c.PublicClaims) > 0 {
		vs = append(vs, c.PublicClaims)
	}

	if len(c.PrivateClaims) > 0 {
		vs = append(vs, c.PrivateClaims)
	}
	return json_.MarshalConcat(c.StandardClaims, vs...)
}

func (c *Claims) UnmarshalJSON(data []byte) error {
	return json_.UnmarshalConcat(data, &c.StandardClaims, &c.PublicClaims, &c.PrivateClaims)
}

func (c *Claims) ToMapClaims() (mapClaims jwt.MapClaims, err error) {
	buf, err := c.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &mapClaims); err != nil {
		return nil, err
	}
	return mapClaims, err
}

type JWTGenerator struct {
	Key *jwt_.AuthKey
	// Duration that a jwt access-token is valid. Optional, defaults to one hour.
	AccessExpireIn time.Duration
	// Duration that a jwt refresh-token is valid. Optional, defaults to seven days.
	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	RefreshExpireIn time.Duration
}

// generateTokens method that clients can use to get a jwt token pair.
func (g *JWTGenerator) GenerateTokens(now time.Time, extraClaims jwt.MapClaims, genRefreshToken bool) (accessToken, refreshToken string, accessTokenExpireIn time.Duration, err error) {
	claims := Claims{}
	claims.PrivateClaims = map[string]interface{}{}
	// append extraClaims as privateClaims
	for k, v := range extraClaims {
		claims.PrivateClaims[k] = v
	}

	accessExpireIn := now.Add(g.AccessExpireIn)
	refreshExpireIn := now.Add(g.RefreshExpireIn)

	if g.AccessExpireIn > 0 {
		claims.ExpiresAt = accessExpireIn.Unix()
	}
	claims.NotBefore = now.Add(-1 * time.Second).Unix()
	claims.IssuedAt = now.Unix()
	if genRefreshToken && g.RefreshExpireIn > 0 {
		claims.ExpiresAt = refreshExpireIn.Unix()
		claims.Id = auth.UUID()
		claims.PrivateClaims["token_type"] = "refresh_token"
		// Refresh Token
		mapClaims, err := claims.ToMapClaims()
		if err != nil {
			return "", "", 0, err
		}
		refreshToken, err = g.refreshToken(mapClaims)
		if err != nil {
			return "", "", 0, err
		}
	}

	// Access Token
	claims.ExpiresAt = accessExpireIn.Unix()
	claims.Id = auth.UUID()
	claims.PrivateClaims["token_type"] = "access_token"

	mapClaims, err := claims.ToMapClaims()
	if err != nil {
		return "", "", 0, err
	}
	accessToken, err = g.refreshToken(mapClaims)
	if err != nil {
		return "", "", 0, err
	}
	return accessToken, refreshToken, g.AccessExpireIn, nil
}
func (g *JWTGenerator) refreshToken(claims jwt.Claims) (token string, err error) {
	jwtToken, err := g.signedString(claims)
	if err != nil {
		return "", err
	}
	return jwtToken, nil
}
func (g *JWTGenerator) signedString(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(g.Key.GetSignedMethod(), claims)
	key, err := g.Key.GetSignedKey(nil)
	if err != nil {
		return "", err
	}
	return token.SignedString(key)
}
