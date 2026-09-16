package main

import (
	"bytes"
	"encoding/hex"
	"io"
	"os"
)

func main() {
	b := []byte("Yandex Practicum 🤔🤔🤔 Go Go")

	d := hex.Dumper(os.Stdout)
	defer d.Close()

	r := bytes.NewReader(b)
	io.Copy(d, r)
}
