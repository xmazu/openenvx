package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
)

func TestIsEnvFileEncrypted(t *testing.T) {
	t.Run("returns true when any value has envx prefix", func(t *testing.T) {
		tmp := t.TempDir()
		envPath := filepath.Join(tmp, ".env")
		// Mixed: one encrypted, one plain
		content := "API_KEY=envx:abc:def\nFOO=plaintext\n"
		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		got, err := IsEnvFileEncrypted(envPath)
		if err != nil {
			t.Fatalf("IsEnvFileEncrypted: %v", err)
		}
		if !got {
			t.Error("IsEnvFileEncrypted = false, want true (file has envx: value)")
		}
	})

	t.Run("returns false when no value has envx prefix", func(t *testing.T) {
		tmp := t.TempDir()
		envPath := filepath.Join(tmp, ".env")
		if err := os.WriteFile(envPath, []byte("API_KEY=secret123\n"), 0644); err != nil {
			t.Fatal(err)
		}
		got, err := IsEnvFileEncrypted(envPath)
		if err != nil {
			t.Fatalf("IsEnvFileEncrypted: %v", err)
		}
		if got {
			t.Error("IsEnvFileEncrypted = true, want false")
		}
	})

	t.Run("returns false for empty file", func(t *testing.T) {
		tmp := t.TempDir()
		envPath := filepath.Join(tmp, ".env")
		if err := os.WriteFile(envPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		got, err := IsEnvFileEncrypted(envPath)
		if err != nil {
			t.Fatalf("IsEnvFileEncrypted: %v", err)
		}
		if got {
			t.Error("IsEnvFileEncrypted = true, want false")
		}
	})
}

func TestErrEncryptedEnvWithoutOpenenvx(t *testing.T) {
	t.Run("returns error when no .openenvx.yaml and at least one .env encrypted", func(t *testing.T) {
		tmp := t.TempDir()
		envPath := filepath.Join(tmp, ".env")
		if err := os.WriteFile(envPath, []byte("KEY=envx:abc:def\n"), 0644); err != nil {
			t.Fatal(err)
		}
		err := ErrEncryptedEnvWithoutOpenenvx(tmp)
		if err == nil {
			t.Fatal("ErrEncryptedEnvWithoutOpenenvx = nil, want error")
		}
		if !containsSubstring(err.Error(), ".openenvx.yaml") {
			t.Errorf("error should mention .openenvx.yaml: %s", err.Error())
		}
	})

	t.Run("returns nil when .openenvx.yaml exists even if .env encrypted", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmp, ".openenvx.yaml"), []byte("public_key: age1xxx\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("KEY=envx:abc:def\n"), 0644); err != nil {
			t.Fatal(err)
		}
		err := ErrEncryptedEnvWithoutOpenenvx(tmp)
		if err != nil {
			t.Errorf("ErrEncryptedEnvWithoutOpenenvx = %v, want nil", err)
		}
	})

	t.Run("returns nil when no .openenvx.yaml and no encrypted .env", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("KEY=plain\n"), 0644); err != nil {
			t.Fatal(err)
		}
		err := ErrEncryptedEnvWithoutOpenenvx(tmp)
		if err != nil {
			t.Errorf("ErrEncryptedEnvWithoutOpenenvx = %v, want nil", err)
		}
	})
}

