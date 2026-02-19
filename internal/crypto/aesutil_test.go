package crypto

import (
	"bytes"
	"testing"
)

func TestAESGCM(t *testing.T) {
	t.Run("encrypt and decrypt", func(t *testing.T) {
		key := make([]byte, 32)
		for i := range key {
			key[i] = byte(i)
		}

		plaintext := []byte("hello world, this is a secret message")
		additionalData := []byte("context")

		ciphertext, err := EncryptAESGCM(key, plaintext, additionalData)
		if err != nil {
			t.Fatalf("EncryptAESGCM() error = %v", err)
		}

		decrypted, err := DecryptAESGCM(key, ciphertext, additionalData)
		if err != nil {
			t.Fatalf("DecryptAESGCM() error = %v", err)
		}

		if !bytes.Equal(plaintext, decrypted) {
			t.Errorf("decrypted != plaintext, got %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("tampered ciphertext fails", func(t *testing.T) {
		key := make([]byte, 32)
		plaintext := []byte("secret")
		aad := []byte("context")

		ciphertext, err := EncryptAESGCM(key, plaintext, aad)
		if err != nil {
			t.Fatalf("EncryptAESGCM() error = %v", err)
		}

		ciphertext[0] ^= 0xFF

		_, err = DecryptAESGCM(key, ciphertext, aad)
		if err == nil {
			t.Error("DecryptAESGCM() should fail with tampered ciphertext")
		}
	})

	t.Run("wrong additional data fails", func(t *testing.T) {
		key := make([]byte, 32)
		plaintext := []byte("secret")
		aad := []byte("context")

		ciphertext, err := EncryptAESGCM(key, plaintext, aad)
		if err != nil {
			t.Fatalf("EncryptAESGCM() error = %v", err)
		}

		_, err = DecryptAESGCM(key, ciphertext, []byte("wrong-context"))
		if err == nil {
			t.Error("DecryptAESGCM() should fail with wrong AAD")
		}
	})

	t.Run("invalid key size", func(t *testing.T) {
		key := make([]byte, 15)
		_, err := NewAESGCM(key)
		if err != ErrInvalidKeySize {
			t.Errorf("NewAESGCM() error = %v, want %v", err, ErrInvalidKeySize)
		}
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		key := make([]byte, 32)
		_, err := DecryptAESGCM(key, []byte("short"), nil)
		if err == nil {
			t.Error("DecryptAESGCM() should fail with short ciphertext")
		}
	})
}

func TestAESGCMStruct(t *testing.T) {
	t.Run("nonce size", func(t *testing.T) {
		key := make([]byte, 32)
		aesgcm, err := NewAESGCM(key)
		if err != nil {
			t.Fatalf("NewAESGCM() error = %v", err)
		}

		if aesgcm.NonceSize() != 12 {
			t.Errorf("NonceSize() = %d, want 12", aesgcm.NonceSize())
		}
	})

	t.Run("seal and open", func(t *testing.T) {
		key := make([]byte, 32)
		aesgcm, err := NewAESGCM(key)
		if err != nil {
			t.Fatalf("NewAESGCM() error = %v", err)
		}

		plaintext := []byte("test data")
		aad := []byte("aad")

		ciphertext, err := aesgcm.Seal(plaintext, aad)
		if err != nil {
			t.Fatalf("Seal() error = %v", err)
		}

		decrypted, err := aesgcm.Open(ciphertext, aad)
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}

		if !bytes.Equal(plaintext, decrypted) {
			t.Errorf("decrypted != plaintext")
		}
	})

	t.Run("open with short ciphertext", func(t *testing.T) {
		key := make([]byte, 32)
		aesgcm, err := NewAESGCM(key)
		if err != nil {
			t.Fatalf("NewAESGCM() error = %v", err)
		}

		_, err = aesgcm.Open([]byte("short"), nil)
		if err != ErrCiphertextShort {
			t.Errorf("Open() error = %v, want %v", err, ErrCiphertextShort)
		}
	})
}
