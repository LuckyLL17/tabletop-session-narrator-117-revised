package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func HashPassword(
	password string,
) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	digest :=
		digestPassword(
			salt, password)
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(digest[:])
}
func CheckPassword(
	encoded, password string,
) bool {
	parts := strings.SplitN(encoded, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	actual :=
		digestPassword(
			salt, password)
	return hex.EncodeToString(actual[:]) == parts[1]
}
func digestPassword(salt []byte, password string) [32]byte {
	return sha256.Sum256(append(append([]byte{}, salt...), []byte(password)...))
}
