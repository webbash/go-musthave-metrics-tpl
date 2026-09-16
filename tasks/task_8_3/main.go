package main

import (
	"bytes"
	"fmt"
)

func main() {
	b := []byte("🌝🌖🌗🌘🌚🌒🌓🌔🌝")

	i := bytes.IndexAny(b, "🌚")

	fmt.Println(i)
}
