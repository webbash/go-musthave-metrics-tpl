package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	encryptor := NewEncryptor(&privateKey.PublicKey)
	decryptor := NewDecryptor(privateKey)
	plaintext := []byte(`[{"id":"Alloc","type":"gauge","value":12.5}]`)

	ciphertext, err := encryptor.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := decryptor.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptRejectsTamperedMessage(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ciphertext, err := NewEncryptor(&privateKey.PublicKey).Encrypt([]byte("payload"))
	require.NoError(t, err)
	ciphertext[len(ciphertext)-1] ^= 1

	_, err = NewDecryptor(privateKey).Decrypt(ciphertext)
	assert.Error(t, err)
}

func TestLoadKeys(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	dir := t.TempDir()
	privatePath := filepath.Join(dir, "private.pem")
	publicPath := filepath.Join(dir, "public.pem")

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicBytes})
	require.NoError(t, os.WriteFile(privatePath, privatePEM, 0600))
	require.NoError(t, os.WriteFile(publicPath, publicPEM, 0644))

	loadedPrivate, err := LoadPrivateKey(privatePath)
	require.NoError(t, err)
	loadedPublic, err := LoadPublicKey(publicPath)
	require.NoError(t, err)
	assert.Equal(t, privateKey.PublicKey.N, loadedPublic.N)
	assert.Equal(t, privateKey.N, loadedPrivate.N)
}
