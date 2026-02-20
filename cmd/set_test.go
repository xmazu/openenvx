package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
)

func TestValidateKey(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		if err := ValidateKey("TEST_KEY"); err != nil {
			t.Errorf("ValidateKey(TEST_KEY) error = %v", err)
		}
	})

	t.Run("key with equals is invalid", func(t *testing.T) {
		if err := ValidateKey("INVALID=FORMAT"); err == nil {
			t.Error("ValidateKey(INVALID=FORMAT) should error")
		}
	})
}

func TestSetEnvValue(t *testing.T) {
	t.Run("sets encrypted variable", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.env")

		identity, err := crypto.GenerateAgeKeyPair()
		if err != nil {
			t.Fatalf("GenerateAgeKeyPair() error = %v", err)
		}

		envFile := envfile.New(testFile)
		if err := envFile.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		strategy := crypto.NewAsymmetricStrategy(identity)
		masterKey, err := strategy.GetMasterKey()
		if err != nil {
			t.Fatalf("GetMasterKey() error = %v", err)
		}

		err = SetEnvValue(testFile, "TEST_KEY", "test_value", masterKey)
		if err != nil {
			t.Fatalf("SetEnvValue() error = %v", err)
		}

		loaded, err := envfile.Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		val, ok := loaded.Get("TEST_KEY")
		if !ok {
			t.Fatal("TEST_KEY not found in file")
		}

		if !strings.HasPrefix(val, "envx:") {
			t.Error("value should be encrypted")
		}
	})

	t.Run("creates non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "nonexistent.env")

		identity, err := crypto.GenerateAgeKeyPair()
		if err != nil {
			t.Fatalf("GenerateAgeKeyPair() error = %v", err)
		}

		strategy := crypto.NewAsymmetricStrategy(identity)
		masterKey, err := strategy.GetMasterKey()
		if err != nil {
			t.Fatalf("GetMasterKey() error = %v", err)
		}

		err = SetEnvValue(testFile, "KEY", "testvalue123", masterKey)
		if err != nil {
			t.Fatalf("SetEnvValue() error = %v", err)
		}

		loaded, err := envfile.Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		val, ok := loaded.Get("KEY")
		if !ok {
			t.Fatal("KEY not found in created file")
		}
		if !strings.HasPrefix(val, "envx:") {
			t.Error("value should be encrypted")
		}
	})

	t.Run("updates existing key", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.env")

		identity, err := crypto.GenerateAgeKeyPair()
		if err != nil {
			t.Fatalf("GenerateAgeKeyPair() error = %v", err)
		}

		envFile := envfile.New(testFile)

		strategy := crypto.NewAsymmetricStrategy(identity)
		masterKey, err := strategy.GetMasterKey()
		if err != nil {
			t.Fatalf("GetMasterKey() error = %v", err)
		}

		env := crypto.NewEnvelope(masterKey)
		encrypted, _ := env.Encrypt([]byte("old_value"), "EXISTING_KEY")
		envFile.Set("EXISTING_KEY", encrypted.String())
		if err := envFile.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		err = SetEnvValue(testFile, "EXISTING_KEY", "new_value", masterKey)
		if err != nil {
			t.Fatalf("SetEnvValue() error = %v", err)
		}

		loaded, _ := envfile.Load(testFile)
		val, ok := loaded.Get("EXISTING_KEY")
		if !ok {
			t.Fatal("EXISTING_KEY not found")
		}
		if val == "" {
			t.Error("value should not be empty")
		}
	})
}
