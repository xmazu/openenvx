package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/config"
	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/workspace"
)

func TestRunKeyAdd(t *testing.T) {
	t.Run("add from env", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Resolve symlinks to get the real path (macOS /var -> /private/var)
		tmpDir, _ = filepath.EvalSymlinks(tmpDir)
		// Create a marker file so FindRoot can find the workspace
		if err := os.WriteFile(filepath.Join(tmpDir, ".git"), []byte(""), 0644); err != nil {
			t.Fatalf("create marker: %v", err)
		}
		t.Setenv(config.ConfigDirEnv, tmpDir)
		identity, _ := crypto.GenerateAgeKeyPair()
		t.Setenv("OPENENVX_PRIVATE_KEY", identity.String())

		// Change to tmpDir so FindRoot finds it
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		keyAddFile = ""
		keyAddEnv = true

		err := runKeyAdd(nil, nil)
		if err != nil {
			t.Fatalf("runKeyAdd() error = %v", err)
		}

		keysFile, err := config.LoadKeysFile()
		if err != nil {
			t.Fatalf("LoadKeysFile() error = %v", err)
		}
		wk, found := keysFile.Get(tmpDir)
		if !found || wk.Private != identity.String() {
			t.Errorf("key not in store: found=%v priv=%q", found, wk.Private)
		}
	})

	t.Run("add from file with raw key", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Resolve symlinks to get the real path (macOS /var -> /private/var)
		tmpDir, _ = filepath.EvalSymlinks(tmpDir)
		// Create a marker file so FindRoot can find the workspace
		if err := os.WriteFile(filepath.Join(tmpDir, ".git"), []byte(""), 0644); err != nil {
			t.Fatalf("create marker: %v", err)
		}
		t.Setenv(config.ConfigDirEnv, tmpDir)
		identity, _ := crypto.GenerateAgeKeyPair()

		inputFile := filepath.Join(tmpDir, "input.keys")
		if err := os.WriteFile(inputFile, []byte(identity.String()+"\n"), 0600); err != nil {
			t.Fatalf("write input: %v", err)
		}

		// Change to tmpDir so FindRoot finds it
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		keyAddFile = inputFile
		keyAddEnv = false

		err := runKeyAdd(nil, nil)
		if err != nil {
			t.Fatalf("runKeyAdd() error = %v", err)
		}

		keysFile, err := config.LoadKeysFile()
		if err != nil {
			t.Fatalf("LoadKeysFile() error = %v", err)
		}
		_, found := keysFile.Get(tmpDir)
		if !found {
			t.Error("key should be in store after add from file")
		}
	})

	t.Run("errors when no input and env unset", func(t *testing.T) {
		keyAddFile = ""
		keyAddEnv = false
		t.Setenv("OPENENVX_PRIVATE_KEY", "")
		keyAddEnv = true

		err := runKeyAdd(nil, nil)
		if err == nil {
			t.Error("runKeyAdd() should error when OPENENVX_PRIVATE_KEY unset")
		}
	})
}

func TestRunKeyShow(t *testing.T) {
	t.Run("no .env in parents errors", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpDir, _ = filepath.EvalSymlinks(tmpDir)
		t.Setenv(config.ConfigDirEnv, tmpDir)
		keyShowWorkdir = tmpDir
		defer func() { keyShowWorkdir = "" }()

		err := runKeyShow(nil, nil)
		if err == nil {
			t.Error("runKeyShow() should error when no .env found")
		}
	})

	t.Run("prints private key for repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpDir, _ = filepath.EvalSymlinks(tmpDir)
		t.Setenv(config.ConfigDirEnv, tmpDir)
		identity, _ := crypto.GenerateAgeKeyPair()
		pub := identity.Recipient().String()
		keysFile, err := config.LoadKeysFile()
		if err != nil {
			t.Fatalf("LoadKeysFile() error = %v", err)
		}
		if err := keysFile.Set(tmpDir, pub, identity.String()); err != nil {
			t.Fatalf("keysFile.Set: %v", err)
		}

		wc := &workspace.WorkspaceConfig{PublicKey: pub}
		if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
			t.Fatalf("WriteWorkspaceFile: %v", err)
		}

		keyShowWorkdir = tmpDir
		defer func() { keyShowWorkdir = "" }()

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		err = runKeyShow(nil, nil)
		w.Close()
		os.Stdout = old
		if err != nil {
			t.Fatalf("runKeyShow() error = %v", err)
		}
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		got := strings.TrimSpace(buf.String())
		if got != identity.String() {
			t.Errorf("runKeyShow() printed %q, want private key", got)
		}
	})

	t.Run("errors when key not in store", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpDir, _ = filepath.EvalSymlinks(tmpDir)
		t.Setenv(config.ConfigDirEnv, tmpDir)
		identity, _ := crypto.GenerateAgeKeyPair()

		wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
		if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
			t.Fatalf("WriteWorkspaceFile: %v", err)
		}

		keyShowWorkdir = tmpDir
		defer func() { keyShowWorkdir = "" }()

		err := runKeyShow(nil, nil)
		if err == nil {
			t.Error("runKeyShow() should error when private key not in store")
		}
	})
}

func TestParsePrivateKeyFromInput(t *testing.T) {
	t.Run("raw AGE-SECRET-KEY line", func(t *testing.T) {
		identity, _ := crypto.GenerateAgeKeyPair()
		input := identity.String() + "\n"
		got := parsePrivateKeyFromInput(input)
		if got != identity.String() {
			t.Errorf("parsePrivateKeyFromInput() = %q, want %q", got, identity.String())
		}
	})

	t.Run("returns empty for non-key input", func(t *testing.T) {
		got := parsePrivateKeyFromInput("not a key")
		if got != "" {
			t.Errorf("parsePrivateKeyFromInput() = %q, want empty", got)
		}
	})
}
