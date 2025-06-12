package main

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeyPair(t *testing.T) {
	privateKey, publicKey, err := generateKeyPair(2048) // Используем 2048 для ускорения тестов
	require.NoError(t, err, "generateKeyPair should succeed")
	assert.NotNil(t, privateKey, "private key should not be nil")
	assert.NotNil(t, publicKey, "public key should not be nil")
	assert.Equal(t, 2048, privateKey.N.BitLen(), "key size should be 2048 bits")
	assert.Equal(t, &privateKey.PublicKey, publicKey, "public key should match private key's public key")
}

func TestSaveKeyToFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "key-test")
	require.NoError(t, err, "failed to create temp dir")
	defer os.RemoveAll(tempDir)

	privateKey, _, err := generateKeyPair(2048)
	require.NoError(t, err, "generateKeyPair should succeed")

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	filename := filepath.Join(tempDir, "private.pem")
	err = saveKeyToFile(filename, privateKeyBlock)
	require.NoError(t, err, "saveKeyToFile should succeed")

	fileBytes, err := os.ReadFile(filename)
	require.NoError(t, err, "failed to read file")
	block, _ := pem.Decode(fileBytes)
	require.NotNil(t, block, "PEM block should not be nil")
	assert.Equal(t, "RSA PRIVATE KEY", block.Type, "PEM block type should match")
	assert.Equal(t, privateKeyBytes, block.Bytes, "PEM block bytes should match")
}

func TestSaveKeyToFileInvalidPath(t *testing.T) {
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("test"),
	}

	filename := "/invalid/path/private.pem"
	err := saveKeyToFile(filename, privateKeyBlock)
	assert.Error(t, err, "saveKeyToFile should fail with invalid path")
}
