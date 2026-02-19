package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceConfig(t *testing.T) {
	t.Run("reads and writes workspace file", func(t *testing.T) {
		tmp := t.TempDir()

		wc := &WorkspaceConfig{PublicKey: "age1abc123"}
		if err := WriteWorkspaceFile(tmp, wc); err != nil {
			t.Fatalf("WriteWorkspaceFile: %v", err)
		}

		got, err := ReadWorkspaceFile(tmp)
		if err != nil {
			t.Fatalf("ReadWorkspaceFile: %v", err)
		}
		if got.PublicKey != "age1abc123" {
			t.Errorf("got public key %q, want %q", got.PublicKey, "age1abc123")
		}
	})

	t.Run("returns error when file missing", func(t *testing.T) {
		tmp := t.TempDir()
		_, err := ReadWorkspaceFile(tmp)
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("file permissions", func(t *testing.T) {
		tmp := t.TempDir()
		wc := &WorkspaceConfig{PublicKey: "age1abc"}
		if err := WriteWorkspaceFile(tmp, wc); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(tmp, WorkspaceFileName)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0644 {
			t.Errorf("got permissions %o, want 0644", info.Mode().Perm())
		}
	})
}
