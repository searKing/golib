package jwt

import (
	"github.com/searKing/golib/net/http_/auth"
	"strings"
)

// for require auth
// JWT realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
func NewJWTAuthenticate(realm string, schema string) *auth.Authenticate {
	var params map[string]string
	if strings.TrimSpace(realm) != "" {
		params = map[string]string{"realm": realm}
	}
	if strings.TrimSpace(schema) == "" {
		schema = AuthSchemaJWT
	}
	return &auth.Authenticate{
		Newauth: schema,
		Params:  params,
	}
}
