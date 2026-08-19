package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSHA256Signer(t *testing.T) {
	signer := NewSHA256Signer("secret")
	hash := signer.Sign([]byte("payload"))

	require.Len(t, hash, 64)
	assert.True(t, signer.Verify([]byte("payload"), hash))
	assert.False(t, signer.Verify([]byte("other payload"), hash))
	assert.False(t, NewSHA256Signer("another secret").Verify([]byte("payload"), hash))
}
