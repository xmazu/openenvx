package envfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xmazu/openenvx/internal/scanner"
)

func TestDetectCommentedSecrets(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		wantLen         int
		wantKeys        []string
		wantIsSecret    []bool
		skipDetectorNil bool
	}{
		{
			name:            "empty content",
			content:         "",
			wantLen:         0,
			skipDetectorNil: true,
		},
		{
			name:            "no comments",
			content:         "FOO=bar\nBAZ=qux\n",
			wantLen:         0,
			skipDetectorNil: true,
		},
		{
			name:            "comment without assignment",
			content:         "# This is a comment\nFOO=bar\n",
			wantLen:         0,
			skipDetectorNil: true,
		},
		{
			name:            "nil skipDetector returns no results even with commented assignments",
			content:         "# FOO=bar\n# BAZ=qux\n",
			wantLen:         0,
			skipDetectorNil: true,
		},
		{
			name:            "comment with non-secret value",
			content:         "# FOO=bar\nBAZ=qux\n",
			wantLen:         1,
			wantKeys:        []string{"FOO"},
			wantIsSecret:    []bool{true}, // FOO=bar not skipped by any rule → might be secret
			skipDetectorNil: false,
		},
		{
			name:         "commented AWS key",
			content:      "# AWS_KEY=AKIAIOSFODNN7EXAMPLE\nFOO=bar\n",
			wantLen:      1,
			wantKeys:     []string{"AWS_KEY"},
			wantIsSecret: []bool{true},
		},
		{
			name:         "commented GitHub token",
			content:      "# GITHUB_TOKEN=ghp_1234567890abcdefghijklmnopqrstuvwxyz12\n",
			wantLen:      1,
			wantKeys:     []string{"GITHUB_TOKEN"},
			wantIsSecret: []bool{true},
		},
		{
			name:         "commented password",
			content:      "# DB_PASSWORD=supersecret123\n# DB_HOST=localhost\n",
			wantLen:      2,
			wantKeys:     []string{"DB_PASSWORD", "DB_HOST"},
			wantIsSecret: []bool{true, false},
		},
		{
			name:         "multiple commented secrets",
			content:      "# AWS_KEY=AKIAIOSFODNN7EXAMPLE\n# GITHUB_TOKEN=ghp_1234567890abcdefghijklmnopqrstuvwxyz12\n",
			wantLen:      2,
			wantKeys:     []string{"AWS_KEY", "GITHUB_TOKEN"},
			wantIsSecret: []bool{true, true},
		},
		{
			name:         "commented secret with spaces",
			content:      "#  AWS_KEY = AKIAIOSFODNN7EXAMPLE  \n",
			wantLen:      1,
			wantKeys:     []string{"AWS_KEY"},
			wantIsSecret: []bool{true},
		},
		{
			name:         "commented secret with quoted value",
			content:      "# GITHUB_TOKEN=\"ghp_1234567890abcdefghijklmnopqrstuvwxyz12\"\n",
			wantLen:      1,
			wantKeys:     []string{"GITHUB_TOKEN"},
			wantIsSecret: []bool{true},
		},
		{
			name:         "URL without credentials skipped",
			content:      "# API_URL=https://example.com\n",
			wantLen:      1,
			wantKeys:     []string{"API_URL"},
			wantIsSecret: []bool{false},
		},
		{
			name:         "boolean value skipped",
			content:      "# DEBUG=true\n",
			wantLen:      1,
			wantKeys:     []string{"DEBUG"},
			wantIsSecret: []bool{false},
		},
		{
			name:         "number value skipped",
			content:      "# PORT=3000\n",
			wantLen:      1,
			wantKeys:     []string{"PORT"},
			wantIsSecret: []bool{false},
		},
		{
			name:         "localhost value skipped",
			content:      "# DB_HOST=localhost\n",
			wantLen:      1,
			wantKeys:     []string{"DB_HOST"},
			wantIsSecret: []bool{false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			envPath := filepath.Join(tmpDir, ".env")
			if err := os.WriteFile(envPath, []byte(tt.content), 0600); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			f, err := Load(envPath)
			if err != nil {
				t.Fatalf("failed to load file: %v", err)
			}

			var detector *NoEncryptDetector
			if !tt.skipDetectorNil {
				detector = NewNoEncryptDetector()
			}

			got := f.DetectCommentedSecrets(detector, scanner.Patterns)
			if len(got) != tt.wantLen {
				t.Errorf("DetectCommentedSecrets() returned %d secrets, want %d", len(got), tt.wantLen)
				return
			}
			if tt.wantKeys != nil {
				for i, key := range tt.wantKeys {
					if i < len(got) && got[i].Line.Key != key {
						t.Errorf("DetectCommentedSecrets()[%d].Line.Key = %q, want %q", i, got[i].Line.Key, key)
					}
				}
			}
			if tt.wantIsSecret != nil {
				for i, isSecret := range tt.wantIsSecret {
					if i < len(got) && got[i].MightBeSecret != isSecret {
						t.Errorf("DetectCommentedSecrets()[%d].MightBeSecret = %v, want %v", i, got[i].MightBeSecret, isSecret)
					}
				}
			}
		})
	}
}

func TestDetectCommentedSecretsValues(t *testing.T) {
	content := `# AWS_KEY=AKIAIOSFODNN7EXAMPLE
# DB_PASSWORD=supersecret123
FOO=bar
`
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	f, err := Load(envPath)
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	detector := NewNoEncryptDetector()
	got := f.DetectCommentedSecrets(detector, scanner.Patterns)
	if len(got) != 2 {
		t.Fatalf("expected 2 commented secrets, got %d", len(got))
	}

	if got[0].Line.Key != "AWS_KEY" {
		t.Errorf("first key = %q, want AWS_KEY", got[0].Line.Key)
	}
	if got[0].Line.Value != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("first value = %q, want AKIAIOSFODNN7EXAMPLE", got[0].Line.Value)
	}
	if got[0].Line.Num != 1 {
		t.Errorf("first line = %d, want 1", got[0].Line.Num)
	}
	if !got[0].MightBeSecret {
		t.Error("first MightBeSecret should be true")
	}

	if got[1].Line.Key != "DB_PASSWORD" {
		t.Errorf("second key = %q, want DB_PASSWORD", got[1].Line.Key)
	}
	if got[1].Line.Value != "supersecret123" {
		t.Errorf("second value = %q, want supersecret123", got[1].Line.Value)
	}
	if got[1].Line.Num != 2 {
		t.Errorf("second line = %d, want 2", got[1].Line.Num)
	}
	if !got[1].MightBeSecret {
		t.Error("second MightBeSecret should be true (value not skipped by detector)")
	}
}

func TestFilterSecrets(t *testing.T) {
	secrets := []CommentedSecret{
		{MightBeSecret: true},
		{MightBeSecret: false},
		{MightBeSecret: true},
		{MightBeSecret: false},
	}

	filtered := FilterSecrets(secrets)
	if len(filtered) != 2 {
		t.Errorf("FilterSecrets() returned %d secrets, want 2", len(filtered))
	}
	for i, s := range filtered {
		if !s.MightBeSecret {
			t.Errorf("filtered[%d].MightBeSecret = false, want true", i)
		}
	}
}
