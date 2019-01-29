package accesstoken

import (
	"context"
	"github.com/searKing/golib/net/http_/oauth2/access/bearer"
	"net/http"
)

// rfc6749 7
// rfc6750 2
func RetrieveAccessTokenType(ctx context.Context, r *http.Request) (*AccessTokenType, error) {
	bearerCredentials, err := bearer.ParseCredentialsFromRequest(r)
	if err != nil {
		return nil, err
	}
	return &AccessTokenType{
		TokenType:   "Bearer",
		AccessToken: bearerCredentials.B64token,
	}, nil
}
