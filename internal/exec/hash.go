package exec

import (
	"crypto/sha256"
	"encoding/hex"
)

// ModelHash generates a short hash from model name for directory naming.
// Returns first 8 characters of SHA-256 hash.
func ModelHash(model string) string {
	hash := sha256.Sum256([]byte(model))
	return hex.EncodeToString(hash[:])[:8]
}
