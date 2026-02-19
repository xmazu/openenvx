package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/tui"
	"github.com/xmazu/openenvx/internal/workspace"
)

func TestRunSet(t *testing.T) {
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

		wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
		if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
			t.Fatalf("WriteWorkspaceFile() error = %v", err)
		}

		os.Setenv("OPENENVX_PRIVATE_KEY", identity.String())
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		setFile = testFile

		tui.SetMock(&tui.MockPrompts{
			HiddenInputFunc: func(title string) (string, error) {
				return "test_value", nil
			},
		})
		defer tui.ClearMock()

		err = runSet(nil, []string{"TEST_KEY"})
		if err != nil {
			t.Fatalf("runSet() error = %v", err)
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

	t.Run("invalid key with equals", func(t *testing.T) {
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

		wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
		if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
			t.Fatalf("WriteWorkspaceFile() error = %v", err)
		}

		setFile = testFile
		os.Setenv("OPENENVX_PRIVATE_KEY", identity.String())
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		err = runSet(nil, []string{"INVALID=FORMAT"})
		if err == nil {
			t.Error("runSet() should error when key contains =")
		}
	})

	t.Run("creates non-existent file with .openenvx", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "nonexistent.env")

		identity, err := crypto.GenerateAgeKeyPair()
		if err != nil {
			t.Fatalf("GenerateAgeKeyPair() error = %v", err)
		}

		wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
		if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
			t.Fatalf("WriteWorkspaceFile() error = %v", err)
		}

		setFile = testFile
		os.Setenv("OPENENVX_PRIVATE_KEY", identity.String())
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		tui.SetMock(&tui.MockPrompts{
			HiddenInputFunc: func(title string) (string, error) {
				return "testvalue123", nil
			},
		})
		defer tui.ClearMock()

		err = runSet(nil, []string{"KEY"})
		if err != nil {
			t.Fatalf("runSet() error = %v", err)
		}

		// Verify file was created and contains encrypted value
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

		wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
		if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
			t.Fatalf("WriteWorkspaceFile() error = %v", err)
		}

		setFile = testFile
		os.Setenv("OPENENVX_PRIVATE_KEY", identity.String())
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		tui.SetMock(&tui.MockPrompts{
			HiddenInputFunc: func(title string) (string, error) {
				return "new_value", nil
			},
		})
		defer tui.ClearMock()

		err = runSet(nil, []string{"EXISTING_KEY"})
		if err != nil {
			t.Fatalf("runSet() error = %v", err)
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
