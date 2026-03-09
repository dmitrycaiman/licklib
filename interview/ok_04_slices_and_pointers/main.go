package main

import "fmt"

type car struct {
	color   string
	mileage int
}

// Что будет выведено на экран?
func main() {
	cars := []car{{"red", 5000}, {"green", 10000}, {"blue", 7000}}

	carPtr := &cars[0]
	carPtr.mileage += 100

	cars = append(cars, car{color: "yellow", mileage: 15000})
	carPtr.mileage += 50

	fmt.Println(cars[0].mileage, cars[0].color)
	fmt.Println(carPtr.mileage, carPtr.color)
}
