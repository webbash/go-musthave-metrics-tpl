package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Sha256Signer struct {
	key []byte
}

func NewSha256Signer(key string) *Sha256Signer {
	return &Sha256Signer{key: []byte(key)}
}

func (s *Sha256Signer) Sign(data []byte) string {
	h := hmac.New(sha256.New, s.key)
	h.Write(data)
	dst := h.Sum(nil)

	return hex.EncodeToString(dst)
}

func (s *Sha256Signer) Verify(data []byte, hash string) bool {
	expected := s.Sign(data)

	return hmac.Equal([]byte(expected), []byte(hash))
}
