package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

// generateKeyPair генерирует пару RSA-ключей.
func generateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// saveKeyToFile сохраняет ключ в PEM-файл.
func saveKeyToFile(filename string, block *pem.Block) error {
	if err := os.MkdirAll("keys", 0755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Fatal(closeErr)
		}
	}()
	return pem.Encode(file, block)
}

func main() {
	privateKey, publicKey, err := generateKeyPair(16384)
	if err != nil {
		log.Fatal(err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	if err := saveKeyToFile("./keys/private.pem", privateKeyBlock); err != nil {
		log.Fatal("error when saving private key: " + err.Error())
	}

	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	if err := saveKeyToFile("./keys/public.pem", publicKeyBlock); err != nil {
		log.Fatal("error when saving public key: " + err.Error())
	}
}
