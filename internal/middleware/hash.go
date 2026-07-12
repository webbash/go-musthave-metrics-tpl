package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	"go.uber.org/zap"
)

// It needs to save body content before we will send it to client.
// After we get content from response body we can sign it.
type hashResponseWriter struct {
	http.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

func (w *hashResponseWriter) Write(p []byte) (int, error) {
	return w.body.Write(p)
}

func (h *hashResponseWriter) WriteHeader(statusCode int) {
	h.statusCode = statusCode
}

func HashCheckMiddleware(signer *crypto.Sha256Signer, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hash := r.Header.Get("HashSHA256")
			if hash == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusInternalServerError)
				return
			}

			if !signer.Verify(body, hash) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Recover body content because we already read it
			r.Body = io.NopCloser(bytes.NewReader(body))

			wh := &hashResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wh, r)

			newHash := signer.Sign(wh.body.Bytes())

			w.Header().Set("HashSHA256", newHash)

			w.WriteHeader(wh.statusCode)
			_, err = w.Write(wh.body.Bytes())
			if err != nil {
				logger.Errorw("failed to write response", "error", err)
			}

			return
		})
	}
}
