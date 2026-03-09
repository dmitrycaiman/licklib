package main

import "fmt"

// Какие проблемы есть в данном коде?
func main() {
	s := ""
	for i := range 1_000_000 {
		s += fmt.Sprintf("%d", i)
	}
}
