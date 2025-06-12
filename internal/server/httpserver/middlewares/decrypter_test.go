package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "failed to generate private key")
	return privateKey, &privateKey.PublicKey
}

func savePrivateKey(t *testing.T, privateKey *rsa.PrivateKey, path string) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	file, err := os.Create(path)
	require.NoError(t, err, "failed to create private key file")
	defer file.Close()
	err = pem.Encode(file, privateKeyBlock)
	require.NoError(t, err, "failed to write private key to file")
}

func TestNewDecrypter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "crypto-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	privateKey, _ := generateKeyPair(t)

	privateKeyPath := filepath.Join(tempDir, "private.pem")
	savePrivateKey(t, privateKey, privateKeyPath)

	decrypter, err := NewDecrypter(privateKeyPath)
	require.NoError(t, err, "NewDecrypter should succeed")
	assert.NotNil(t, decrypter, "decrypter should not be nil")
	assert.Equal(t, privateKey, decrypter.key, "private key should match")
}

func TestNewDecrypterInvalidPath(t *testing.T) {
	decrypter, err := NewDecrypter("/invalid/path/private.pem")
	assert.Error(t, err, "NewDecrypter should fail with invalid path")
	assert.Nil(t, decrypter, "decrypter should be nil")
}

func TestDecrypterMiddleware(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "crypto-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	privateKey, publicKey := generateKeyPair(t)

	privateKeyPath := filepath.Join(tempDir, "private.pem")
	savePrivateKey(t, privateKey, privateKeyPath)

	decrypter, err := NewDecrypter(privateKeyPath)
	require.NoError(t, err, "NewDecrypter should succeed")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", decrypter.DecrypterMiddleware(), func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, string(body))
	})

	plaintext := []byte("test message")
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plaintext, nil)
	require.NoError(t, err, "failed to encrypt plaintext")

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(ciphertext))
	require.NoError(t, err, "failed to create request")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "status should be OK")
	assert.Equal(t, string(plaintext), w.Body.String(), "response body should match plaintext")
}

func TestDecrypterMiddlewareInvalidCiphertext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "crypto-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	privateKey, _ := generateKeyPair(t)

	privateKeyPath := filepath.Join(tempDir, "private.pem")
	savePrivateKey(t, privateKey, privateKeyPath)

	decrypter, err := NewDecrypter(privateKeyPath)
	require.NoError(t, err, "NewDecrypter should succeed")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", decrypter.DecrypterMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, "should not reach here")
	})

	invalidCiphertext := []byte("invalid ciphertext")
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(invalidCiphertext))
	require.NoError(t, err, "failed to create request")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "status should be InternalServerError")
}
