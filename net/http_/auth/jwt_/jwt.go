package jwt_

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/searKing/golib/crypto/auth"
	"github.com/searKing/golib/encoding/json_"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
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
// Users can get a token by posting a json request to AuthorizationEndointHandler. The token then needs to be passed in
// the Authentication header. Example: Authorization:Bearer XXX_TOKEN_XXX
type JWTAuth struct {
	// Realm name to display to the user. Required.
	// https://tools.ietf.org/html/rfc7235#section-2.2
	Realm string `options:"optional" default:""`

	// https://jwt.io/introduction/
	// https://auth0.com/learn/json-web-tokens/
	// Whenever the user wants to access a protected route or resource,
	// the user agent should send the JWT,
	// typically in the Authorization header using the Bearer schema.
	// The content of the header should look like the following:
	// 		Authorization: Bearer <token>
	TokenSchema string `options:"optional" default:"Bearer"`

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
	AuthenticatorFunc func(ctx context.Context, password *ClientPassword) (pass bool) `options:"optional"`

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
	JWTExtraFunc func(ctx context.Context, clientId string) map[string]interface{} `options:"optional"`

	// User can define own UnauthorizedFunc func.
	UnauthorizedFunc func(ctx context.Context, w http.ResponseWriter, status int) `options:"optional"`

	// TimeNowFunc provides the current time. You can override it to use another time value.
	// This is useful for testing or if your server uses a different time zone than your tokens.
	TimeNowFunc func(ctx context.Context) time.Time `options:"optional"`
}

func NewJWTAuthFromRandom(alg string) (*JWTAuth, error) {
	scheme, err := NewAuthenticationSchemeFromRandom(alg)
	if err != nil {
		return nil, err
	}
	return &JWTAuth{
		Scheme:          scheme,
		AccessExpireIn:  defaultAccessTokenExpireIn,
		RefreshExpireIn: defaultRefreshTokenExpireIn,
	}, err
}

func NewJWTAuth(alg string, privateKey []byte, publicKey []byte, password ...string) (*JWTAuth, error) {
	scheme, err := NewAuthenticationScheme(alg, privateKey, publicKey, password...)
	if err != nil {
		return nil, err
	}
	return &JWTAuth{
		Scheme:          scheme,
		AccessExpireIn:  defaultAccessTokenExpireIn,
		RefreshExpireIn: defaultRefreshTokenExpireIn,
	}, err
}

func NewJWTAuthFromFile(alg string, privateKeyFile string, publicKeyFile string, password ...string) (*JWTAuth, error) {
	scheme, err := NewAuthenticationSchemeFromFile(alg, privateKeyFile, publicKeyFile, password...)
	if err != nil {
		return nil, err
	}
	return &JWTAuth{
		Scheme:          scheme,
		AccessExpireIn:  defaultAccessTokenExpireIn,
		RefreshExpireIn: defaultRefreshTokenExpireIn,
	}, err
}

// Type returns t.TokenType if non-empty, else "Bearer".
// https://jwt.io/introduction/
// https://auth0.com/learn/json-web-tokens/
// Whenever the user wants to access a protected route or resource,
// the user agent should send the JWT,
// typically in the Authorization header using the Bearer schema.
// The content of the header should look like the following:
// 		Authorization: Bearer <token>
func (mw *JWTAuth) Schema() string {
	if strings.EqualFold(mw.TokenSchema, "bearer") {
		return "Bearer"
	}
	if strings.EqualFold(mw.TokenSchema, "mac") {
		return "MAC"
	}
	if strings.EqualFold(mw.TokenSchema, "basic") {
		return "Basic"
	}
	if mw.TokenSchema != "" {
		return mw.TokenSchema
	}
	return "Bearer"
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
		if token_type, ok := mapClams["token_type"]; !ok || token_type != "access_token" {
			mw.Unauthorized(ctx, w, http.StatusForbidden)
			return
		}

		if !mw.Authorizator(ctx, mapClams, w) {
			mw.Unauthorized(ctx, w, http.StatusForbidden)
			return
		}
	})
}

// rfc6749 2.3.1
type ClientPassword struct {
	// The client identifier issued to the client during
	// the registration process
	ClientId string `json:"client_id" options:"required"`
	// The client secret.  The client MAY omit the
	// parameter if the client secret is an empty string.
	ClientSecret string `json:"client_secret,omitempty" options:"required"`
}

