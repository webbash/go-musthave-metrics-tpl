package main

import (
	"bytes"
	"fmt"
	"log"
)

func main() {
	var buf bytes.Buffer

	// допишите код
	// 1) создайте переменную типа *log.Logger
	// 2) запишите в неё нужные строки
	var logger = log.New(&buf, "mylog: ", 0)
	logger.Println("Hello, world!")
	logger.Println("Goodbye")

	fmt.Print(&buf)
	// должна вывести
	// mylog: Hello, world!
	// mylog: Goodbye
}
