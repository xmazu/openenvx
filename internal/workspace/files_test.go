package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestListEnvFiles(t *testing.T) {
	t.Run("finds all .env files", func(t *testing.T) {
		tmp := t.TempDir()
		dirs := []string{
			tmp,
			filepath.Join(tmp, "apps", "web"),
			filepath.Join(tmp, "packages", "db"),
		}
		for _, d := range dirs {
			if err := os.MkdirAll(d, 0755); err != nil {
				t.Fatal(err)
			}
		}
		files := []string{
			filepath.Join(tmp, ".env"),
			filepath.Join(tmp, "apps", "web", ".env"),
			filepath.Join(tmp, "packages", "db", ".env"),
			filepath.Join(tmp, "packages", "db", ".env.local"),
		}
		for _, f := range files {
			if err := os.WriteFile(f, []byte("KEY=value\n"), 0644); err != nil {
				t.Fatal(err)
			}
		}

		got, err := ListEnvFiles(tmp)
		if err != nil {
			t.Fatalf("ListEnvFiles: %v", err)
		}

		sort.Strings(got)
		want := files
		sort.Strings(want)

		if len(got) != len(want) {
			t.Errorf("got %d files, want %d", len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("file[%d]: got %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("returns empty when no .env files", func(t *testing.T) {
		tmp := t.TempDir()

		got, err := ListEnvFiles(tmp)
		if err != nil {
			t.Fatalf("ListEnvFiles: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d files, want 0", len(got))
		}
	})

	t.Run("skips .env files in excluded dirs", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.MkdirAll(filepath.Join(tmp, "node_modules"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(tmp, "apps", "web", ".turbo"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "node_modules", ".env"), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "apps", "web", ".turbo", ".env"), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ListEnvFiles(tmp)
		if err != nil {
			t.Fatalf("ListEnvFiles: %v", err)
		}
		wantRoot := filepath.Join(tmp, ".env")
		if len(got) != 1 || got[0] != wantRoot {
			t.Errorf("got %v, want only root .env (node_modules/.env and apps/web/.turbo/.env must be skipped)", got)
		}
	})
}
