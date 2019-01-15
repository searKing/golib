package jwt_

import (
	"context"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/searKing/golib/crypto/auth"
	"net/http"
	"time"
)

const (
	// https://jwt.io/introduction/
	// https://auth0.com/learn/json-web-tokens/
	// Whenever the user wants to access a protected route or resource,
	// the user agent should send the JWT,
	// typically in the Authorization header using the Bearer schema.
	// The content of the header should look like the following:
	// 		Authorization: Bearer <token>
	AuthSchemaJWT               = "Bearer"
	defaultAccessTokenExpireIn  = time.Hour
	defaultRefreshTokenExpireIn = 7 * 24 * time.Hour
	defaultIssuer               = "default-go-jwt"
	defaultSubject              = "default-go-jwt-authorize-server"
	defaultAudience             = "*"
	defaultTimeFormat           = time.RFC3339
)

// Login Handler
type LoginResp struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    string `json:"expires_in,omitempty"`
}

// Refresh Handler
type RefreshResp struct {
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   string `json:"expires_in,omitempty"`
}

// JWTAuth provides a Json-Web-Token authentication implementation. On failure, a 401 HTTP response
// is returned. On success, the wrapped middleware is called, and the userID is made available as
// c.Get("userID").(string).
// Users can get a token by posting a json request to LoginHandler. The token then needs to be passed in
// the Authentication header. Example: Authorization:Bearer XXX_TOKEN_XXX
type JWTAuth struct {
	// Realm name to display to the user. Required.
	// https://tools.ietf.org/html/rfc7235#section-2.2
	Realm string `options:"optional" default:""`

	// Whenever the user wants to access a protected route or resource,
	// the user agent should send the JWT,
	// https://jwt.io/introduction/
	Schema string `options:"optional" default:"Bearer"`

	Scheme *AuthenticationScheme `options:"optional"`

	// Duration that a jwt access-token is valid. Optional, defaults to one hour.
	AccessExpireIn time.Duration `options:"optional" deault:""`
	// Duration that a jwt refresh-token is valid. Optional, defaults to seven days.
	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	RefreshExpireIn time.Duration `options:"optional"`

	// Callback function that should perform the authentication of the user based on userID and
	// password. Must return true on success, false on failure. Required.
	// Option return user id, if so, user id will be stored in Claim Array.
	AuthenticatorFunc func(ctx context.Context, r *http.Request) (clientId string, pass bool) `options:"optional"`

	// Callback function that should perform the authorization of the authenticated user. Called
	// only after an authentication success. Must return true on success, false on failure.
	// Optional, default to success.
	AuthorizatorFunc func(ctx context.Context, claims jwt.MapClaims, w http.ResponseWriter) (pass bool) `options:"optional"`

	// Callback function that will be called during login.
	// Using this function it is possible to add additional payload data to the webtoken.
	// The data is then made available during requests via c.Get("JWT_PAYLOAD").
	// Note that the payload is not encrypted.
	// The attributes mentioned on jwt.io can't be used as keys for the map.
	// Optional, by default no additional data will be set.
	PayloadFunc func(ctx context.Context, clientId string) map[string]interface{} `options:"optional"`

	// User can define own UnauthorizedFunc func.
	UnauthorizedFunc func(ctx context.Context, w http.ResponseWriter, status int) `options:"optional"`

	// TimeNowFunc provides the current time. You can override it to use another time value.
	// This is useful for testing or if your server uses a different time zone than your tokens.
	TimeNowFunc func(ctx context.Context) time.Time `options:"optional"`
}

func NewJWTAuth(alg string, keys ...[]byte) *JWTAuth {
	return &JWTAuth{
		Scheme:          NewAuthenticationScheme(alg, keys...),
		AccessExpireIn:  defaultAccessTokenExpireIn,
		RefreshExpireIn: defaultRefreshTokenExpireIn,
	}
}
func NewJWTAuthFromFile(alg string, keyFiles ...string) *JWTAuth {
	return &JWTAuth{
		Scheme:          NewAuthenticationSchemeFromFile(alg, keyFiles...),
		AccessExpireIn:  defaultAccessTokenExpireIn,
		RefreshExpireIn: defaultRefreshTokenExpireIn,
	}
}

// AuthorizateHandler makes JWTAuth implement the Middleware interface.
// 认证
func (mw *JWTAuth) AuthenticateHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.Scheme == nil {
			mw.Unauthorized(ctx, w, http.StatusInternalServerError)
			return
		}
		if err := mw.Scheme.ReadHTTP(r); err != nil {
			mw.Unauthorized(ctx, w, http.StatusForbidden)
			return
		}
		claims := mw.Scheme.Claims
		var mapClams jwt.MapClaims
		if claims != nil {
			mc, ok := claims.(jwt.MapClaims)
			if ok {
				mapClams = mc
			}
		}

		if !mw.Authorizator(ctx, mapClams, w) {
			mw.Unauthorized(ctx, w, http.StatusForbidden)
			return
		}
	})
}

// LoginHandler can be used by clients to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"access_token": "ACCESS_TOKEN", "refresh_token": "REFRESH_TOKEN", "expires_in": "EXPIRES_IN"}.
func (mw *JWTAuth) LoginHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientId, ok := mw.Authenticator(ctx, r)
		if !ok {
			mw.Unauthorized(ctx, w, http.StatusUnauthorized)
			return
		}
		// Create the token
		claims := jwt.MapClaims{}
		for key, value := range mw.Payload(ctx, clientId) {
			claims[key] = value
		}
		now := mw.TimeNow(ctx)

		accessToken, refreshToken, accessTokenExpireIn, err := mw.generateTokens(now, claims, true)
		if err != nil {
			mw.Unauthorized(ctx, w, http.StatusInternalServerError)
			return
		}
		mw.Scheme.WriteHTTP(w)
		resp := LoginResp{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    accessTokenExpireIn.Format(defaultTimeFormat),
		}
		respBytes, err := json.MarshalIndent(&resp, "", "\t")
		if err != nil {
			mw.Unauthorized(ctx, w, http.StatusInternalServerError)
			return
		}
		w.Write(respBytes)
	})
}

