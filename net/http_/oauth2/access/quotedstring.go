package access

type QuotedString struct {
	UnquotedString string
}

func NewQuotedString(str string) *QuotedString {
	return &QuotedString{
		UnquotedString: str,
	}
}
func ParseQuotedString(raw string) (*QuotedString, error) {
	unquotedString, err := parseQuotedString(raw)
	if err != nil {
		return nil, err
	}
	return &QuotedString{
		UnquotedString: unquotedString,
	}, nil
}
func (q *QuotedString) String() string {
	unquoted := ""
	if q != nil {
		unquoted = q.UnquotedString
	}
	return sanitizeQuotedString(unquoted)
}

func parseQuotedString(raw string) (string, error) {
	// Strip the quotes.
	if len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		return raw[1 : len(raw)-1], nil
	}
	return "", &BadStringError{"malformed quoted-string", raw}
}

func sanitizeQuotedString(v string) string {
	return `"` + v + `"`
}
