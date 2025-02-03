package utils

import "crypto/rand"

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Create byte slice for random data
	randomBytes := make([]byte, length)

	// Read random bytes
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err)
	}

	// Convert random bytes to string using charset
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[randomBytes[i]%byte(len(charset))]
	}

	return string(result)
}
