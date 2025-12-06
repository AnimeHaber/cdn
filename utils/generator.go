package utils

import (
	"math/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// GenerateRandomString generates a random string of fixed length
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateCDNFileName creates a filename in the format:
// XXXXX-XXXXX-XXXXX-XXXXX (5-5-5-5)
func GenerateCDNFileName() string {
	parts := make([]string, 4)
	for i := 0; i < 4; i++ {
		parts[i] = GenerateRandomString(5)
	}
	return strings.Join(parts, "-")
}
