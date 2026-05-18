package excalidraw_test

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"testing"

	"obsidian-excalidraw-downloader/internal/excalidraw"
)

func TestDecrypt_TooShort(t *testing.T) {
	key := base64.RawURLEncoding.EncodeToString(make([]byte, 16))
	_, err := excalidraw.Decrypt(key, []byte("short"))
	if err == nil {
		t.Fatal("expected error for short payload")
	}
}

func TestDecryptGCM_InvalidKey(t *testing.T) {
	_, err := excalidraw.DecryptGCM("not-valid-base64!!!", []byte("ciphertext"), []byte("iv"))
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestDecrypt_Valid(t *testing.T) {
	keyBytes := []byte("1234567890123456")
	key := base64.RawURLEncoding.EncodeToString(keyBytes)
	plaintext := []byte("hello excalidraw")

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		t.Fatalf("create cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("create GCM: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	decrypted, err := excalidraw.Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecryptGCM_Valid(t *testing.T) {
	keyBytes := []byte("abcdefghijklmnop")
	key := base64.RawURLEncoding.EncodeToString(keyBytes)
	plaintext := []byte("secret drawing data")

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		t.Fatalf("create cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("create GCM: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	decrypted, err := excalidraw.DecryptGCM(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}
