package shortener

import "math/rand"

// charset contains all allowed characters for a short code (Base62).
// 62^6 = ~56 billion possible combinations.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateCode returns a random string of the given length using Base62 characters.
func GenerateCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}
