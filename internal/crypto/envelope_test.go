package crypto

import (
	"bytes"
	"strings"
	"testing"
)

func deriveMasterKeyFromIdentity(t *testing.T) *MasterKey {
	t.Helper()
	identity, err := GenerateAgeKeyPair()
	if err != nil {
		t.Fatalf("GenerateAgeKeyPair() error = %v", err)
	}
	strategy := NewAsymmetricStrategy(identity)
	mk, err := strategy.GetMasterKey()
	if err != nil {
		t.Fatalf("GetMasterKey() error = %v", err)
	}
	return mk
}

func TestParseEncryptedValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *EncryptedValue
		wantErr bool
	}{
		{
			name:  "valid encrypted value",
			input: "envx:d3JhcHBlZC1kZWs=:Y2lwaGVydGV4dA==",
			want: &EncryptedValue{
				WrappedDEK: "d3JhcHBlZC1kZWs=",
				Ciphertext: "Y2lwaGVydGV4dA==",
			},
			wantErr: false,
		},
		{
			name:    "not encrypted prefix",
			input:   "plaintext-value",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing separator",
			input:   "envx:only-one-part",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "empty parts",
			input: "envx::",
			want: &EncryptedValue{
				WrappedDEK: "",
				Ciphertext: "",
			},
			wantErr: false,
		},
		{
			name:  "multiple colons in ciphertext",
			input: "envx:d3JhcHBlZC1kZWs=:Y2k6cGhlcjp0ZXh0",
			want: &EncryptedValue{
				WrappedDEK: "d3JhcHBlZC1kZWs=",
				Ciphertext: "Y2k6cGhlcjp0ZXh0",
			},
			wantErr: false,
		},
		{
			name:    "enc: prefix rejected",
			input:   "enc:d3JhcHBlZC1kZWs=:Y2lwaGVydGV4dA==",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEncryptedValue(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEncryptedValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.WrappedDEK != tt.want.WrappedDEK {
					t.Errorf("WrappedDEK = %v, want %v", got.WrappedDEK, tt.want.WrappedDEK)
				}
				if got.Ciphertext != tt.want.Ciphertext {
					t.Errorf("Ciphertext = %v, want %v", got.Ciphertext, tt.want.Ciphertext)
				}
			}
		})
	}
}

func TestEncryptedValueString(t *testing.T) {
	tests := []struct {
		name string
		ev   EncryptedValue
		want string
	}{
		{
			name: "normal values",
			ev: EncryptedValue{
				WrappedDEK: "wrapped-dek",
				Ciphertext: "ciphertext",
			},
			want: "envx:wrapped-dek:ciphertext",
		},
		{
			name: "empty values",
			ev: EncryptedValue{
				WrappedDEK: "",
				Ciphertext: "",
			},
			want: "envx::",
		},
		{
			name: "with special characters",
			ev: EncryptedValue{
				WrappedDEK: "d3JhcHBlZC1kZWs=",
				Ciphertext: "Y2lwaGVydGV4dA==",
			},
			want: "envx:d3JhcHBlZC1kZWs=:Y2lwaGVydGV4dA==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ev.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEnvelope(t *testing.T) {
	mk := deriveMasterKeyFromIdentity(t)
	env := NewEnvelope(mk)

	if env == nil {
		t.Fatal("NewEnvelope() returned nil")
	}

	if env.masterKey == nil {
		t.Error("NewEnvelope() masterKey is nil")
	}
}

func TestEnvelopeEncryptDecrypt(t *testing.T) {
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
			data:    bytes.Repeat([]byte("Large data block "), 100),
			keyName: "large-key",
		},
		{
			name:    "unicode data",
			data:    []byte("Hello, 世界! 🌍"),
			keyName: "unicode-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mk := deriveMasterKeyFromIdentity(t)
			env := NewEnvelope(mk)

			ev, err := env.Encrypt(tt.data, tt.keyName)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Verify format
			if !strings.HasPrefix(ev.String(), EncryptedValuePrefix) {
				t.Error("encrypted value missing envx: prefix")
			}

			// Verify parts are base64
			if ev.WrappedDEK == "" {
				t.Error("encrypted value has empty WrappedDEK")
			}
			if ev.Ciphertext == "" {
				t.Error("encrypted value has empty Ciphertext")
			}

			// Decrypt with keyName
			plaintext, err := env.Decrypt(ev, tt.keyName)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(plaintext, tt.data) {
				t.Errorf("Decrypt() = %v, want %v", plaintext, tt.data)
			}
		})
	}
}