func TestEncryptEnvFile(t *testing.T) {
	identity, err := crypto.GenerateAgeKeyPair()
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}

	masterKey, err := crypto.NewAsymmetricStrategy(identity).GetMasterKey()
	if err != nil {
		t.Fatalf("derive master key: %v", err)
	}

	publicKey := identity.Recipient().String()
	envelope := crypto.NewEnvelope(masterKey)

	t.Run("encrypts plaintext file, skips localhost", func(t *testing.T) {
		tmp := t.TempDir()
		envPath := filepath.Join(tmp, ".env")
		if err := os.WriteFile(envPath, []byte("API_KEY=secret123\nDB_HOST=localhost\n"), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &EncryptConfig{
			PublicKey:         publicKey,
			NoEncryptDetector: envfile.NewNoEncryptDetector(),
			Envelope:          envelope,
		}

		envFile, result := EncryptEnvFile(envPath, cfg)

		if result.Err != nil {
			t.Fatalf("EncryptEnvFile error: %v", result.Err)
		}
		if result.AlreadyEncrypted {
			t.Error("AlreadyEncrypted should be false")
		}
		if result.Encrypted != 1 {
			t.Errorf("Encrypted = %d, want 1 (API_KEY)", result.Encrypted)
		}
		if result.Skipped != 1 {
			t.Errorf("Skipped = %d, want 1 (DB_HOST=localhost is skipped)", result.Skipped)
		}

		if err := envFile.Save(); err != nil {
			t.Fatalf("Save error: %v", err)
		}

		content, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatal(err)
		}

		if !containsSubstring(string(content), "envx:") {
			t.Error("file should contain encrypted values")
		}
	})

	t.Run("skips already encrypted file", func(t *testing.T) {
		tmp := t.TempDir()
		envPath := filepath.Join(tmp, ".env")

		encEnv := crypto.NewEnvelope(masterKey)
		encVal, _ := encEnv.Encrypt([]byte("test"), "KEY")
		content := "KEY=" + encVal.String() + "\n"
		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &EncryptConfig{
			PublicKey:         publicKey,
			NoEncryptDetector: envfile.NewNoEncryptDetector(),
			Envelope:          envelope,
		}

		_, result := EncryptEnvFile(envPath, cfg)

		if result.Err != nil {
			t.Fatalf("EncryptEnvFile error: %v", result.Err)
		}
		if !result.AlreadyEncrypted {
			t.Error("AlreadyEncrypted should be true")
		}
		if result.Encrypted != 0 {
			t.Errorf("Encrypted = %d, want 0", result.Encrypted)
		}
	})

	t.Run("detects commented secrets", func(t *testing.T) {
		tmp := t.TempDir()
		envPath := filepath.Join(tmp, ".env")
		content := `# AWS_KEY=AKIAIOSFODNN7EXAMPLE
# DB_PASSWORD=supersecret123
API_KEY=active_key
`
		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &EncryptConfig{
			PublicKey:         publicKey,
			NoEncryptDetector: envfile.NewNoEncryptDetector(),
			Envelope:          envelope,
		}

		envFile, result := EncryptEnvFile(envPath, cfg)

		if result.Err != nil {
			t.Fatalf("EncryptEnvFile error: %v", result.Err)
		}

		if err := envFile.Save(); err != nil {
			t.Fatalf("Save error: %v", err)
		}

		if len(result.CommentedSecrets) != 2 {
			t.Errorf("CommentedSecrets count = %d, want 2", len(result.CommentedSecrets))
		}

		if len(result.CommentedSecrets) > 0 {
			foundAWS := false
			foundPassword := false
			for _, s := range result.CommentedSecrets {
				if s.Line.Key == "AWS_KEY" {
					foundAWS = true
					if !s.MightBeSecret {
						t.Error("AWS_KEY should be flagged as secret")
					}
				}
				if s.Line.Key == "DB_PASSWORD" {
					foundPassword = true
					if !s.MightBeSecret {
						t.Error("DB_PASSWORD should be flagged as secret")
					}
				}
			}
			if !foundAWS {
				t.Error("AWS_KEY not found in commented secrets")
			}
			if !foundPassword {
				t.Error("DB_PASSWORD not found in commented secrets")
			}
		}
	})

	t.Run("skips port numbers, encrypts other values", func(t *testing.T) {
		tmp := t.TempDir()
		envPath := filepath.Join(tmp, ".env")
		if err := os.WriteFile(envPath, []byte("PORT=3000\nAPI_KEY=secret123\n"), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &EncryptConfig{
			PublicKey:         publicKey,
			NoEncryptDetector: envfile.NewNoEncryptDetector(),
			Envelope:          envelope,
		}

		envFile, result := EncryptEnvFile(envPath, cfg)

		if result.Err != nil {
			t.Fatalf("EncryptEnvFile error: %v", result.Err)
		}

		if err := envFile.Save(); err != nil {
			t.Fatalf("Save error: %v", err)
		}

		if result.Encrypted != 1 {
			t.Errorf("Encrypted = %d, want 1 (API_KEY)", result.Encrypted)
		}
		if result.Skipped != 1 {
			t.Errorf("Skipped = %d, want 1 (PORT=3000 is a number)", result.Skipped)
		}
	})

	t.Run("returns empty result for non-existent file", func(t *testing.T) {
		cfg := &EncryptConfig{
			PublicKey:         publicKey,
			NoEncryptDetector: envfile.NewNoEncryptDetector(),
			Envelope:          envelope,
		}

		// envfile.Load creates a new file if it doesn't exist, so no error
		envFile, result := EncryptEnvFile("/nonexistent/.env", cfg)

		if result.Err != nil {
			t.Errorf("unexpected error: %v", result.Err)
		}

		// Should return an empty file (no variables to encrypt)
		if len(envFile.Keys()) != 0 {
			t.Errorf("expected empty file, got %d keys", len(envFile.Keys()))
		}
	})
}

func TestFilterSecrets(t *testing.T) {
	secrets := []envfile.CommentedSecret{
		{Line: &envfile.Line{Key: "AWS_KEY", Value: "AKIAIOSFODNN7EXAMPLE"}, MightBeSecret: true},
		{Line: &envfile.Line{Key: "FOO", Value: "bar"}, MightBeSecret: false},
		{Line: &envfile.Line{Key: "DB_PASSWORD", Value: "secret"}, MightBeSecret: true},
		{Line: &envfile.Line{Key: "URL", Value: "https://example.com"}, MightBeSecret: false},
	}

	filtered := envfile.FilterSecrets(secrets)

	if len(filtered) != 2 {
		t.Errorf("FilterSecrets returned %d secrets, want 2", len(filtered))
	}

	for _, s := range filtered {
		if !s.MightBeSecret {
			t.Errorf("FilterSecrets returned non-secret: %s", s.Line.Key)
		}
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
