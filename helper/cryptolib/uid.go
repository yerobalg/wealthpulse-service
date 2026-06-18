package cryptolib

import (
	"crypto/rand"
	"fmt"
)

const uidCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateShortUID returns a short UID in the format "J-XXXXXX" using 6 random
// uppercase-alphanumeric characters sourced from crypto/rand.
func GenerateShortUID(prefix string) (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i, v := range b {
		b[i] = uidCharset[int(v)%len(uidCharset)]
	}
	return fmt.Sprintf("%s-%s", prefix, b), nil
}
