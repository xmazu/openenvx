package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/workspace"
)

func TestRunGet(t *testing.T) {
	t.Run("single key returns raw value by default", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, ".env")
		identityStr, _ := mustSetupEncryptedEnv(t, tmpDir, testFile, map[string]string{"MY_KEY": "my-value"})

		getFile = testFile
		os.Setenv("OPENENVX_PRIVATE_KEY", identityStr)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runGet(nil, []string{"MY_KEY"})
		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = old

		if err != nil {
			t.Fatalf("runGet() error = %v", err)
		}
		if got := strings.TrimSpace(string(out)); got != "my-value" {
			t.Errorf("get MY_KEY = %q, want my-value", got)
		}
	})

	t.Run("single key with format shell", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, ".env")
		identityStr, _ := mustSetupEncryptedEnv(t, tmpDir, testFile, map[string]string{"A": "b"})

		getFile = testFile
		getFormat = "shell"
		defer func() { getFormat = "json" }()
		os.Setenv("OPENENVX_PRIVATE_KEY", identityStr)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runGet(nil, []string{"A"})
		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = old

		if err != nil {
			t.Fatalf("runGet() error = %v", err)
		}
		if got := string(out); got != "A=b" {
			t.Errorf("get A --format shell = %q, want A=b", got)
		}
	})

	t.Run("all keys as JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, ".env")
		identityStr, _ := mustSetupEncryptedEnv(t, tmpDir, testFile, map[string]string{"K1": "v1", "K2": "v2"})

		getFile = testFile
		os.Setenv("OPENENVX_PRIVATE_KEY", identityStr)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runGet(nil, nil)
		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = old

		if err != nil {
			t.Fatalf("runGet() error = %v", err)
		}
		var decoded map[string]string
		if err := json.Unmarshal(out, &decoded); err != nil {
			t.Fatalf("output not valid JSON: %v", err)
		}
		if decoded["K1"] != "v1" || decoded["K2"] != "v2" {
			t.Errorf("get (all) = %v", decoded)
		}
	})

	t.Run("missing key returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, ".env")
		identityStr, _ := mustSetupEncryptedEnv(t, tmpDir, testFile, map[string]string{"X": "y"})

		getFile = testFile
		os.Setenv("OPENENVX_PRIVATE_KEY", identityStr)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		err := runGet(nil, []string{"MISSING"})
		if err == nil {
			t.Error("runGet() should error for missing key")
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		getFile = filepath.Join(tmpDir, "nonexistent.env")
		os.Setenv("OPENENVX_PRIVATE_KEY", "AGE-SECRET-KEY-1QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ")
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		err := runGet(nil, []string{"ANY"})
		if err == nil {
			t.Error("runGet() should error for missing file")
		}
	})
}

func mustSetupEncryptedEnv(t *testing.T, tmpDir, path string, vars map[string]string) (identityStr string, _ *envfile.File) {
	t.Helper()
	identity, err := crypto.GenerateAgeKeyPair()
	if err != nil {
		t.Fatalf("GenerateAgeKeyPair() error = %v", err)
	}

	wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
	if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
		t.Fatalf("WriteWorkspaceFile: %v", err)
	}

	envFile := envfile.New(path)
	strategy := crypto.NewAsymmetricStrategy(identity)
	masterKey, err := strategy.GetMasterKey()
	if err != nil {
		t.Fatalf("GetMasterKey() error = %v", err)
	}
	env := crypto.NewEnvelope(masterKey)
	for k, v := range vars {
		encrypted, err := env.Encrypt([]byte(v), k)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}
		envFile.Set(k, encrypted.String())
	}
	if err := envFile.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	return identity.String(), envFile
}
