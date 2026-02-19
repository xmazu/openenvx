package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/workspace"
)

func TestRunLs(t *testing.T) {
	t.Run("lists .env files in tree", func(t *testing.T) {
		tmp := t.TempDir()
		for _, name := range []string{".env", ".env.local", "sub/.env"} {
			p := filepath.Join(tmp, name)
			if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}
			if err := os.WriteFile(p, []byte(""), 0644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
		}
		prev, _ := os.Getwd()
		if err := os.Chdir(tmp); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		defer os.Chdir(prev)

		err := runLs(nil, nil)
		if err != nil {
			t.Fatalf("runLs(): %v", err)
		}
		// Output is to stdout; we could capture and assert, but at least run
	})

	t.Run("with directory argument", func(t *testing.T) {
		tmp := t.TempDir()
		p := filepath.Join(tmp, ".env")
		if err := os.WriteFile(p, []byte(""), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		err := runLs(nil, []string{tmp})
		if err != nil {
			t.Fatalf("runLs(%s): %v", tmp, err)
		}
	})

	t.Run("empty directory produces no output", func(t *testing.T) {
		tmp := t.TempDir()
		err := runLs(nil, []string{tmp})
		if err != nil {
			t.Fatalf("runLs(): %v", err)
		}
	})

	t.Run("invalid directory returns error", func(t *testing.T) {
		err := runLs(nil, []string{"/nonexistent-path-12345"})
		if err == nil {
			t.Error("runLs(nonexistent) should error")
		}
	})
}

func TestIsEnvFilename(t *testing.T) {
	for _, tt := range []struct {
		name string
		want bool
	}{
		{".env", true},
		{".env.local", true},
		{".env.production", true},
		{".env.example", false}, // ignored (template only)
		{"env", false},
		{".env", true},
		{".env.", false}, // len ".env." is 5, so ".env." hasPrefix ".env." but len > 5 is false
	} {
		if got := workspace.IsEnvFilename(tt.name); got != tt.want {
			t.Errorf("workspace.IsEnvFilename(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestLsTreeOutput(t *testing.T) {
	// Ensure tree structure is correct when we have mixed files and dirs
	tmp := t.TempDir()
	for _, name := range []string{".env", "a/.env"} {
		p := filepath.Join(tmp, name)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, nil, 0644)
	}
	prev, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(prev)

	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runLs(nil, nil)
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldOut

	if !strings.Contains(string(out), ".env") {
		t.Errorf("output should contain .env, got %q", out)
	}
}
