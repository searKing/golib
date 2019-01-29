package access

import "strings"

// rfc1945 11
// realm          = "realm" "=" realm-value
// realm-value    = quoted-string
type Realm struct {
	RealmValue *QuotedString `json:"realm_value"`
}

func NewRealm(realm string) *Realm {
	return &Realm{
		RealmValue: NewQuotedString(realm),
	}
}
func ParseRealm(realm string) (*Realm, error) {
	s := strings.SplitN(realm, " ", 2)
	if len(s) != 2 || s[0] != "realm" {
		return nil, &BadStringError{"malformed realm", realm}
	}

	realmValue := s[1]
	value, err := ParseQuotedString(realmValue)
	if err != nil {
		return nil, err
	}

	return &Realm{
		RealmValue: value,
	}, nil
}

func (r *Realm) String() string {
	var b strings.Builder
	if r != nil && r.RealmValue != nil {
		b.WriteString("realm")
		b.WriteRune('=')
		b.WriteString(r.RealmValue.String())
	}
	return b.String()
}
