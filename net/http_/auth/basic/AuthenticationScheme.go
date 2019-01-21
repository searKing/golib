package basic

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/searKing/golib/net/http_"
	"github.com/searKing/golib/net/http_/auth/internal"
	"io"
	"net/http"
	"strings"
)

// basic-credentials = "Basic" SP basic-cookie
type AuthenticationScheme struct {
	UserID   string
	Password string
}

func NewAuthenticationScheme() *AuthenticationScheme {
	return &AuthenticationScheme{}
}

// basic-credentials = "Basic" SP basic-cookie
func (a *AuthenticationScheme) ReadString(basicCredentials string) {
	// Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==

	// basic-credentials = "Basic" SP basic-cookie
	s := strings.SplitN(basicCredentials, " ", 2)
	if len(s) != 2 || s[0] != AuthSchemaBasic {
		return
	}
	basicCookie := s[1]

	// basic-cookie      = <base64 [5] encoding of userid-password,
	//                   except not limited to 76 char/line>
	useridPassword, err := base64.StdEncoding.DecodeString(basicCookie)
	if err != nil {
		return
	}

	// userid-password   = [ token ] ":" *TEXT
	pair := strings.SplitN(string(useridPassword), ":", 2)
	if len(pair) != 2 {
		return
	}

	a.UserID = pair[0]
	a.Password = pair[1]
	return
}

// ParseBasicAuthenticationScheme parses userID and passwordf frome http's Header
// Reference : https://tools.ietf.org/html/rfc1945#section-11 11.1
func (a *AuthenticationScheme) ReadHTTP(r *http.Request) {
	// Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==
	// Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==
	// QWxhZGRpbjpvcGVuIHNlc2FtZQ==
	a.ReadString(internal.ParseAuthenticationCredentials(r))
}

func (a *AuthenticationScheme) Write(w io.Writer) error {
	useridPassword := fmt.Sprintf(`%s:%s`, a.UserID, a.Password)
	basicCookie := base64.StdEncoding.EncodeToString([]byte(useridPassword))

	fmt.Fprintf(w, `%s %s`, AuthSchemaBasic, basicCookie)
	return nil
}

func (a *AuthenticationScheme) String() string {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	if err := a.Write(bw); err != nil {
		return ""
	}
	bw.Flush()
	return buf.String()
}

// Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==
func (a *AuthenticationScheme) WriteHTTP(w http.ResponseWriter) {
	w.Header().Set(http_.HeaderFieldAuthorization, a.String())
}
