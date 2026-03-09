package main

import "fmt"

// Что будет выведено на экран?
func main() {
	x, y := 10, 20
	defer func(val int) { fmt.Println("x:", val) }(x)
	defer func() { fmt.Println("y:", y) }()
	x, y = 100, 200
}
