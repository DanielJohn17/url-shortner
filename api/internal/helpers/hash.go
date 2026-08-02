package helpers

import "crypto/sha256"

func HashString(data string) [32]byte {
	return sha256.Sum256([]byte(data))
}