// rfc6749 2.3.1
func RetrieveClientPassword(ctx context.Context, r *http.Request) (*ClientPassword, error) {
	var body []byte
	defer func() {
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
	}()

	// rfc6749 2.3.1
	//  The authorization server MUST support the HTTP Basic
	//   authentication scheme for authenticating clients that were issued a
	//   client password.
	if clientId, clientSecret, ok := r.BasicAuth(); ok {
		return &ClientPassword{
			ClientId:     clientId,
			ClientSecret: clientSecret,
		}, nil
	}
	//Alternatively, the authorization server MAY support including the
	//client credentials in the request-body using the following
	//parameters
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	content, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch content {
	case "application/x-www-form-urlencoded", "text/plain":
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		return &ClientPassword{
			ClientId:     vals.Get("client_id"),
			ClientSecret: vals.Get("client_secret"),
		}, nil
	case "application/json":
		var cp ClientPassword
		if err = json.Unmarshal(body, &cp); err != nil {

			return nil, err
		}
		return &cp, nil
	default:
		vars := r.URL.Query()
		clientIds, ok := vars["client_id"]
		if !ok || len(clientIds) == 0 {
			return nil, errors.New("missing client_id")
		}
		clientSecrets, ok := vars["client_secret"]
		if !ok || len(clientSecrets) == 0 {
			return nil, errors.New("missing client_secret")

		}
		return &ClientPassword{
			ClientId:     clientIds[0],
			ClientSecret: clientSecrets[0],
		}, nil

	}
}
func (mw *JWTAuth) AccessTokenHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

// AuthorizationEndointHandler can be used by clients to interact with the resource
// owner and obtain an authorization grant.
// MUST support "GET" and MAY support "POST" method as well.
func (mw *JWTAuth) AuthorizationEndointHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodGet {

		}
		clientPassword, err := RetrieveClientPassword(ctx, r)
		if err != nil {
			mw.Unauthorized(ctx, w, http.StatusUnauthorized)
			return
		}

		// The authorization server MUST first verify the identity of the resource owner.
		if !mw.Authenticator(ctx, clientPassword) {
			mw.Unauthorized(ctx, w, http.StatusUnauthorized)
			return
		}

		// Create the token
		claims := jwt.MapClaims{}
		for key, value := range mw.JWTExtra(ctx, clientPassword.ClientId) {
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
		if token_type, ok := claims["token_type"]; !ok || token_type != "refresh_token" {
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
func (mw *JWTAuth) Authenticator(ctx context.Context, password *ClientPassword) (pass bool) {
	if mw.AuthenticatorFunc != nil {
		return mw.AuthenticatorFunc(ctx, password)
	}
	return true
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
func (mw *JWTAuth) JWTExtra(ctx context.Context, clientId string) map[string]interface{} {
	if mw.JWTExtraFunc != nil {
		return mw.JWTExtraFunc(ctx, clientId)
	}
	return nil
}

// show 401 UnauthorizedFunc error.
func (mw *JWTAuth) Unauthorized(ctx context.Context, w http.ResponseWriter, statusCode int) {
	jwtAuth := NewJWTAuthenticate(mw.Realm, mw.Schema())
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

// generateTokens method that clients can use to get a jwt token pair.
func (mw *JWTAuth) generateTokens(now time.Time, extraClaims jwt.MapClaims, genRefreshToken bool) (accessToken, refreshToken string, accessTokenExpireIn time.Time, err error) {
	claims := Claims{}
	claims.PrivateClaims = map[string]interface{}{}
	// append extraClaims as privateClaims
	for k, v := range extraClaims {
		claims.PrivateClaims[k] = v
	}

	accessExpireIn := now.Add(mw.AccessExpireIn)
	refreshExpireIn := now.Add(mw.RefreshExpireIn)

	if mw.AccessExpireIn > 0 {
		claims.ExpiresAt = accessExpireIn.Unix()
	}
	claims.NotBefore = now.Add(-1 * time.Second).Unix()
	claims.IssuedAt = now.Unix()
	if genRefreshToken && mw.RefreshExpireIn > 0 {
		claims.ExpiresAt = refreshExpireIn.Unix()
		claims.Id = auth.UUID()
		claims.PrivateClaims["token_type"] = "refresh_token"
		// Refresh Token
		mw.Scheme.Claims, err = claims.ToMapClaims()
		if err != nil {
			return "", "", now, err
		}
		refreshToken, err = mw.Scheme.RefreshToken()
		if err != nil {
			return "", "", now, err
		}
	}

	// Access Token
	claims.ExpiresAt = accessExpireIn.Unix()
	claims.Id = auth.UUID()
	claims.PrivateClaims["token_type"] = "access_token"

	mw.Scheme.Claims, err = claims.ToMapClaims()
	if err != nil {
		return "", "", now, err
	}
	accessToken, err = mw.Scheme.RefreshToken()
	if err != nil {
		return "", "", now, err
	}
	return accessToken, refreshToken, accessExpireIn, nil
}
