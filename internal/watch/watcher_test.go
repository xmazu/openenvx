package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWatcher(t *testing.T) {
	t.Run("detects file changes", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")

		if err := os.WriteFile(envFile, []byte("KEY=value\n"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		w, err := NewFileWatcher()
		if err != nil {
			t.Fatalf("NewFileWatcher: %v", err)
		}
		defer w.Close()

		if err := w.Add(envFile); err != nil {
			t.Fatalf("Add: %v", err)
		}

		changes := w.Start()

		time.Sleep(50 * time.Millisecond)

		if err := os.WriteFile(envFile, []byte("KEY=changed\n"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		select {
		case <-changes:
		case <-time.After(2 * time.Second):
			t.Error("expected change notification")
		}
	})

	t.Run("debounces rapid changes", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")

		if err := os.WriteFile(envFile, []byte("KEY=value\n"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		w, err := NewFileWatcher()
		if err != nil {
			t.Fatalf("NewFileWatcher: %v", err)
		}
		defer w.Close()

		if err := w.Add(envFile); err != nil {
			t.Fatalf("Add: %v", err)
		}

		changes := w.Start()

		time.Sleep(50 * time.Millisecond)

		for i := 0; i < 5; i++ {
			if err := os.WriteFile(envFile, []byte("KEY=value"+string(rune('0'+i))+"\n"), 0644); err != nil {
				t.Fatalf("write file: %v", err)
			}
			time.Sleep(50 * time.Millisecond)
		}

		select {
		case <-changes:
		case <-time.After(2 * time.Second):
			t.Error("expected change notification")
		}

		select {
		case <-changes:
		case <-time.After(600 * time.Millisecond):
		}
	})

	t.Run("watches non-existent file directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")

		w, err := NewFileWatcher()
		if err != nil {
			t.Fatalf("NewFileWatcher: %v", err)
		}
		defer w.Close()

		if err := w.Add(envFile); err != nil {
			t.Fatalf("Add: %v", err)
		}

		changes := w.Start()

		time.Sleep(50 * time.Millisecond)

		if err := os.WriteFile(envFile, []byte("KEY=value\n"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		select {
		case <-changes:
		case <-time.After(2 * time.Second):
			t.Error("expected change notification for created file")
		}
	})

	t.Run("Files returns watched files", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")

		w, err := NewFileWatcher()
		if err != nil {
			t.Fatalf("NewFileWatcher: %v", err)
		}
		defer w.Close()

		if err := w.Add(envFile); err != nil {
			t.Fatalf("Add: %v", err)
		}

		files := w.Files()
		if len(files) != 1 {
			t.Errorf("Files() = %d files, want 1", len(files))
		}
	})
}
