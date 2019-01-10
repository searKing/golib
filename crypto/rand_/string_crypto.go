package rand_

import (
	"crypto/rand"
	"math/big"
)

// take in a character set and a length and will generate a random string using that character set.
func StringCryptoWithCharset(length int64, charset string) (string, error) {

	maxBigInt := big.NewInt(int64(len(charset)))

	b := make([]byte, length)
	for i := range b {
		r, err := rand.Int(rand.Reader, maxBigInt)
		if err != nil {
			return "", err
		}
		b[i] = charset[r.Int64()]
	}
	return string(b), nil
}

// only take in a length, and will use a default characters set to generate a random string
func StringCrypto(length int64) (string, error) {
	return StringCryptoWithCharset(length, CharsetAlphaNum)
}
