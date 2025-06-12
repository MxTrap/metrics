package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
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

func savePublicKey(t *testing.T, publicKey *rsa.PublicKey, path string) {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	publicKeyBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	file, err := os.Create(path)
	require.NoError(t, err, "failed to create public key file")
	defer file.Close()
	err = pem.Encode(file, publicKeyBlock)
	require.NoError(t, err, "failed to write public key to file")
}

func TestNewEncrypterSvc(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "crypto-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	_, publicKey := generateKeyPair(t)

	publicKeyPath := filepath.Join(tempDir, "public.pem")
	savePublicKey(t, publicKey, publicKeyPath)

	svc, err := NewEncrypterSvc(publicKeyPath)
	require.NoError(t, err, "NewEncrypterSvc should succeed")
	assert.NotNil(t, svc, "service should not be nil")
	assert.Equal(t, publicKey, svc.key, "public key should match")
}

func TestNewEncrypterSvcInvalidPath(t *testing.T) {
	svc, err := NewEncrypterSvc("/invalid/path/public.pem")
	assert.Error(t, err, "NewEncrypterSvc should fail with invalid path")
	assert.Nil(t, svc, "service should be nil")
}

func TestEncrypt(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "crypto-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	privateKey, publicKey := generateKeyPair(t)

	publicKeyPath := filepath.Join(tempDir, "public.pem")
	savePublicKey(t, publicKey, publicKeyPath)

	svc, err := NewEncrypterSvc(publicKeyPath)
	require.NoError(t, err, "NewEncrypterSvc should succeed")

	plaintext := []byte("test message")
	ciphertext, err := svc.Encrypt(plaintext)
	require.NoError(t, err, "Encrypt should succeed")
	assert.NotEqual(t, plaintext, ciphertext, "ciphertext should differ from plaintext")

	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	require.NoError(t, err, "Decrypt should succeed")
	assert.Equal(t, plaintext, decrypted, "decrypted text should match plaintext")
}

func TestEncryptEmptyPlaintext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "crypto-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	_, publicKey := generateKeyPair(t)

	publicKeyPath := filepath.Join(tempDir, "public.pem")
	savePublicKey(t, publicKey, publicKeyPath)

	svc, err := NewEncrypterSvc(publicKeyPath)
	require.NoError(t, err, "NewEncrypterSvc should succeed")

	plaintext := []byte("")
	ciphertext, err := svc.Encrypt(plaintext)
	require.NoError(t, err, "Encrypt should succeed with empty plaintext")
	assert.NotEmpty(t, ciphertext, "ciphertext should not be empty")
}

func TestEncryptLargePlaintext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "crypto-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	_, publicKey := generateKeyPair(t)

	publicKeyPath := filepath.Join(tempDir, "public.pem")
	savePublicKey(t, publicKey, publicKeyPath)

	svc, err := NewEncrypterSvc(publicKeyPath)
	require.NoError(t, err, "NewEncrypterSvc should succeed")

	plaintext := make([]byte, 1000)
	_, err = rand.Read(plaintext)
	require.NoError(t, err, "failed to generate random plaintext")

	ciphertext, err := svc.Encrypt(plaintext)
	assert.Error(t, err, "Encrypt should fail with large plaintext")
	assert.Nil(t, ciphertext, "ciphertext should be nil")
}
