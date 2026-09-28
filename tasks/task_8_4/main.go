package main

import (
	"bytes"
	"fmt"
)

var alphabet map[rune]rune = map[rune]rune{
	'😅': 'a',
	'😎': 'e',
	'😂': 'g',
	'🥶': 'h',
	'🤓': 'j',
	'🥰': 'o',
	'😶': 'p',
	'🐱': 'r',
}

func main() {
	b := []byte("😂🥰😶🥶😎🐱🌚")

	fmt.Println(string(bytes.Map(func(r rune) rune {
		rune, ok := alphabet[r]
		if ok {
			return rune
		}
		return -1
	}, b)))
}
