package basic

import (
	"github.com/searKing/golib/net/http_/auth"
)

// for require auth
func NewBasicAuthenticate(realm string) *auth.Authenticate {
	return &auth.Authenticate{
		Newauth: AuthSchemaBasic,
		Params:  map[string]string{"realm": realm},
	}
}
