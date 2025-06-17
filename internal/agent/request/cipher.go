package request

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hash"
	"os"
)

type Cipher struct {
	publicKey *rsa.PublicKey
	hash      hash.Hash
}

func (c *Cipher) CipherJson(message []byte) (ciphered []byte, err error) {
	cipherByte, err := rsa.EncryptOAEP(
		c.hash,
		rand.Reader,
		c.publicKey,
		[]byte(message),
		[]byte(""),
	)
	c.hash.Reset()
	if err != nil {
		return message, err
	}
	return cipherByte, nil
}

func NewCipher(publicKeyPath string) (*Cipher, error) {
	pubPem, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubPem)
	if block == nil {
		return nil, errors.New("Failed to decode public key")
	}

	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	if key, ok := pub.(*rsa.PublicKey); ok {
		return &Cipher{
			publicKey: key,
			hash:      sha256.New(),
		}, nil
	}
	return nil, errors.New("Invalid public key format")
}