func TestEnvelopeDeterministicEncryption(t *testing.T) {
	t.Run("same key name produces decryptable values", func(t *testing.T) {
		mk := deriveMasterKeyFromIdentity(t)
		env := NewEnvelope(mk)
		data := []byte("test data for deterministic test")
		keyName := "deterministic-key"

		// Encrypt twice with same key name
		ev1, err1 := env.Encrypt(data, keyName)
		ev2, err2 := env.Encrypt(data, keyName)

		if err1 != nil || err2 != nil {
			t.Fatalf("encryption errors: %v, %v", err1, err2)
		}

		// Both should be decryptable and produce same plaintext
		plain1, err := env.Decrypt(ev1, keyName)
		if err != nil {
			t.Fatalf("decrypt ev1 error: %v", err)
		}

		plain2, err := env.Decrypt(ev2, keyName)
		if err != nil {
			t.Fatalf("decrypt ev2 error: %v", err)
		}

		if !bytes.Equal(plain1, data) || !bytes.Equal(plain2, data) {
			t.Error("decrypted values don't match original")
		}
	})

	t.Run("different key names produce different encrypted values", func(t *testing.T) {
		mk := deriveMasterKeyFromIdentity(t)
		env := NewEnvelope(mk)
		data := []byte("test data")

		ev1, _ := env.Encrypt(data, "key-a")
		ev2, _ := env.Encrypt(data, "key-b")

		if ev1.String() == ev2.String() {
			t.Error("different key names should produce different encrypted values")
		}
	})
}

func TestEnvelopeDecryptErrors(t *testing.T) {
	mk := deriveMasterKeyFromIdentity(t)
	env := NewEnvelope(mk)

	t.Run("invalid base64 wrapped DEK", func(t *testing.T) {
		ev := &EncryptedValue{
			WrappedDEK: "!!!invalid-base64!!!",
			Ciphertext: "dGVzdA==",
		}

		_, err := env.Decrypt(ev, "test-key")
		if err == nil {
			t.Error("expected error for invalid wrapped DEK base64")
		}
	})

	t.Run("invalid base64 ciphertext", func(t *testing.T) {
		// First create a valid encrypted value
		data := []byte("test")
		ev, _ := env.Encrypt(data, "test-key")

		// Corrupt the ciphertext
		ev.Ciphertext = "!!!invalid!!!"

		_, err := env.Decrypt(ev, "test-key")
		if err == nil {
			t.Error("expected error for invalid ciphertext base64")
		}
	})

	t.Run("corrupted wrapped DEK", func(t *testing.T) {
		data := []byte("test")
		ev, _ := env.Encrypt(data, "test-key")

		// Corrupt the wrapped DEK (but keep valid base64)
		ev.WrappedDEK = "Y29ycnVwdGVk" // "corrupted" in base64

		_, err := env.Decrypt(ev, "test-key")
		if err == nil {
			t.Error("expected error for corrupted wrapped DEK")
		}
	})

	t.Run("corrupted ciphertext", func(t *testing.T) {
		mk1 := deriveMasterKeyFromIdentity(t)
		env1 := NewEnvelope(mk1)
		data := []byte("test")
		ev, _ := env1.Encrypt(data, "test-key")

		// Create different envelope with different key
		mk2 := deriveMasterKeyFromIdentity(t)
		env2 := NewEnvelope(mk2)

		_, err := env2.Decrypt(ev, "test-key")
		if err == nil {
			t.Error("expected error when decrypting with wrong master key")
		}
	})
}

func TestWrapDEKUsesRandomNonce(t *testing.T) {
	masterKey := deriveMasterKeyFromIdentity(t)
	envelope := NewEnvelope(masterKey)
	dek1, _ := GenerateDEK()
	dek2, _ := GenerateDEK()

	wrapped1, _ := envelope.wrapDEK(dek1)
	wrapped2, _ := envelope.wrapDEK(dek2)

	// First 12 bytes are nonce - should be different
	if string(wrapped1[:12]) == string(wrapped2[:12]) {
		t.Error("DEK wrapping must use random nonces - got same nonce twice!")
	}
}

func TestEnvelopeRoundTrip(t *testing.T) {
	mk := deriveMasterKeyFromIdentity(t)
	env := NewEnvelope(mk)

	// Encrypt
	original := []byte("secret data to be encrypted")
	keyName := "my-secret-key"
	ev, err := env.Encrypt(original, keyName)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Serialize to string
	serialized := ev.String()

	// Parse back
	parsed, err := ParseEncryptedValue(serialized)
	if err != nil {
		t.Fatalf("ParseEncryptedValue() error = %v", err)
	}

	// Decrypt with keyName
	decrypted, err := env.Decrypt(parsed, keyName)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, original) {
		t.Error("round-trip decryption failed")
	}
}

func TestReencryptionSafety(t *testing.T) {
	masterKey := deriveMasterKeyFromIdentity(t)
	envelope := NewEnvelope(masterKey)

	// Encrypt, then re-encrypt same value
	ev1, _ := envelope.Encrypt([]byte("password123"), "DATABASE_URL")
	ev2, _ := envelope.Encrypt([]byte("password123"), "DATABASE_URL")

	// Both should decrypt correctly
	dec1, _ := envelope.Decrypt(ev1, "DATABASE_URL")
	dec2, _ := envelope.Decrypt(ev2, "DATABASE_URL")

	if string(dec1) != "password123" || string(dec2) != "password123" {
		t.Error("Re-encryption should produce valid decryptable values")
	}

	// Ciphertexts should be different (different DEKs and nonces)
	if ev1.Ciphertext == ev2.Ciphertext {
		t.Error("Re-encryption should produce different ciphertexts")
	}
}
