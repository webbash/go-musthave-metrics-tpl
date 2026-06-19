package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(sugar *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			responseData := &responseData{
				status: http.StatusOK,
				size:   0,
			}

			lw := &loggingResponseWriter{ResponseWriter: w, responseData: responseData}
			next.ServeHTTP(lw, r)

			duration := time.Since(start)

			sugar.Infow(
				"request completed",
				"uri", uri,
				"method", method,
				"duration", duration,
				"status", responseData.status,
				"size", responseData.size,
			)
		})
	}
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.responseData.status = status
}
