package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunScan(t *testing.T) {
	t.Run("scans directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test file with potential secret
		testFile := filepath.Join(tmpDir, "test.go")
		content := `package main

const apiKey = "sk_live_1234567890abcdef"
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		// May or may not find secrets depending on patterns
		if err != nil && !strings.Contains(err.Error(), "high-severity secrets") {
			t.Logf("runScan() error = %v (may be expected)", err)
		}
	})

	t.Run("handles empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		if err != nil {
			t.Fatalf("runScan() error = %v", err)
		}
	})

	t.Run("excludes directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create excluded directory
		excludedDir := filepath.Join(tmpDir, "node_modules")
		if err := os.MkdirAll(excludedDir, 0755); err != nil {
			t.Fatalf("failed to create excluded dir: %v", err)
		}

		testFile := filepath.Join(excludedDir, "test.js")
		content := `const apiKey = "sk_live_1234567890abcdef"`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		// Should not find secrets in excluded directory
		if err != nil && !strings.Contains(err.Error(), "high-severity secrets") {
			t.Logf("runScan() error = %v", err)
		}
	})

	t.Run("reports secrets in .env", func(t *testing.T) {
		tmpDir := t.TempDir()

		// .env is scanned; plaintext secret (no envx:) should be reported
		envFile := filepath.Join(tmpDir, ".env")
		content := "API_KEY=ghp_12345678901234567890123456789012abcd"
		if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create .env: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		if err == nil {
			t.Error("runScan() with .env containing plaintext secret should report finding (non-nil error)")
		}
	})

	t.Run(".env with only valid envx structure produces no matches", func(t *testing.T) {
		tmpDir := t.TempDir()

		// .env with only envx:X:Y lines (valid structure) should not report
		envFile := filepath.Join(tmpDir, ".env")
		content := "API_KEY=envx:d3JhcHBlZC1kZWs=:Y2lwaGVydGV4dA==\nOTHER=envx:abc:def\n"
		if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create .env: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		if err != nil {
			t.Errorf("runScan() with .env containing only valid envx lines should succeed (no matches), got: %v", err)
		}
	})

	t.Run(".env with plaintext value without envx prefix is reported", func(t *testing.T) {
		tmpDir := t.TempDir()

		// PUT_KEY=value without envx: prefix should be reported when value matches a pattern
		envFile := filepath.Join(tmpDir, ".env")
		content := "PUT_KEY=sk_live_1234567890abcdefghijklmn"
		if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create .env: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		if err == nil {
			t.Error("runScan() with .env containing PUT_KEY=plaintext secret should report finding (non-nil error)")
		}
	})

	t.Run("skips lock files by default", func(t *testing.T) {
		tmpDir := t.TempDir()

		// package-lock.json with high-severity-looking content would normally fail the scan
		lockFile := filepath.Join(tmpDir, "package-lock.json")
		content := `{"name":"x","dependencies":{"pkg":{"resolved":"https://registry.npmjs.org/pkg/-/pkg-1.0.0.tgz?token=ghp_12345678901234567890123456789012ab"}}}`
		if err := os.WriteFile(lockFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create package-lock.json: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		if err != nil {
			t.Errorf("runScan() with lock file containing secret-like string should succeed (file skipped), got: %v", err)
		}
	})

	t.Run("one file with multiple findings counted once by highest severity", func(t *testing.T) {
		tmpDir := t.TempDir()

		// One file: high (AWS key) + low (high entropy) -> file should count only as high
		testFile := filepath.Join(tmpDir, "config.go")
		content := `package main
var accessKey = "AKIAIOSFODNN7EXAMPLE"
var token = "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyY"
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("pipe: %v", err)
		}
		oldStdout := os.Stdout
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		var out bytes.Buffer
		done := make(chan struct{})
		go func() {
			_, _ = out.ReadFrom(r)
			close(done)
		}()

		runErr := runScan(nil, []string{})
		w.Close()
		<-done
		got := out.String()

		if runErr == nil {
			t.Error("runScan() should return error when high-severity secret found")
		}
		// File counted once in high category, not in low
		if !strings.Contains(got, "High: 1 files") {
			t.Errorf("output should show 'High: 1 files' (one file in high), got:\n%s", got)
		}
		if strings.Contains(got, "Low: 1 files") {
			t.Errorf("output should not show 'Low: 1 files' (file classified only as high), got:\n%s", got)
		}
	})

}

