package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// https://tools.ietf.org/html/rfc7235#section-4.1
// realm="apps", type=1, title="Login to \"apps\"", Basic realm="simple"
//
//
//      Note: The challenge grammar production uses the list syntax as
//      well.  Therefore, a sequence of comma, whitespace, and comma can
//      be considered either as applying to the preceding challenge, or to
//      be an empty entry in the list of challenges.  In practice, this
//      ambiguity does not affect the semantics of the header field value
//      and thus is harmless.
type AuthenticationParameters map[string]string

func (params AuthenticationParameters) Write(w io.Writer) error {
	first := true
	for key, value := range params {
		val, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if !first {
			w.Write([]byte(`, `))
		}
		w.Write([]byte(fmt.Sprintf(`%s=%s`, key, string(val))))
	}
	return nil
}
func (params AuthenticationParameters) String() string {
	b := bytes.NewBuffer([]byte{})
	bw := bufio.NewWriter(b)
	if err := params.Write(bw); err != nil {
		return ""
	}
	return b.String()
}
