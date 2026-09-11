package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
)

// DecryptMiddleware decrypts marked request bodies before compression and
// JSON middlewares/handlers process them.
func DecryptMiddleware(decryptor *crypto.Decryptor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(crypto.EncryptedHeader) != crypto.EncryptedHeaderValue {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read encrypted body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			plaintext, err := decryptor.Decrypt(body)
			if err != nil {
				http.Error(w, "failed to decrypt request body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(plaintext))
			r.ContentLength = int64(len(plaintext))
			next.ServeHTTP(w, r)
		})
	}
}
