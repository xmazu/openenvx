package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRoot(t *testing.T) {
	t.Run("finds pnpm workspace root", func(t *testing.T) {
		tmp := t.TempDir()
		sub := filepath.Join(tmp, "apps", "web")
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "pnpm-workspace.yaml"), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}

		root, err := FindRoot(sub)
		if err != nil {
			t.Fatalf("FindRoot: %v", err)
		}
		if root != tmp {
			t.Errorf("got %q, want %q", root, tmp)
		}
	})

	t.Run("finds git root as fallback", func(t *testing.T) {
		tmp := t.TempDir()
		sub := filepath.Join(tmp, "src")
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0755); err != nil {
			t.Fatal(err)
		}

		root, err := FindRoot(sub)
		if err != nil {
			t.Fatalf("FindRoot: %v", err)
		}
		if root != tmp {
			t.Errorf("got %q, want %q", root, tmp)
		}
	})

	t.Run("returns current dir when no markers", func(t *testing.T) {
		tmp := t.TempDir()

		root, err := FindRoot(tmp)
		if err != nil {
			t.Fatalf("FindRoot: %v", err)
		}
		if root != tmp {
			t.Errorf("got %q, want %q", root, tmp)
		}
	})

	t.Run("respects marker priority", func(t *testing.T) {
		tmp := t.TempDir()
		nested := filepath.Join(tmp, "packages", "lib")
		if err := os.MkdirAll(nested, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(nested, "pnpm-workspace.yaml"), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
		sub := filepath.Join(nested, "src")
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatal(err)
		}

		root, err := FindRoot(sub)
		if err != nil {
			t.Fatalf("FindRoot: %v", err)
		}
		if root != nested {
			t.Errorf("got %q, want %q (pnpm should beat .git)", root, nested)
		}
	})
}

func TestIsWorkspace(t *testing.T) {
	t.Run("true for pnpm workspace", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmp, "pnpm-workspace.yaml"), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
		if !IsWorkspace(tmp) {
			t.Error("expected IsWorkspace true")
		}
	})

	t.Run("false for non-workspace", func(t *testing.T) {
		tmp := t.TempDir()
		if IsWorkspace(tmp) {
			t.Error("expected IsWorkspace false")
		}
	})
}
