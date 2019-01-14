package jwt

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
)

type AuthKey struct {
	alg string

	// Private key
	privKey *rsa.PrivateKey

	// Public key
	pubKey *rsa.PublicKey

	// Secret key used for signing. Required.
	symmetricKey []byte
}

// SymmetricKey : key
// else: privKey publicKey
func NewAuthKey(alg string, keys ...[]byte) *AuthKey {
	authKey := &AuthKey{
		alg: alg,
	}
	if len(keys) == 0 {
		return authKey
	}

	if authKey.IsSymmetricKey() {
		authKey.symmetricKey = keys[0]
		return authKey
	}

	authKey.setPrivateKey(keys[0])
	if len(keys) >= 2 {
		authKey.setPublicKey(keys[1])
	}
	return authKey
}
func NewAuthKeyFromFile(alg string, keyFiles ...string) *AuthKey {
	authKey := NewAuthKey(alg)

	if len(keyFiles) == 0 {
		return authKey
	}

	if authKey.IsSymmetricKey() {
		authKey.setKeyFile(keyFiles[0], authKey.setSymmetricKey)
		return authKey
	}

	authKey.setKeyFile(keyFiles[0], authKey.setPrivateKey)
	if len(keyFiles) >= 2 {
		authKey.setKeyFile(keyFiles[0], authKey.setPrivateKey)
	}
	return authKey
}

func (a *AuthKey) setKeyFile(keyFile string, keyFunc func(keyData []byte) error) error {
	keyData, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return err
	}
	return keyFunc(keyData)
}

func (a *AuthKey) setSymmetricKey(keyData []byte) error {
	a.symmetricKey = keyData
	return nil
}

func (a *AuthKey) setPrivateKey(keyData []byte) error {
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return err
	}
	a.privKey = privKey
	return nil
}

func (a *AuthKey) setPublicKey(keyData []byte) error {
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return err
	}
	a.pubKey = pubKey
	return nil
}

func (a *AuthKey) GetSignedKey(token *jwt.Token) (interface{}, error) {
	if token != nil && jwt.GetSigningMethod(a.alg) != token.Method {
		return nil, ErrInvalidSigningAlgorithm
	}
	if a.IsSymmetricKey() {
		return a.symmetricKey, nil
	}
	return a.pubKey, nil
}
func (a *AuthKey) GetSignedMethod() jwt.SigningMethod {
	return jwt.GetSigningMethod(a.alg)
}
func (a *AuthKey) IsSymmetricKey() bool {
	switch a.alg {
	case SigningMethodRS256, SigningMethodRS384, SigningMethodRS512:
		return false
	}
	return true
}
