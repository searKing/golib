package auth

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// https://tools.ietf.org/html/rfc7235#section-4.1
// Newauth realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
type Authenticate struct {
	Newauth string
	Params  AuthenticationParameters
}

// Newauth realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
func (a *Authenticate) Write(w io.Writer) error {
	fmt.Fprintf(w, `%s`, a.Newauth)
	if len(a.Params) > 0 {
		fmt.Fprintf(w, ` `)
		return a.Params.Write(w)
	}
	return nil
}

func (a *Authenticate) String() string {
	b := bytes.NewBuffer([]byte{})
	bw := bufio.NewWriter(b)
	if err := a.Write(bw); err != nil {
		return ""
	}
	return b.String()
}

// WWW-Authenticate: Newauth realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
func (a *Authenticate) WriteHTTP(w http.ResponseWriter) {
	a.WriteHTTPWithStatusCode(w, http.StatusUnauthorized)
}

func (a *Authenticate) WriteHTTPWithStatusCode(w http.ResponseWriter, statusCode int) {
	w.Header().Set(HeaderFieldAuthenticate, a.String())
	w.WriteHeader(statusCode)
	w.Write([]byte(http.StatusText(statusCode)))
}
