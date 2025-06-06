package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

func main() {

	privateKey, err := rsa.GenerateKey(rand.Reader, 16384)
	if err != nil {
		log.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	err = os.MkdirAll("keys", 0755)
	if err != nil {
		log.Fatal(err)
	}
	privatePem, err := os.Create("./keys/private.pem")
	defer func(privatePem *os.File) {
		err := privatePem.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(privatePem)
	if err != nil {
		log.Fatal("error when create private.pem: " + err.Error())
	}
	err = pem.Encode(privatePem, privateKeyBlock)
	if err != nil {
		log.Fatal("error when encode private pem: " + err.Error())
	}

	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicPem, err := os.Create("./keys/public.pem")
	defer func(publicPem *os.File) {
		err := publicPem.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(publicPem)
	if err != nil {
		log.Fatal("error when create public.pem: " + err.Error())
	}
	err = pem.Encode(publicPem, publicKeyBlock)
	if err != nil {
		log.Fatal("error when encode public pem: " + err.Error())
	}
}
