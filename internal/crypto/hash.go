package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// PreHashPassword combines username:password and SHA-256 hashes it.
func PreHashPassword(username, password string) string {
	h := sha256.New()
	h.Write([]byte(username))
	h.Write([]byte(":"))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}
