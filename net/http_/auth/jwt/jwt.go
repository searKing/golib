package jwt

import (
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

	Scheme AuthenticationScheme `options:"required"`

	// Duration that a jwt access-token is valid. Optional, defaults to one hour.
	AccessExpireIn time.Duration `options:"optional"`
	// Duration that a jwt refresh-token is valid. Optional, defaults to seven days.
	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is MaxRefresh + Timeout.
	// Optional, defaults to 0 meaning not refreshable.
	RefreshExpireIn time.Duration `options:"optional"`

	// Callback function that should perform the authentication of the user based on userID and
	// password. Must return true on success, false on failure. Required.
	// Option return user id, if so, user id will be stored in Claim Array.
	Authenticator func(r *http.Request) (appId string, pass bool) `options:"optional"`

	// Callback function that should perform the authorization of the authenticated user. Called
	// only after an authentication success. Must return true on success, false on failure.
	// Optional, default to success.
	Authorizator func(userID string, w http.ResponseWriter) bool `options:"optional"`

	// Callback function that will be called during login.
	// Using this function it is possible to add additional payload data to the webtoken.
	// The data is then made available during requests via c.Get("JWT_PAYLOAD").
	// Note that the payload is not encrypted.
	// The attributes mentioned on jwt.io can't be used as keys for the map.
	// Optional, by default no additional data will be set.
	PayloadFunc func(appId string) map[string]interface{} `options:"optional"`

	// User can define own Unauthorized func.
	Unauthorized func(w http.ResponseWriter, status int) `options:"optional"`

	// Set the identity handler function
	IdentityHandler func(claims jwt.Claims) string `options:"optional"`

	// TimeNowFunc provides the current time. You can override it to use another time value.
	// This is useful for testing or if your server uses a different time zone than your tokens.
	TimeNowFunc func() time.Time `options:"optional"`
}

func (mw *JWTAuth) usingPublicKeyAlgo() bool {
	return !mw.Scheme.Key.IsSymmetricKey()
}

// LoadDefault initialize jwt configs.
func (mw *JWTAuth) LoadDefault() error {

	if mw.AccessExpireIn == 0 {
		mw.AccessExpireIn = defaultAccessTokenExpireIn
	}

	if mw.RefreshExpireIn == 0 {
		mw.RefreshExpireIn = defaultRefreshTokenExpireIn
	}

	if mw.TimeNowFunc == nil {
		mw.TimeNowFunc = time.Now
	}

	if mw.Authorizator == nil {
		mw.Authorizator = func(userID string, w http.ResponseWriter) bool {
			return true
		}
	}

	if mw.Unauthorized == nil {
		mw.Unauthorized = func(w http.ResponseWriter, statusCode int) {
		}
	}

	if mw.IdentityHandler == nil {
		mw.IdentityHandler = func(claims jwt.Claims) string {
			return ""
		}
	}
	return nil
}

// AuthenticateHandler makes JWTAuth implement the Middleware interface.
func (mw *JWTAuth) AuthenticateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := mw.LoadDefault(); err != nil {
			mw.unauthorized(w, http.StatusInternalServerError)
			return
		}
		mw.Scheme.ReadHTTP(r)
		claims := mw.Scheme.Claims

		id := mw.IdentityHandler(claims)

		if !mw.Authorizator(id, w) {
			mw.unauthorized(w, http.StatusForbidden)
			return
		}
	})
}

// LoginHandler can be used by clients to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"access_token": "ACCESS_TOKEN", "refresh_token": "REFRESH_TOKEN", "expires_in": "EXPIRES_IN"}.
func (mw *JWTAuth) LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Initial middleware default setting.
		if err := mw.LoadDefault(); err != nil {
			mw.unauthorized(w, http.StatusInternalServerError)
			return
		}

		if mw.Authenticator == nil {
			mw.unauthorized(w, http.StatusInternalServerError)
			return
		}
		appId, ok := mw.Authenticator(r)
		if !ok {
			mw.unauthorized(w, http.StatusUnauthorized)
			return
		}
		// Create the token
		claims := jwt.MapClaims{}
		if mw.PayloadFunc != nil {
			for key, value := range mw.PayloadFunc(appId) {
				claims[key] = value
			}
		}

		accessToken, refreshToken, accessTokenExpireIn, err := mw.GenerateTokens(claims, true)
		if err != nil {
			mw.unauthorized(w, http.StatusInternalServerError)
			return
		}
		mw.Scheme.WriteHTTP(w)
		resp := LoginResp{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    accessTokenExpireIn.Format(defaultTimeFormat),
		}
		respBytes, err := json.Marshal(&resp)
		if err == nil {
			mw.unauthorized(w, http.StatusInternalServerError)
			return
		}
		w.Write(respBytes)
	})
}

// RefreshHandler can be used to refresh a token. The token still needs to be valid on refresh.
// Shall be put under an endpoint that is using the JWTAuth.
// Reply will be of the form {"access_token": "ACCESS_TOKEN", "expires_in": "EXPIRES_IN"}.
func (mw *JWTAuth) RefreshHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Initial middleware default setting.
		if err := mw.LoadDefault(); err != nil {
			mw.unauthorized(w, http.StatusInternalServerError)
			return
		}
		if err := mw.Scheme.ReadHTTP(r); err != nil {
			mw.unauthorized(w, http.StatusUnauthorized)
			return
		}

		jwt.TimeFunc = time.Now
		var claims jwt.MapClaims
		claims, ok := mw.Scheme.Claims.(jwt.MapClaims)
		if !ok {
			mw.unauthorized(w, http.StatusUnauthorized)
			return
		}

		accessToken, _, accessTokenExpireIn, err := mw.GenerateTokens(claims, false)
		if err != nil {
			mw.unauthorized(w, http.StatusInternalServerError)
			return
		}
		mw.Scheme.WriteHTTP(w)
		resp := RefreshResp{
			AccessToken: accessToken,
			ExpiresIn:   accessTokenExpireIn.Format(defaultTimeFormat),
		}
		respBytes, err := json.Marshal(&resp)
		if err == nil {
			mw.unauthorized(w, http.StatusInternalServerError)
			return
		}
		w.Write(respBytes)
		return
	})
}

// GenerateTokens method that clients can use to get a jwt token pair.
func (mw *JWTAuth) GenerateTokens(claims jwt.MapClaims, genRefreshToken bool) (accessToken, refreshToken string, accessTokenExpireIn time.Time, err error) {
	now := mw.TimeNowFunc()
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

	if genRefreshToken && refreshExpireIn == (time.Time{}) {
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

// show 401 unauthorized error.
func (mw *JWTAuth) unauthorized(w http.ResponseWriter, statusCode int) {
	auth := NewJWTAuthenticate(mw.Realm, mw.Schema)
	auth.WriteHTTPWithStatusCode(w, statusCode)
	mw.Unauthorized(w, statusCode)
	return
}
