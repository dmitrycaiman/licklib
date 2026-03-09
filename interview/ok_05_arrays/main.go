package main

import "fmt"

func modifyArray(arr [3]int) {
	arr[0] = 10
	fmt.Println("Inside modifyArray:", arr)
}

func modifySlice(slice []int) {
	slice[0] = 10
	fmt.Println("Inside modifySlice:", slice)
}

// Что будет выведено на экран?
func main() {
	array := [3]int{1, 2, 3}
	slice := array[:]

	fmt.Println("Before modifyArray:", array)
	modifyArray(array)
	fmt.Println("After modifyArray:", array)

	fmt.Println("Before modifySlice:", slice)
	modifySlice(slice)
	fmt.Println("After modifySlice:", slice)
	fmt.Println("Final array:", array)
}
