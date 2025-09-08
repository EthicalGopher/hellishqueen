package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"os"
)

var secretKey []byte

// Init loads the encryption key from an environment variable.
// It must be called once at application startup.
func Init() error {
	godotenv.Load()
	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		return fmt.Errorf("ENCRYPTION_KEY environment variable not set")
	}
	// AES-256 requires a 32-byte key, which is 64 hex characters.
	if len(keyHex) != 64 {
		return fmt.Errorf("ENCRYPTION_KEY must be a 64-character hex string for a 32-byte key")
	}

	var err error
	secretKey, err = hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("failed to decode ENCRYPTION_KEY from hex: %w", err)
	}
	return nil
}

// Encrypt takes a plaintext string and returns a hex-encoded encrypted string.
func Encrypt(text string) (string, error) {
	if secretKey == nil {
		return "", fmt.Errorf("crypto package not initialized")
	}

	plaintext := []byte(text)

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	// GCM is an authenticated encryption mode that provides confidentiality and authenticity.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// A nonce is a number used once. It's required for GCM and must be unique for each encryption.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal encrypts the data. We prepend the nonce to the ciphertext for use during decryption.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt takes a hex-encoded encrypted string and returns the original plaintext string.
func Decrypt(encryptedHex string) (string, error) {
	if secretKey == nil {
		return "", fmt.Errorf("crypto package not initialized")
	}

	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext is too short")
	}

	// The nonce was prepended to the ciphertext, so we split it off.
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// This error typically means the key is wrong or the data has been tampered with.
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
