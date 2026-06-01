package common

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

// GenerateReference returns a URL-safe opaque token for QR deep links (tables, rooms, etc.).
func GenerateReference() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return GenerateUUID()
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// GeneratePasscode returns an n-digit numeric passcode (e.g. for guest room access).
// It uses crypto/rand and falls back to a leading "0" only if the entropy source fails.
func GeneratePasscode(n int) string {
	if n <= 0 {
		n = 6
	}
	digits := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			digits[i] = '0'
			continue
		}
		digits[i] = byte('0' + num.Int64())
	}
	return string(digits)
}
