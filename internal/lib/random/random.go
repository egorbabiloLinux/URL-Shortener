package random

import (
	"math/rand"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func NewRandomString(size int) string {
	result := make([]byte, size)

	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}