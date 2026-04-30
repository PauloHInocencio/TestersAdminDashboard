package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func GenerateToken() (string, error) {
	// Create a 32-byte buffer: Allocates a byte slice to hold random data
	bytes := make([]byte, 32)

	// Fills it with cryptographically secure random bytes:
	// Uses crypto/rand.Read() to generate random data from the system's secure random number generator.
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	//Converts the random bytes to a base64-encoded string using RawURLEncoding, which:
	//	- Uses URL-safe characters (replacing + with - and / with _)
	//	- Omits padding characters (=) at the end
	//	- Results in a 43-character string (32 bytes → 43 chars in base64 without padding)
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashToken(token string) string {
	// Takes the input token string, converts it to bytes, and computes its SHA-256 hash using sha256.Sum256().
	// - sha256.Sum256() returns a fixed-size array [32]byte (not a slice)
	hash := sha256.Sum256([]byte(token))

	//Converts the hash bytes to a hexadecimal string representation.
	//- hex.EncodeToString() expects a slice []byte as input (not an array)
	//- hash[:] creates a slice that references the entire array
	return hex.EncodeToString(hash[:])
}
