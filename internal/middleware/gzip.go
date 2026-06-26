package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
			// который будем передавать следующей функции
			ow := w

			// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")

			if supportsGzip {
				// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
				cw := newCompressWriter(w)
				// меняем оригинальный http.ResponseWriter на новый
				ow = cw
				// не забываем отправить клиенту все сжатые данные после завершения middleware
				defer cw.Close()
			}

			// проверяем, что клиент отправил серверу сжатые данные в формате gzip
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
				cr, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// меняем тело запроса на новое
				r.Body = cr
				defer cr.Close()
			}

			next.ServeHTTP(ow, r)
		})
	}
}

type compressWriter struct {
	w           http.ResponseWriter
	zw          *gzip.Writer
	initialized bool
	compressed  bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{w: w, zw: gzip.NewWriter(w)}
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressWriter) Write(b []byte) (int, error) {
	cw.initCompression()

	if cw.compressed {
		return cw.zw.Write(b)
	}

	return cw.w.Write(b)
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	cw.initCompression()
	cw.w.WriteHeader(statusCode)
}

func (cw *compressWriter) Close() error {
	if !cw.compressed {
		return nil
	}

	return cw.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func (cw *compressWriter) initCompression() {
	if cw.initialized {
		return
	}

	contentType := cw.Header().Get("Content-Type")

	cw.compressed =
		strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html")

	if cw.compressed {
		cw.Header().Set("Content-Encoding", "gzip")
	}

	cw.initialized = true
}
