// Package crypto provides HMAC-SHA256 signing helpers.
package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type SHA256Signer struct {
	key []byte
}

func NewSHA256Signer(key string) *SHA256Signer {
	return &SHA256Signer{key: []byte(key)}
}

func (s *SHA256Signer) Sign(data []byte) string {
	h := hmac.New(sha256.New, s.key)
	h.Write(data)
	dst := h.Sum(nil)

	return hex.EncodeToString(dst)
}

func (s *SHA256Signer) Verify(data []byte, hash string) bool {
	expected := s.Sign(data)

	return hmac.Equal([]byte(expected), []byte(hash))
}
