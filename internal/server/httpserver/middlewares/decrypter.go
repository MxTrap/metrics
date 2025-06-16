package middlewares

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

type Decrypter struct {
	key *rsa.PrivateKey
}

func NewDecrypter(keyPath string) (*Decrypter, error) {
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	privateBlock, _ := pem.Decode(privateKeyBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return &Decrypter{
		key: privateKey,
	}, nil
}

func (d *Decrypter) DecrypterMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		var bodyBuffer bytes.Buffer

		_, err := bodyBuffer.ReadFrom(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err = c.Request.Body.Close(); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		decryptedBytes, err := d.key.Decrypt(nil, bodyBuffer.Bytes(), &rsa.OAEPOptions{Hash: crypto.SHA256})
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedBytes))

	}
}