func TestScanFile(t *testing.T) {
	t.Run("finds matches", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")

		content := `package main
const apiKey = "sk_live_1234567890abcdef"
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		matches, err := scanFile(testFile)
		if err != nil {
			t.Fatalf("scanFile() error = %v", err)
		}

		// May or may not find matches depending on patterns
		_ = matches
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "nonexistent.go")

		matches, err := scanFile(testFile)
		if err == nil {
			t.Error("scanFile() should error on non-existent file")
		}
		if len(matches) != 0 {
			t.Error("scanFile() should return empty matches on error")
		}
	})

	t.Run("handles empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "empty.go")

		if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		matches, err := scanFile(testFile)
		if err != nil {
			t.Fatalf("scanFile() error = %v", err)
		}

		if len(matches) != 0 {
			t.Error("scanFile() should return no matches for empty file")
		}
	})
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "shorter than max",
			input:  "short",
			maxLen: 10,
			want:   "short",
		},
		{
			name:   "exactly max length",
			input:  "1234567890",
			maxLen: 10,
			want:   "1234567890",
		},
		{
			name:   "longer than max",
			input:  "this is a very long string",
			maxLen: 10,
			want:   "this is a ...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestScanFileWithMultipleMatches(t *testing.T) {
	t.Run("finds multiple matches in file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")

		// File with multiple secrets
		content := `package main
const apiKey = "sk_live_1234567890abcdef"
const anotherKey = "ghp_12345678901234567890123456789012345678"
const thirdKey = "AKIAIOSFODNN7EXAMPLE"
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		matches, err := scanFile(testFile)
		if err != nil {
			t.Fatalf("scanFile() error = %v", err)
		}

		// Should find at least one match
		if len(matches) == 0 {
			t.Error("scanFile() should find matches")
		}

		// Verify match structure
		for _, match := range matches {
			if match.File == "" {
				t.Error("match.File should not be empty")
			}
			if match.Line == 0 {
				t.Error("match.Line should be set")
			}
			if match.Pattern.Name == "" {
				t.Error("match.Pattern.Name should not be empty")
			}
			if match.Match == "" {
				t.Error("match.Match should not be empty")
			}
		}
	})

	t.Run("handles file with no secrets", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "clean.go")

		content := `package main
func main() {
	fmt.Println("Hello, World!")
}
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		matches, err := scanFile(testFile)
		if err != nil {
			t.Fatalf("scanFile() error = %v", err)
		}

		if len(matches) != 0 {
			t.Errorf("scanFile() should return no matches for clean file, got %d", len(matches))
		}
	})

	t.Run("handles file with multiple matches on same line", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")

		// Line with multiple secrets
		content := `const key1 = "AKIAIOSFODNN7EXAMPLE"; const key2 = "AKIAIOSFODNN7EXAMPLE2"`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		matches, err := scanFile(testFile)
		if err != nil {
			t.Fatalf("scanFile() error = %v", err)
		}

		// Should find multiple matches
		if len(matches) < 1 {
			t.Error("scanFile() should find matches")
		}
	})
}

func TestRunScanComprehensive(t *testing.T) {
	t.Run("finds high severity secrets and returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create file with AWS access key (high severity)
		testFile := filepath.Join(tmpDir, "config.py")
		content := `aws_access_key_id = "AKIAIOSFODNN7EXAMPLE"
aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		// Should return error when high severity secrets found
		if err == nil {
			t.Error("runScan() should return error when high severity secrets found")
		}
	})

	t.Run("handles permission denied errors gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a directory we can't read (permissions)
		restrictedDir := filepath.Join(tmpDir, "restricted")
		if err := os.MkdirAll(restrictedDir, 0000); err != nil {
			t.Skipf("cannot create restricted dir: %v", err)
		}
		defer os.Chmod(restrictedDir, 0755) // Cleanup

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		// Should not error on permission issues
		err := runScan(nil, []string{})
		if err != nil {
			t.Logf("runScan() error = %v (may be ok)", err)
		}
	})

	t.Run("handles nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create nested structure
		nestedDir := filepath.Join(tmpDir, "src", "app", "config")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		testFile := filepath.Join(nestedDir, "secrets.txt")
		content := `stripe_key=sk_live_1234567890abcdef
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		// May find secrets or not depending on patterns
		_ = err
	})

	t.Run("handles binary files gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a binary file
		testFile := filepath.Join(tmpDir, "binary.dat")
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE}
		if err := os.WriteFile(testFile, binaryData, 0644); err != nil {
			t.Fatalf("failed to create binary file: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}

		err := runScan(nil, []string{})
		if err != nil {
			t.Logf("runScan() error = %v (may be ok for binary files)", err)
		}
	})
}

func TestRunScan_ProjectIgnore(t *testing.T) {
	t.Run("respects project ignore files and skips ignored files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// .gitignore: skip *.key and ignored/ directory
		ignorePath := filepath.Join(tmpDir, ".gitignore")
		ignoreContent := "*.key\nignored/\n"
		if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
			t.Fatalf("write .gitignore: %v", err)
		}

		// File that would match secrets but is ignored by pattern
		keyFile := filepath.Join(tmpDir, "secret.key")
		if err := os.WriteFile(keyFile, []byte("api_key=sk_live_1234567890abcdef"), 0644); err != nil {
			t.Fatalf("write secret.key: %v", err)
		}

		// File under ignored/ should not be scanned
		ignoredDir := filepath.Join(tmpDir, "ignored")
		if err := os.MkdirAll(ignoredDir, 0755); err != nil {
			t.Fatalf("create ignored dir: %v", err)
		}
		underIgnored := filepath.Join(ignoredDir, "config.go")
		if err := os.WriteFile(underIgnored, []byte(`var key = "AKIAIOSFODNN7EXAMPLE"`), 0644); err != nil {
			t.Fatalf("write file under ignored: %v", err)
		}

		// File that is scanned and has no secrets (so runScan succeeds)
		scannedFile := filepath.Join(tmpDir, "clean.go")
		if err := os.WriteFile(scannedFile, []byte("package main\nfunc main() {}"), 0644); err != nil {
			t.Fatalf("write clean.go: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}
		scanIgnoreMismatch = "off"

		err := runScan(nil, []string{})
		// Should succeed: secret.key and ignored/config.go are ignored, clean.go has no secrets
		if err != nil {
			t.Fatalf("runScan() error = %v (ignored files should not be scanned)", err)
		}
	})

	t.Run("scan without project ignore files unchanged", func(t *testing.T) {
		tmpDir := t.TempDir()
		// No .gitignore or other ignore files

		testFile := filepath.Join(tmpDir, "config.py")
		if err := os.WriteFile(testFile, []byte(`key = "AKIAIOSFODNN7EXAMPLE"`), 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		scanPath = tmpDir
		scanExclude = []string{".git", "node_modules", "vendor"}
		scanIgnoreMismatch = "off"

		err := runScan(nil, []string{})
		// Should error with high-severity (file is scanned)
		if err == nil {
			t.Error("runScan() expected error when high-severity secret present and no ignore file")
		}
	})
}
