package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateDEK(t *testing.T) {
	t.Run("generates 32-byte key", func(t *testing.T) {
		dek, err := GenerateDEK()
		if err != nil {
			t.Fatalf("GenerateDEK() error = %v", err)
		}

		if len(dek.key) != 32 {
			t.Errorf("key size = %d, want 32", len(dek.key))
		}
	})

	t.Run("generates unique keys", func(t *testing.T) {
		dek1, err := GenerateDEK()
		if err != nil {
			t.Fatalf("GenerateDEK() error = %v", err)
		}

		dek2, err := GenerateDEK()
		if err != nil {
			t.Fatalf("GenerateDEK() error = %v", err)
		}

		if bytes.Equal(dek1.key, dek2.key) {
			t.Error("GenerateDEK() produced identical keys")
		}
	})
}

func TestDEKFromBytes(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
	}{
		{
			name: "32-byte key",
			key:  bytes.Repeat([]byte{0xAB}, 32),
		},
		{
			name: "16-byte key",
			key:  bytes.Repeat([]byte{0xCD}, 16),
		},
		{
			name: "empty key",
			key:  []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dek := DEKFromBytes(tt.key)

			if !bytes.Equal(dek.key, tt.key) {
				t.Error("DEKFromBytes() did not preserve key bytes")
			}
		})
	}
}

func TestDEKBytes(t *testing.T) {
	key := bytes.Repeat([]byte{0xEF}, 32)
	dek := DEKFromBytes(key)

	got := dek.Bytes()
	if !bytes.Equal(got, key) {
		t.Error("Bytes() returned wrong key")
	}
}

func TestDEKEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		keyName string
	}{
		{
			name:    "simple text",
			data:    []byte("Hello, World!"),
			keyName: "test-key",
		},
		{
			name:    "empty data",
			data:    []byte{},
			keyName: "empty-test",
		},
		{
			name:    "binary data",
			data:    []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC},
			keyName: "binary-key",
		},
		{
			name:    "large data",
			data:    bytes.Repeat([]byte("Lorem ipsum dolor sit amet. "), 1000),
			keyName: "large-key",
		},
		{
			name:    "unicode data",
			data:    []byte("Hello, 世界! 🌍 Привет мир"),
			keyName: "unicode-key",
		},
		{
			name:    "no key name",
			data:    []byte("test data"),
			keyName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dek, err := GenerateDEK()
			if err != nil {
				t.Fatalf("GenerateDEK() error = %v", err)
			}

			nonce, ciphertext, err := dek.Encrypt(tt.data, tt.keyName)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Verify nonce is returned
			if len(nonce) == 0 {
				t.Error("Encrypt() returned empty nonce")
			}

			// Verify ciphertext is longer than plaintext (due to nonce + tag)
			if len(ciphertext) <= len(tt.data) {
				t.Error("ciphertext too short")
			}

			// Decrypt and verify with same keyName
			plaintext, err := dek.Decrypt(ciphertext, tt.keyName)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(plaintext, tt.data) {
				t.Errorf("Decrypt() = %v, want %v", plaintext, tt.data)
			}
		})
	}
}

func TestDEKEncryptDeterministic(t *testing.T) {
	t.Run("same key name produces different ciphertexts", func(t *testing.T) {
		dek, _ := GenerateDEK()
		data := []byte("test data for encryption")
		keyName := "deterministic-test-key"

		nonce1, ciphertext1, err1 := dek.Encrypt(data, keyName)
		nonce2, ciphertext2, err2 := dek.Encrypt(data, keyName)

		if err1 != nil || err2 != nil {
			t.Fatalf("Encrypt() errors: %v, %v", err1, err2)
		}

		// Nonces should be different (random)
		if bytes.Equal(nonce1, nonce2) {
			t.Error("encryption must use random nonces - got same nonce twice!")
		}

		// Ciphertexts should be different due to different nonces
		if bytes.Equal(ciphertext1, ciphertext2) {
			t.Error("encryption should produce different ciphertexts with different nonces")
		}
	})
}

func TestEncryptUsesRandomNonce(t *testing.T) {
	dek, _ := GenerateDEK()
	plaintext := []byte("secret-value")

	// Encrypt same plaintext twice with same key name
	_, ciphertext1, _ := dek.Encrypt(plaintext, "DATABASE_URL")
	_, ciphertext2, _ := dek.Encrypt(plaintext, "DATABASE_URL")

	// Nonce is first 12 bytes - should be different
	if string(ciphertext1[:12]) == string(ciphertext2[:12]) {
		t.Error("Data encryption must use random nonces - got deterministic nonce!")
	}
}

func TestDEKDecryptErrors(t *testing.T) {
	t.Run("ciphertext too short", func(t *testing.T) {
		dek, _ := GenerateDEK()

		_, err := dek.Decrypt([]byte{0x01, 0x02}, "test-key") // Too short
		if err == nil {
			t.Error("expected error for short ciphertext")
		}
	})

	t.Run("corrupted ciphertext", func(t *testing.T) {
		dek, _ := GenerateDEK()
		data := []byte("test data")
		_, ciphertext, _ := dek.Encrypt(data, "test-key")

		// Corrupt a byte in the middle
		ciphertext[len(ciphertext)/2] ^= 0xFF

		_, err := dek.Decrypt(ciphertext, "test-key")
		if err == nil {
			t.Error("expected error for corrupted ciphertext")
		}
	})

	t.Run("wrong key decryption", func(t *testing.T) {
		dek1, _ := GenerateDEK()
		dek2, _ := GenerateDEK()
		data := []byte("test data")

		_, ciphertext, _ := dek1.Encrypt(data, "test-key")

		_, err := dek2.Decrypt(ciphertext, "test-key")
		if err == nil {
			t.Error("expected error when decrypting with wrong key")
		}
	})

	t.Run("wrong associated data", func(t *testing.T) {
		dek, _ := GenerateDEK()
		data := []byte("test data")
		_, ciphertext, _ := dek.Encrypt(data, "correct-key")

		_, err := dek.Decrypt(ciphertext, "wrong-key")
		if err == nil {
			t.Error("expected error for wrong associated data")
		}
	})
}

func TestDEKEncryptWithInvalidKey(t *testing.T) {
	t.Run("encrypt with invalid key size", func(t *testing.T) {
		// Create DEK with wrong size key (not 16, 24, or 32 bytes for AES)
		dek := DEKFromBytes([]byte{0x01, 0x02, 0x03}) // 3 bytes - invalid

		_, _, err := dek.Encrypt([]byte("test"), "key-name")
		if err == nil {
			t.Error("expected error for invalid key size")
		}
	})

	t.Run("decrypt with invalid key size", func(t *testing.T) {
		// Create DEK with wrong size key
		dek := DEKFromBytes([]byte{0x01, 0x02, 0x03}) // 3 bytes - invalid

		_, err := dek.Decrypt([]byte("some data that looks encrypted"), "key-name")
		if err == nil {
			t.Error("expected error for invalid key size")
		}
	})
}

func TestAssociatedDataPreventsSwapping(t *testing.T) {
	dek, _ := GenerateDEK()
	plaintext := []byte("secret-value")

	// Encrypt with context "DATABASE_URL"
	_, ciphertext, _ := dek.Encrypt(plaintext, "DATABASE_URL")

	// Try to decrypt as "API_KEY" - should fail
	_, err := dek.Decrypt(ciphertext, "API_KEY")
	if err == nil {
		t.Error("Associated data should prevent context swapping - decryption should fail")
	}
}
