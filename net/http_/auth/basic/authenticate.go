package basic

import (
	"github.com/searKing/golib/net/http_/auth/internal"
)

// for require auth
func NewBasicAuthenticate(realm string) *internal.Authenticate {
	return &internal.Authenticate{
		Newauth: AuthSchemaBasic,
		Params:  map[string]string{"realm": realm},
	}
}
