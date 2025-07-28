package utils

import "math/rand"

const (
	letters       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lettersLength = len(letters)
)

func RandomString(length int) string {
	output := make([]byte, length)
	for i := range output {
		output[i] = letters[rand.Intn(lettersLength)]
	}
	return string(output)
}
