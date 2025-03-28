package util

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"os"

	"github.com/pocketbase/pocketbase/tools/security"
)

// LooksLikeEncrypted checks if the given string looks like base64 encoded and AES encrypted data.
// This is format used by pocketbase encryption.
func LooksLikeEncrypted(s string) bool {
	ciphertext, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}

	// Use a dummy 32-byte key since we only want the nonce size
	dummyKey := make([]byte, 32)

	block, err := aes.NewCipher(dummyKey)
	if err != nil {
		return false
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return false
	}

	nonceSize := gcm.NonceSize()

	// At minimum: nonce + GCM auth tag (16 bytes)
	return len(ciphertext) >= nonceSize+16
}

// Return true if the provided value can be decrypted using the secret key
func CanDecryptSecret(ciphertext string) bool {
	encryptionKey := os.Getenv("POCKETBASE_ENCRYPTION_KEY")
	if len(encryptionKey) == 0 {
		return false
	}

	decryptedSecret, err := security.Decrypt(ciphertext, encryptionKey)

	if len(decryptedSecret) > 0 && err == nil {
		return true
	}

	return false
}
