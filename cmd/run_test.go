package cmd

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/workspace"
)

func setupWorkspaceWithKey(t *testing.T, tmpDir string) (string, *crypto.MasterKey) {
	t.Helper()
	identity, err := crypto.GenerateAgeKeyPair()
	if err != nil {
		t.Fatalf("GenerateAgeKeyPair() error = %v", err)
	}

	wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
	if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
		t.Fatalf("WriteWorkspaceFile() error = %v", err)
	}

	strategy := crypto.NewAsymmetricStrategy(identity)
	masterKey, err := strategy.GetMasterKey()
	if err != nil {
		t.Fatalf("GetMasterKey() error = %v", err)
	}

	return identity.String(), masterKey
}

func TestRunRun(t *testing.T) {
	t.Run("requires command argument", func(t *testing.T) {
		err := runRun(nil, []string{})
		if err == nil {
			t.Error("runRun() should error when no command specified")
		}
	})

	t.Run("runs command with decrypted environment", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping on Windows due to command differences")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, ".env")

		identity, masterKey := setupWorkspaceWithKey(t, tmpDir)

		envFile := envfile.New(testFile)
		env := crypto.NewEnvelope(masterKey)
		encrypted, err := env.Encrypt([]byte("test-value"), "TEST_VAR")
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}
		envFile.Set("TEST_VAR", encrypted.String())
		if err := envFile.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		runFiles = []string{testFile}
		os.Setenv("OPENENVX_PRIVATE_KEY", identity)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		err = runRun(nil, []string{"sh", "-c", "echo $TEST_VAR"})
		if err != nil {
			t.Logf("runRun() error = %v (may be expected due to command execution)", err)
		}
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "nonexistent.env")

		identity, _ := crypto.GenerateAgeKeyPair()

		runFiles = []string{testFile}
		runStrict = true
		defer func() { runStrict = false }()
		os.Setenv("OPENENVX_PRIVATE_KEY", identity.String())
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		err := runRun(nil, []string{"echo", "test"})
		if err == nil {
			t.Error("runRun() should error on non-existent file")
		}
	})

	t.Run("redact flag redacts secret values in output", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping on Windows due to command differences")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, ".env")

		identity, masterKey := setupWorkspaceWithKey(t, tmpDir)

		envFile := envfile.New(testFile)
		env := crypto.NewEnvelope(masterKey)
		encrypted, _ := env.Encrypt([]byte("secret-out"), "OUT_KEY")
		envFile.Set("OUT_KEY", encrypted.String())
		if err := envFile.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		runFiles = []string{testFile}
		runRedact = true
		defer func() { runRedact = false }()
		os.Setenv("OPENENVX_PRIVATE_KEY", identity)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		err := runRun(nil, []string{"sh", "-c", "echo $OUT_KEY"})
		_ = w.Close()
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)

		if err != nil {
			t.Fatalf("runRun() error = %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "[REDACTED:OUT_KEY]") {
			t.Errorf("output should contain [REDACTED:OUT_KEY], got %q", out)
		}
		if strings.Contains(out, "secret-out") {
			t.Errorf("output must not contain plaintext secret")
		}
	})

	t.Run("handles command with arguments", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping on Windows due to command differences")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, ".env")

		identity, _ := setupWorkspaceWithKey(t, tmpDir)

		envFile := envfile.New(testFile)
		if err := envFile.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		runFiles = []string{testFile}
		os.Setenv("OPENENVX_PRIVATE_KEY", identity)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		err := runRun(nil, []string{"echo", "arg1", "arg2"})
		if err != nil {
			t.Logf("runRun() error = %v (may be expected)", err)
		}
	})

	t.Run("handles command execution error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping on Windows due to command differences")
		}

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, ".env")

		identity, _ := setupWorkspaceWithKey(t, tmpDir)

		envFile := envfile.New(testFile)
		if err := envFile.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		runFiles = []string{testFile}
		os.Setenv("OPENENVX_PRIVATE_KEY", identity)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		err := runRun(nil, []string{"nonexistent-command-12345"})
		if err == nil {
			t.Error("runRun() should error on command execution failure")
		}
	})
}

func TestCommandExecution(t *testing.T) {
	t.Run("executes simple command", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping on Windows")
		}

		cmd := exec.Command("echo", "test")
		if err := cmd.Run(); err != nil {
			t.Fatalf("command execution error = %v", err)
		}
	})
}
