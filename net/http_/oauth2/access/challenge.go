package access

import "strings"

// rfc1945 11
// challenge      = auth-scheme 1*SP realm *( "," auth-param )
// auth-scheme    = token
// auth-param     = token "=" quoted-string
// realm          = "realm" "=" realm-value
// realm-value    = quoted-string

type Challenge struct {
	AuthScheme string
	Realm      *Realm
	AuthParams []*AuthParam
}

func ParseChallenge(challenge string) (*Challenge, error) {
	s := strings.SplitN(challenge, " ", 2)
	if len(s) != 2 {
		return nil, &BadStringError{"malformed challenge", challenge}
	}
	chal := &Challenge{
		AuthScheme: s[0],
	}

	for _, p := range strings.Split(s[1], ",") {
		authParam, err := ParseAuthParam(p)
		if err != nil {
			return nil, err
		}
		if authParam == nil {
			continue
		}
		chal.AuthParams = append(chal.AuthParams, authParam)
	}

	return chal, nil
}

func (c *Challenge) String() string {
	var b strings.Builder
	b.WriteString(c.AuthScheme)
	b.WriteRune(' ')
	b.WriteString(c.Realm.String())
	for _, authParam := range c.AuthParams {
		if authParam == nil {
			continue
		}
		b.WriteRune(',')
		b.WriteString(authParam.String())
	}
	return b.String()
}
