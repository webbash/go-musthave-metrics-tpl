package main

import "fmt"

func main() {
	ch := generator("Hello")
	for msg := range ch {
		fmt.Println(msg)
	}
}

// Тут ваш генератор
func generator(msg string) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		for i := 0; i < 5; i++ {
			ch <- msg + fmt.Sprintf(" %d", i)
		}
	}()

	return ch
}
