package excalidraw

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// Decrypt decrypts AES-128-GCM data produced by Excalidraw's web client.
//
// Wire format: 12-byte IV || ciphertext || 16-byte GCM tag
// keyStr: base64url (no padding) encoded raw 128-bit key bytes
func Decrypt(keyStr string, data []byte) ([]byte, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("payload too short: %d bytes", len(data))
	}

	keyBytes, err := base64.RawURLEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	iv := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

// DecryptGCM decrypts AES-GCM data where the IV is provided separately.
// Used for room broadcasts where IV and ciphertext arrive as separate Socket.IO arguments.
func DecryptGCM(keyStr string, ciphertext, iv []byte) ([]byte, error) {
	keyBytes, err := base64.RawURLEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}
