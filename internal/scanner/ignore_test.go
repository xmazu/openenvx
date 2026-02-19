package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func writeIgnoreFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestLoadGitignore_NoFile_ReturnsNil(t *testing.T) {
	tmpDir := t.TempDir()

	matcher, err := LoadGitignore(tmpDir)
	if err != nil {
		t.Fatalf("LoadGitignore() error = %v", err)
	}
	if matcher != nil {
		t.Errorf("LoadGitignore() = %v; want nil when no .gitignore", matcher)
	}
}

func TestLoadGitignore_EmptyFile_ReturnsNil(t *testing.T) {
	tmpDir := t.TempDir()
	writeIgnoreFile(t, tmpDir, ".gitignore", "")

	matcher, err := LoadGitignore(tmpDir)
	if err != nil {
		t.Fatalf("LoadGitignore() error = %v", err)
	}
	if matcher != nil {
		t.Error("LoadGitignore() should return nil for empty .gitignore")
	}
}

func TestLoadGitignore_CommentsAndBlanks_Ignored(t *testing.T) {
	tmpDir := t.TempDir()
	writeIgnoreFile(t, tmpDir, ".gitignore", "# comment\n\n   \n# another\n")

	matcher, err := LoadGitignore(tmpDir)
	if err != nil {
		t.Fatalf("LoadGitignore() error = %v", err)
	}
	if matcher != nil {
		t.Fatal("LoadGitignore() should return nil when only comments/blanks")
	}
}

func TestLoadGitignore_RootIsFile_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notadir")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("create file: %v", err)
	}
	_, err := LoadGitignore(filePath)
	if err == nil {
		t.Error("LoadGitignore() error = nil; want error when root is a file")
	}
}

func TestLoadGitignore_WithPatterns_ParsesAndMatches(t *testing.T) {
	tmpDir := t.TempDir()
	writeIgnoreFile(t, tmpDir, ".gitignore", "**/*.key\n**/vendor/\n**/.env\n/dist/\n")

	matcher, err := LoadGitignore(tmpDir)
	if err != nil {
		t.Fatalf("LoadGitignore() error = %v", err)
	}
	if matcher == nil {
		t.Fatal("LoadGitignore() matcher = nil")
	}

	tests := []struct {
		name    string
		relPath string
		isDir   bool
		want    bool
	}{
		{"file matches glob *.key", "secret.key", false, true},
		{"file matches glob in subdir", "sub/secret.key", false, true},
		{"file does not match *.key", "secret.txt", false, false},
		{"dir matches **/vendor/", "vendor", true, true},
		{"dir matches **/vendor/ in subdir", "a/b/vendor", true, true},
		{"file under vendor ignored", "a/vendor/foo.go", false, true},
		{"file matches .env", ".env", false, true},
		{"file matches .env in subdir", "sub/.env", false, true},
		{"root dist/ dir ignored", "dist", true, true},
		{"root dist/ file under ignored", "dist/out", false, true},
		{"subdir dist not anchored", "sub/dist", true, false}, // /dist/ is root-only, so sub/dist is not ignored
		{"dir-only pattern does not ignore file named vendor", "vendor", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matcher.ShouldIgnore(tt.relPath, tt.isDir)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q, %v) = %v; want %v", tt.relPath, tt.isDir, got, tt.want)
			}
		})
	}
}

func TestIgnoreMatcher_ShouldIgnore_EmptyMatcher_ReturnsFalse(t *testing.T) {
	// Empty matcher (no rules) should return false for everything
	matcher := &IgnoreMatcher{rules: nil}
	if matcher.ShouldIgnore("anything", false) {
		t.Error("ShouldIgnore(anything) = true; want false for matcher with no rules")
	}
}

func TestIgnoreMatcher_ShouldIgnore_NilMatcher_ReturnsFalse(t *testing.T) {
	var m *IgnoreMatcher
	if m.ShouldIgnore("any", false) {
		t.Error("nil IgnoreMatcher.ShouldIgnore() = true; want false")
	}
}

func TestIgnoreMatcher_ShouldIgnore_DirOnlyPattern_DoesNotIgnoreFile(t *testing.T) {
	tmpDir := t.TempDir()
	writeIgnoreFile(t, tmpDir, ".gitignore", "build/\n")
	matcher, err := LoadGitignore(tmpDir)
	if err != nil {
		t.Fatalf("LoadGitignore() error = %v", err)
	}

	if matcher.ShouldIgnore("build", false) {
		t.Error("ShouldIgnore(build, false) = true; want false for file when pattern is dir-only")
	}
	if !matcher.ShouldIgnore("build", true) {
		t.Error("ShouldIgnore(build, true) = false; want true for directory")
	}
}

func TestIgnoreMatcher_ShouldIgnore_AnchoredPattern_MatchesOnlyAtRoot(t *testing.T) {
	tmpDir := t.TempDir()
	writeIgnoreFile(t, tmpDir, ".gitignore", "/foo\n")
	matcher, err := LoadGitignore(tmpDir)
	if err != nil {
		t.Fatalf("LoadGitignore() error = %v", err)
	}

	tests := []struct {
		relPath string
		isDir   bool
		want    bool
	}{
		{"foo", false, true},
		{"foo", true, true},
		{"foo/bar", false, true},
		{"sub/foo", false, false},
		{"sub/foo", true, false},
	}
	for _, tt := range tests {
		got := matcher.ShouldIgnore(tt.relPath, tt.isDir)
		if got != tt.want {
			t.Errorf("ShouldIgnore(%q, %v) = %v; want %v", tt.relPath, tt.isDir, got, tt.want)
		}
	}
}

func TestLoadGitignore_UnreadableFile_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Chmod(tmpDir, 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(tmpDir, 0755)
	})

	_, err := LoadGitignore(tmpDir)
	if err == nil {
		t.Error("LoadGitignore() error = nil; want error when file cannot be opened")
	}
}

func TestDefaultScanExcludedFiles(t *testing.T) {
	files := DefaultScanExcludedFiles()
	if len(files) == 0 {
		t.Error("DefaultScanExcludedFiles() returned empty slice")
	}

	// Check that .openenvx.yaml is included
	hasOpenenvx := false
	hasGoMod := false
	for _, f := range files {
		if f == ".openenvx.yaml" {
			hasOpenenvx = true
		}
		if f == "go.mod" {
			hasGoMod = true
		}
	}
	if !hasOpenenvx {
		t.Error("DefaultScanExcludedFiles() should include .openenvx.yaml")
	}
	if !hasGoMod {
		t.Error("DefaultScanExcludedFiles() should include go.mod")
	}
}

func TestFileMatchesExclude(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		patterns []string
		want     bool
	}{
		{"base name exact match", "go.mod", []string{"go.mod"}, true},
		{"base name glob match", "package-lock.json", []string{"*.json"}, true},
		{"base name no match", "main.go", []string{"*.json"}, false},
		{"path with ** glob", "cmd/scan_test.go", []string{"**/*_test.go"}, true},
		{"path with ** glob nested", "internal/scanner/ignore_test.go", []string{"**/*_test.go"}, true},
		{"path with directory glob", "test/main.go", []string{"test/*"}, true},
		{"path with directory glob not matching", "test/sub/main.go", []string{"test/*"}, false},
		{".openenvx.yaml excluded", ".openenvx.yaml", []string{".openenvx.yaml"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileMatchesExclude(tt.relPath, tt.patterns); got != tt.want {
				t.Errorf("FileMatchesExclude(%q, %v) = %v, want %v", tt.relPath, tt.patterns, got, tt.want)
			}
		})
	}
}
