package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetWorkspacePublicKey(t *testing.T) {
	t.Run("reads from .openenvx.yaml file", func(t *testing.T) {
		tmp := t.TempDir()
		wc := &WorkspaceConfig{PublicKey: "age1abc"}
		if err := WriteWorkspaceFile(tmp, wc); err != nil {
			t.Fatal(err)
		}

		got, err := GetWorkspacePublicKey(tmp)
		if err != nil {
			t.Fatalf("GetWorkspacePublicKey: %v", err)
		}
		if got != "age1abc" {
			t.Errorf("got %q, want %q", got, "age1abc")
		}
	})

	t.Run("returns empty when no key found", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("KEY=value\n"), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := GetWorkspacePublicKey(tmp)
		if err != nil {
			t.Fatalf("GetWorkspacePublicKey: %v", err)
		}
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}
