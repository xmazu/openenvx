package crypto

import (
	"testing"

	"github.com/xmazu/openenvx/internal/config"
)

func TestAsymmetricStrategy(t *testing.T) {
	t.Run("creates strategy with identity", func(t *testing.T) {
		identity, err := GenerateAgeKeyPair()
		if err != nil {
			t.Fatalf("GenerateAgeKeyPair() error = %v", err)
		}

		strategy := NewAsymmetricStrategy(identity)

		if strategy == nil {
			t.Fatal("NewAsymmetricStrategy() returned nil")
		}

		if strategy.identity != identity {
			t.Error("strategy should store identity")
		}

		if strategy.publicKey != identity.Recipient().String() {
			t.Error("strategy should set public key from identity")
		}
	})

	t.Run("requires identity for master key", func(t *testing.T) {
		// Create strategy without identity
		strategy := NewAsymmetricStrategy(nil)

		_, err := strategy.GetMasterKey()
		if err == nil {
			t.Error("GetMasterKey() should error without identity")
		}
	})

	t.Run("derives master key with identity", func(t *testing.T) {
		identity, _ := GenerateAgeKeyPair()
		strategy := NewAsymmetricStrategy(identity)

		masterKey, err := strategy.GetMasterKey()
		if err != nil {
			t.Fatalf("GetMasterKey() error = %v", err)
		}

		if masterKey == nil {
			t.Fatal("GetMasterKey() returned nil")
		}

		// Verify caching
		masterKey2, _ := strategy.GetMasterKey()
		if masterKey != masterKey2 {
			t.Error("GetMasterKey() should return cached key")
		}
	})
}

func TestAgeKeyPairFunctions(t *testing.T) {
	t.Run("generates key pair", func(t *testing.T) {
		identity, err := GenerateAgeKeyPair()
		if err != nil {
			t.Fatalf("GenerateAgeKeyPair() error = %v", err)
		}

		if identity == nil {
			t.Fatal("GenerateAgeKeyPair() returned nil")
		}

		// Verify it's a valid identity
		recipient := identity.Recipient()
		if recipient == nil {
			t.Error("identity should have recipient")
		}
	})

	t.Run("parses identity", func(t *testing.T) {
		identity, _ := GenerateAgeKeyPair()
		identityStr := identity.String()

		parsed, err := ParseAgeIdentity(identityStr)
		if err != nil {
			t.Fatalf("ParseAgeIdentity() error = %v", err)
		}

		if parsed.String() != identityStr {
			t.Error("parsed identity should match original")
		}
	})

	t.Run("rejects invalid identity", func(t *testing.T) {
		_, err := ParseAgeIdentity("invalid-identity")
		if err == nil {
			t.Error("ParseAgeIdentity() should error on invalid input")
		}
	})

}

func TestGetPrivateKey(t *testing.T) {
	t.Run("falls back to environment when file doesn't exist", func(t *testing.T) {
		identity, _ := GenerateAgeKeyPair()
		identityStr := identity.String()

		// Set environment variable
		t.Setenv("OPENENVX_PRIVATE_KEY", identityStr)

		// Use non-existent path
		tmpDir := t.TempDir()
		envPath := tmpDir + "/nonexistent.env"

		got, ok := GetPrivateKey(envPath)
		if !ok {
			t.Error("GetPrivateKey() should fall back to environment variable")
		}

		if got.String() != identityStr {
			t.Errorf("GetPrivateKey() = %v, want %v", got.String(), identityStr)
		}
	})

	t.Run("returns false when neither file nor env var set", func(t *testing.T) {
		t.Setenv("OPENENVX_PRIVATE_KEY", "")

		tmpDir := t.TempDir()
		envPath := tmpDir + "/.env"

		_, ok := GetPrivateKey(envPath)
		if ok {
			t.Error("GetPrivateKey() should return false when neither source available")
		}
	})

	t.Run("works with empty path (env var only)", func(t *testing.T) {
		identity, _ := GenerateAgeKeyPair()
		identityStr := identity.String()

		// Set environment variable
		t.Setenv("OPENENVX_PRIVATE_KEY", identityStr)

		got, ok := GetPrivateKey("")
		if !ok {
			t.Error("GetPrivateKey() should work with empty path")
		}

		if got.String() != identityStr {
			t.Errorf("GetPrivateKey() = %v, want %v", got.String(), identityStr)
		}
	})

	t.Run("retrieves from global store when publicKey set", func(t *testing.T) {
		identity, _ := GenerateAgeKeyPair()
		publicKey := identity.Recipient().String()
		privateKeyStr := identity.String()

		// No env - use global store
		t.Setenv("OPENENVX_PRIVATE_KEY", "")

		tmpDir := t.TempDir()
		t.Setenv("OPENENVX_CONFIG_DIR", tmpDir)
		keysFile, err := config.LoadKeysFile()
		if err != nil {
			t.Fatalf("LoadKeysFile: %v", err)
		}
		if err := keysFile.Set(tmpDir, publicKey, privateKeyStr); err != nil {
			t.Fatalf("keysFile.Set: %v", err)
		}

		// GetPrivateKey with workspace path -> should hit global store
		got, ok := GetPrivateKey(tmpDir)
		if !ok {
			t.Fatal("GetPrivateKey() should find key in global store")
		}
		if got.String() != privateKeyStr {
			t.Errorf("GetPrivateKey() = %v, want %v", got.String(), privateKeyStr)
		}
	})
}
