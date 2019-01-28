package access

import (
	"net/http"
	"strings"
)

// rfc7235 4.2
// Authorization = credentials
type Authorization struct {
	Credentials string
}

func ParseAuthorization(auth string) (*Authorization, error) {
	return &Authorization{
		Credentials: auth,
	}, nil
}

func (a *Authorization) String() string {
	var b strings.Builder
	b.WriteString(a.Credentials)
	return b.String()
}

func (a *Authorization) SetAuthHeader(w http.ResponseWriter) {
	w.Header().Add("Authorization", a.String())
}
