package access

import "strings"

// rfc1945 11.1
// auth-param     = token "=" quoted-string
type AuthParam struct {
	Name  string
	Value *QuotedString
}

func ParseAuthParam(authParam string) (*AuthParam, error) {
	s := strings.SplitN(authParam, "=", 2)
	if len(s) != 2 {
		return nil, &BadStringError{"malformed auth-param", authParam}
	}

	value, err := ParseQuotedString(s[1])
	if err != nil {
		return nil, err
	}

	return &AuthParam{
		Name:  s[0],
		Value: value,
	}, nil
}

func (r *AuthParam) String() string {
	var b strings.Builder
	b.WriteString(r.Name)
	b.WriteRune('=')
	b.WriteString(r.Value.String())
	return b.String()
}
