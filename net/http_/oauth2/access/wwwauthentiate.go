package access

import (
	"net/http"
	"strings"
)

// rfc7235 4.1
// WWW-Authenticate = 1#challenge
type WWWAuthentiate struct {
	Challenges []*Challenge
}

func ParseWWWAuthentiate(auth string) (*WWWAuthentiate, error) {
	wwwAuthentiate := &WWWAuthentiate{}
	for _, p := range splitChallenge(auth) {
		chal, err := ParseChallenge(p)
		if err != nil {
			return nil, err
		}
		if chal == nil {
			continue
		}
		wwwAuthentiate.Challenges = append(wwwAuthentiate.Challenges, chal)
	}
	if !wwwAuthentiate.Valid() {
		return nil, &BadStringError{"malformed WWW-Authenticate", auth}
	}
	return wwwAuthentiate, nil
}

func (w *WWWAuthentiate) String() string {
	var b strings.Builder
	if len(w.Challenges) == 0 {
		return ""
	}
	firstChalIdx := -1
	for idx, chal := range w.Challenges {
		if chal == nil {
			continue
		}
		b.WriteString(chal.String())
		firstChalIdx = idx
		break
	}
	if firstChalIdx == -1 || firstChalIdx == len(w.Challenges) {
		return ""
	}
	for _, chal := range w.Challenges[firstChalIdx+1:] {
		if chal == nil {
			continue
		}
		b.WriteRune(',')
		b.WriteString(chal.String())
	}
	return b.String()
}

func (w *WWWAuthentiate) Valid() bool {
	return len(w.Challenges) > 0
}

func (w *WWWAuthentiate) SetAuthHeader(r *http.Request) {
	r.Header.Set("WWW-Authenticate", w.String())
}

func split2Challenge(auth string) []string {
	if auth == "" {
		return nil
	}
	unparsedAuth := auth
	bufferedChal := ""
	for unparsedAuth != "" {
		s := strings.SplitN(unparsedAuth, ",", 2)
		if len(s) != 2 {
			bufferedChal = bufferedChal + unparsedAuth
			unparsedAuth = ""
			break
		}
		ss := strings.SplitN(s[0], "=", 2)
		if len(ss) != 2 { //malformed challenge
			return []string{auth}
		}
		if strings.Contains(ss[0], " ") {
			break
		}
		bufferedChal = bufferedChal + s[0]
		unparsedAuth = unparsedAuth[len(s[0]):]
	}
	return []string{
		bufferedChal, unparsedAuth,
	}
}

func splitChallenge(auth string) []string {
	unparsedAuth := auth

	a := []string{}
	for unparsedAuth != "" {
		s := split2Challenge(unparsedAuth)
		if len(s) == 0 || len(s) == 1 {
			a = append(a, unparsedAuth)
			unparsedAuth = ""
			break
		}
		a = append(a, s[0])
		unparsedAuth = s[1]
	}
	return a

}
