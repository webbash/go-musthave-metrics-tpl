package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitedWriter struct {
	w       io.Writer
	limiter *rate.Limiter
}

func NewRateLimitedWriter(w io.Writer, ops int) *RateLimitedWriter {
	return &RateLimitedWriter{
		w:       w,
		limiter: rate.NewLimiter(rate.Every(time.Second), ops),
	}
}

func (w *RateLimitedWriter) Write(p []byte) (n int, err error) {
	if !w.limiter.Allow() {
		return 0, fmt.Errorf("rate limit exceeded")
	}

	return w.w.Write(p)
}

func main() {
	var byteBuffer bytes.Buffer
	writer := NewRateLimitedWriter(&byteBuffer, 10)
	for i := 0; i < 10; i++ {
		_, err := writer.Write([]byte(strconv.Itoa(i)))
		if err != nil {
			panic(err)
		}
	}
	fmt.Println(byteBuffer.String())

	byteBuffer.Reset()
	writer = NewRateLimitedWriter(&byteBuffer, 9)
	for i := 0; i < 10; i++ {
		_, err := writer.Write([]byte(strconv.Itoa(i)))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}
