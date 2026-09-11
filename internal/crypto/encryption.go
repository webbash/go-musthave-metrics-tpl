package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const (
	// EncryptedHeader marks requests whose body is encrypted by the agent.
	EncryptedHeader      = "X-Encrypted"
	EncryptedHeaderValue = "rsa-oaep-sha256"
)

// Encryptor encrypts data with the server's public RSA key.
type Encryptor struct {
	publicKey *rsa.PublicKey
}

// Decryptor decrypts data with the server's private RSA key.
type Decryptor struct {
	privateKey *rsa.PrivateKey
}

// NewEncryptor creates an encryptor from a public RSA key.
func NewEncryptor(publicKey *rsa.PublicKey) *Encryptor {
	return &Encryptor{publicKey: publicKey}
}

// NewDecryptor creates a decryptor from a private RSA key.
func NewDecryptor(privateKey *rsa.PrivateKey) *Decryptor {
	return &Decryptor{privateKey: privateKey}
}

// Encrypt encrypts one message with RSA-OAEP and SHA-256.
func (e *Encryptor) Encrypt(message []byte) ([]byte, error) {
	if e == nil || e.publicKey == nil {
		return nil, errors.New("encrypt: public key is nil")
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, message, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypt message: %w", err)
	}

	return ciphertext, nil
}

// Decrypt decrypts one RSA-OAEP message authenticated with SHA-256.
func (d *Decryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if d == nil || d.privateKey == nil {
		return nil, errors.New("decrypt: private key is nil")
	}

	message, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, d.privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt message: %w", err)
	}

	return message, nil
}

// LoadPublicKey reads an RSA public key from either an X.509 certificate,
// PKIX public key, or PKCS#1 public key PEM file.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("decode public key PEM: block not found")
	}

	if certificate, err := x509.ParseCertificate(block.Bytes); err == nil {
		publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate has unsupported public key type %T", certificate.PublicKey)
		}
		return publicKey, nil
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		publicKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key has unsupported type %T", key)
		}
		return publicKey, nil
	}

	publicKey, pkcs1Err := x509.ParsePKCS1PublicKey(block.Bytes)
	if pkcs1Err != nil {
		return nil, fmt.Errorf("parse public key: pkix: %v; pkcs1: %w", err, pkcs1Err)
	}

	return publicKey, nil
}

// LoadPrivateKey reads an RSA private key from a PKCS#8 or PKCS#1 PEM file.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("decode private key PEM: block not found")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		privateKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key has unsupported type %T", key)
		}
		return privateKey, nil
	}

	privateKey, pkcs1Err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if pkcs1Err != nil {
		return nil, fmt.Errorf("parse private key: pkcs8: %v; pkcs1: %w", err, pkcs1Err)
	}

	return privateKey, nil
}
