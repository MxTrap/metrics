package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
)

type CryptoEncodeSvc struct {
	key *rsa.PublicKey
}

func NewEncrypterSvc(publicKeyPath string) (*CryptoEncodeSvc, error) {
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	publicKeyBlock, _ := pem.Decode(publicKeyBytes)

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return &CryptoEncodeSvc{
		key: publicKey,
	}, nil
}

func (svc *CryptoEncodeSvc) Encrypt(plaintext []byte) ([]byte, error) {
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		svc.key,
		plaintext,
		nil)
	if err != nil {
		return nil, err
	}
	return encryptedBytes, nil
}
