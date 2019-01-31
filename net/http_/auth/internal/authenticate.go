package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/searKing/golib/net/http_"
	"io"
	"net/http"
)

// rfc7235 4.1
// Newauth realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
type Authenticate struct {
	Newauth string
	Params  AuthenticationParameters
}

// Newauth realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
func (a *Authenticate) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, `%s`, a.Newauth)
	if err != nil {
		return err
	}
	if len(a.Params) > 0 {
		fmt.Fprintf(w, ` `)
		return a.Params.Write(w)
	}
	return nil
}

func (a *Authenticate) String() string {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	if err := a.Write(bw); err != nil {
		return ""
	}
	bw.Flush()
	return buf.String()
}

// WWW-Authenticate: Newauth realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
func (a *Authenticate) WriteHTTP(w http.ResponseWriter) {
	a.WriteHTTPWithStatusCode(w, http.StatusUnauthorized)
}

func (a *Authenticate) WriteHTTPWithStatusCode(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set(http_.HeaderFieldAuthenticate, a.String())
}