// RefreshHandler can be used to refresh a token. The token still needs to be valid on refresh.
// Shall be put under an endpoint that is using the JWTAuth.
// Reply will be of the form {"access_token": "ACCESS_TOKEN", "expires_in": "EXPIRES_IN"}.
func (mw *JWTAuth) RefreshHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := mw.Scheme.ReadHTTP(r); err != nil {
			mw.Unauthorized(ctx, w, http.StatusForbidden)
			return
		}

		jwt.TimeFunc = time.Now
		var claims jwt.MapClaims
		claims, ok := mw.Scheme.Claims.(jwt.MapClaims)
		if !ok {
			mw.Unauthorized(ctx, w, http.StatusForbidden)
			return
		}

		now := mw.TimeNowFunc(ctx)

		accessToken, _, accessTokenExpireIn, err := mw.generateTokens(now, claims, false)
		if err != nil {
			mw.Unauthorized(ctx, w, http.StatusInternalServerError)
			return
		}
		mw.Scheme.WriteHTTP(w)
		resp := RefreshResp{
			AccessToken: accessToken,
			ExpiresIn:   accessTokenExpireIn.Format(defaultTimeFormat),
		}
		respBytes, err := json.MarshalIndent(&resp, "", "\t")
		if err != nil {
			mw.Unauthorized(ctx, w, http.StatusInternalServerError)
			return
		}
		w.Write(respBytes)
		return
	})
}

// Callback function that should perform the authentication of the user based on userID and
// password. Must return true on success, false on failure. Required.
// Option return user id, if so, user id will be stored in Claim Array.
func (mw *JWTAuth) Authenticator(ctx context.Context, r *http.Request) (clientId string, pass bool) {
	if mw.AuthenticatorFunc != nil {
		return mw.AuthenticatorFunc(ctx, r)
	}
	return "", true
}

// Callback function that should perform the authorization of the authenticated user. Called
// only after an authentication success. Must return true on success, false on failure.
// Optional, default to success.
func (mw *JWTAuth) Authorizator(ctx context.Context, claims jwt.MapClaims, w http.ResponseWriter) (pass bool) {
	if mw.AuthorizatorFunc != nil {
		return mw.AuthorizatorFunc(ctx, claims, w)
	}
	return true
}

// Callback function that will be called during login.
// Using this function it is possible to add additional payload data to the webtoken.
// The data is then made available during requests via c.Get("JWT_PAYLOAD").
// Note that the payload is not encrypted.
// The attributes mentioned on jwt.io can't be used as keys for the map.
// Optional, by default no additional data will be set.
func (mw *JWTAuth) Payload(ctx context.Context, clientId string) map[string]interface{} {
	if mw.PayloadFunc != nil {
		return mw.PayloadFunc(ctx, clientId)
	}
	return nil
}

// show 401 UnauthorizedFunc error.
func (mw *JWTAuth) Unauthorized(ctx context.Context, w http.ResponseWriter, statusCode int) {
	jwtAuth := NewJWTAuthenticate(mw.Realm, mw.Schema)
	jwtAuth.WriteHTTPWithStatusCode(w, statusCode)

	if mw.UnauthorizedFunc != nil {
		mw.UnauthorizedFunc(ctx, w, statusCode)
		return
	}

	return
}

// TimeNowFunc provides the current time. You can override it to use another time value.
// This is useful for testing or if your server uses a different time zone than your tokens.
func (mw *JWTAuth) TimeNow(ctx context.Context) time.Time {
	if mw.TimeNowFunc != nil {
		return mw.TimeNowFunc(ctx)
	}
	return time.Now()
}

// generateTokens method that clients can use to get a jwt token pair.
func (mw *JWTAuth) generateTokens(now time.Time, claims jwt.MapClaims, genRefreshToken bool) (accessToken, refreshToken string, accessTokenExpireIn time.Time, err error) {
	accessExpireIn := now.Add(mw.AccessExpireIn)
	refreshExpireIn := now.Add(mw.RefreshExpireIn)
	mw.Scheme.Claims = claims

	if claims[ClaimsIssuer] == "" {
		claims[ClaimsIssuer] = defaultIssuer
	}
	if claims[ClaimsSubject] == "" {
		claims[ClaimsSubject] = defaultSubject
	}
	if claims[ClaimsAudience] == "" {
		claims[ClaimsAudience] = defaultAudience
	}
	claims[ClaimsNotBefore] = now.Add(-1 * time.Second).Unix()
	claims[ClaimsIssuedAt] = now.Unix()

	if genRefreshToken && refreshExpireIn != (time.Time{}) {
		claims[ClaimsExpirationTime] = refreshExpireIn.Unix()
		claims[ClaimsJWTID] = auth.UUID()
		// Refresh Token
		refreshToken, err = mw.Scheme.RefreshToken()
		if err != nil {
			return "", "", time.Now(), err
		}
	}

	// Access Token
	claims[ClaimsExpirationTime] = accessExpireIn.Unix()
	claims[ClaimsJWTID] = auth.UUID()
	accessToken, err = mw.Scheme.RefreshToken()
	if err != nil {
		return "", "", time.Now(), err
	}
	return accessToken, refreshToken, accessExpireIn, nil
}
