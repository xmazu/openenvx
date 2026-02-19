package envfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/crypto"
)

func deriveMasterKeyFromIdentity(t *testing.T) *crypto.MasterKey {
	t.Helper()
	identity, err := crypto.GenerateAgeKeyPair()
	if err != nil {
		t.Fatalf("GenerateAgeKeyPair() error = %v", err)
	}
	strategy := crypto.NewAsymmetricStrategy(identity)
	mk, err := strategy.GetMasterKey()
	if err != nil {
		t.Fatalf("GetMasterKey() error = %v", err)
	}
	return mk
}

func TestNew(t *testing.T) {
	t.Run("creates new file", func(t *testing.T) {
		f := New("test.env")

		if f.path != "test.env" {
			t.Errorf("path = %v, want %v", f.path, "test.env")
		}

		if len(f.Keys()) != 0 {
			t.Error("new file should have no variables")
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("load existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.env")

		content := `# OpenEnvX encrypted environment file
# Header comment
KEY1=value1
KEY2=value2

# Inline comment
KEY3=value3
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		f, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(f.Keys()) != 3 {
			t.Errorf("loaded %d variables, want 3", len(f.Keys()))
		}

		if val, ok := f.Get("KEY1"); !ok || val != "value1" {
			t.Error("KEY1 not loaded correctly")
		}
	})

	t.Run("load non-existent file creates new", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "nonexistent.env")

		f, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if f == nil {
			t.Fatal("Load() returned nil")
		}

		if len(f.Keys()) != 0 {
			t.Error("new file should have no variables")
		}
	})

	t.Run("load with malformed lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "malformed.env")

		content := `KEY1=value1
malformed line without equals
KEY2=value2
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		f, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(f.Keys()) != 2 {
			t.Errorf("loaded %d variables, want 2", len(f.Keys()))
		}
	})
}

func TestFileGetSet(t *testing.T) {
	t.Run("get existing key", func(t *testing.T) {
		f := New("test.env")
		f.Set("KEY", "value")

		val, ok := f.Get("KEY")
		if !ok {
			t.Error("Get() returned false for existing key")
		}
		if val != "value" {
			t.Errorf("Get() = %v, want %v", val, "value")
		}
	})

	t.Run("get non-existing key", func(t *testing.T) {
		f := New("test.env")

		_, ok := f.Get("NONEXISTENT")
		if ok {
			t.Error("Get() returned true for non-existing key")
		}
	})

	t.Run("set updates value", func(t *testing.T) {
		f := New("test.env")
		f.Set("KEY", "value1")
		f.Set("KEY", "value2")

		val, _ := f.Get("KEY")
		if val != "value2" {
			t.Errorf("Set() did not update value, got %v", val)
		}
	})

	t.Run("set new key", func(t *testing.T) {
		f := New("test.env")
		f.Set("NEWKEY", "newvalue")

		val, ok := f.Get("NEWKEY")
		if !ok || val != "newvalue" {
			t.Error("Set() failed to set new key")
		}
	})
}

func TestFileKeys(t *testing.T) {
	f := New("test.env")
	f.Set("Z_KEY", "z")
	f.Set("A_KEY", "a")
	f.Set("M_KEY", "m")

	keys := f.Keys()

	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}
}

func TestFileSave(t *testing.T) {
	t.Run("save new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "save.env")

		f := New(testFile)
		f.Set("KEY1", "value1")
		f.Set("KEY2", "value2")

		if err := f.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read saved file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "KEY1=value1") {
			t.Error("saved file missing KEY1")
		}
		if !strings.Contains(contentStr, "KEY2=value2") {
			t.Error("saved file missing KEY2")
		}
	})

	t.Run("save preserves order", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "order.env")

		f := New(testFile)
		f.Set("Z_KEY", "z")
		f.Set("A_KEY", "a")
		f.Set("M_KEY", "m")

		if err := f.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		content, _ := os.ReadFile(testFile)
		lines := strings.Split(string(content), "\n")

		var keyLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.Contains(line, "=") {
				keyLines = append(keyLines, line)
			}
		}

		if len(keyLines) != 3 {
			t.Fatalf("expected 3 key lines, got %d", len(keyLines))
		}

		if !strings.HasPrefix(keyLines[0], "Z_KEY") {
			t.Errorf("first key should be Z_KEY (order preserved), got %v", keyLines[0])
		}
	})
}

func TestFileDecryptAll(t *testing.T) {
	t.Run("decrypt all values including plaintext", func(t *testing.T) {
		mk := deriveMasterKeyFromIdentity(t)
		env := crypto.NewEnvelope(mk)

		f := New("test.env")

		encrypted1, _ := env.Encrypt([]byte("secret1"), "KEY1")
		encrypted2, _ := env.Encrypt([]byte("secret2"), "KEY2")

		f.Set("KEY1", encrypted1.String())
		f.Set("KEY2", encrypted2.String())
		f.Set("PLAIN", "plaintext")

		decrypted, err := f.DecryptAll(env)
		if err != nil {
			t.Fatalf("DecryptAll() error = %v", err)
		}

		if len(decrypted) != 3 {
			t.Errorf("DecryptAll() returned %d values, want 3", len(decrypted))
		}

		if decrypted["KEY1"] != "secret1" {
			t.Errorf("KEY1 = %v, want %v", decrypted["KEY1"], "secret1")
		}

		if decrypted["KEY2"] != "secret2" {
			t.Errorf("KEY2 = %v, want %v", decrypted["KEY2"], "secret2")
		}

		if decrypted["PLAIN"] != "plaintext" {
			t.Errorf("PLAIN = %v, want plaintext", decrypted["PLAIN"])
		}
	})

	t.Run("decrypt with wrong key", func(t *testing.T) {
		mk1 := deriveMasterKeyFromIdentity(t)
		mk2 := deriveMasterKeyFromIdentity(t)
		env1 := crypto.NewEnvelope(mk1)
		env2 := crypto.NewEnvelope(mk2)

		f := New("test.env")
		encrypted, _ := env1.Encrypt([]byte("secret"), "KEY")
		f.Set("KEY", encrypted.String())

		_, err := f.DecryptAll(env2)
		if err == nil {
			t.Error("DecryptAll() expected error with wrong key")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		mk := deriveMasterKeyFromIdentity(t)
		env := crypto.NewEnvelope(mk)

		f := New("test.env")
		decrypted, err := f.DecryptAll(env)

		if err != nil {
			t.Errorf("DecryptAll() error = %v", err)
		}

		if len(decrypted) != 0 {
			t.Error("DecryptAll() should return empty map for empty file")
		}
	})
}

func TestFilePath(t *testing.T) {
	f := New("/path/to/file.env")
	if f.path != "/path/to/file.env" {
		t.Errorf("path = %v, want %v", f.path, "/path/to/file.env")
	}
}

func TestRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "roundtrip.env")

	f1 := New(testFile)
	f1.Set("KEY_A", "value_a")
	f1.Set("KEY_B", "value_b")

	if err := f1.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	f2, err := Load(testFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	valA, _ := f2.Get("KEY_A")
	if valA != "value_a" {
		t.Errorf("KEY_A = %v, want %v", valA, "value_a")
	}

	valB, _ := f2.Get("KEY_B")
	if valB != "value_b" {
		t.Errorf("KEY_B = %v, want %v", valB, "value_b")
	}
}

func TestFormatPreservation(t *testing.T) {
	t.Run("preserve order, comments and empty lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "format.env")

		content := `# Header comment
# Another header

DATABASE_URL=postgres://localhost:5432/db

API_SECRET=secret123

# Between comment
DEBUG=true
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		f, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		f.Set("API_SECRET", "envx:newvalue")

		if err := f.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		saved, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read saved file: %v", err)
		}

		savedStr := string(saved)

		if !strings.Contains(savedStr, "# Header comment") {
			t.Error("header comment not preserved")
		}
		if !strings.Contains(savedStr, "# Another header") {
			t.Error("second header comment not preserved")
		}
		if !strings.Contains(savedStr, "# Between comment") {
			t.Error("between comment not preserved")
		}
		if !strings.Contains(savedStr, "DATABASE_URL=postgres://localhost:5432/db") {
			t.Error("DATABASE_URL not preserved")
		}
		if !strings.Contains(savedStr, "API_SECRET=envx:newvalue") {
			t.Error("API_SECRET not updated")
		}
		if !strings.Contains(savedStr, "DEBUG=true") {
			t.Error("DEBUG not preserved")
		}
	})

	t.Run("preserve inline comments", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "inline.env")

		content := `PORT=3000 # dev server
HOST=localhost # local only
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		f, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		val, ok := f.Get("PORT")
		if !ok || val != "3000" {
			t.Errorf("PORT value = %q, want 3000", val)
		}

		lines := f.lines
		if len(lines) > 0 && lines[0].InlineComment != " # dev server" {
			t.Errorf("PORT inline comment = %q, want ' # dev server'", lines[0].InlineComment)
		}

		f.Set("PORT", "4000")

		if err := f.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		saved, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read saved file: %v", err)
		}

		if !strings.Contains(string(saved), "PORT=4000 # dev server") {
			t.Errorf("inline comment not preserved after update, got: %s", string(saved))
		}
	})

	t.Run("preserve variable order", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "order.env")

		content := `ZEBRA=z
ALPHA=a
MIDDLE=m
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		f, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if err := f.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		saved, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read saved file: %v", err)
		}

		savedStr := string(saved)
		zebraIdx := strings.Index(savedStr, "ZEBRA")
		alphaIdx := strings.Index(savedStr, "ALPHA")
		middleIdx := strings.Index(savedStr, "MIDDLE")

		if !(zebraIdx < alphaIdx && alphaIdx < middleIdx) {
			t.Error("variable order not preserved (expected: ZEBRA, ALPHA, MIDDLE)")
		}
	})
}

func TestInlineCommentParsing(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		expectedKey    string
		expectedValue  string
		expectedInline string
	}{
		{"simple value", "KEY=value", "KEY", "value", ""},
		{"value with inline comment", "KEY=value # comment", "KEY", "value", " # comment"},
		{"quoted value with hash", `KEY="val#ue"`, "KEY", "val#ue", ""},
		{"quoted value with comment", `KEY="value" # comment`, "KEY", "value", " # comment"},
		{"single quoted value", `KEY='val#ue'`, "KEY", "val#ue", ""},
		{"value with multiple hashes", "KEY=val#ue#comment", "KEY", "val", "#ue#comment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.env")

			if err := os.WriteFile(testFile, []byte(tt.line+"\n"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			f, err := Load(testFile)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			val, ok := f.Get(tt.expectedKey)
			if !ok {
				t.Fatalf("key %s not found", tt.expectedKey)
			}
			if val != tt.expectedValue {
				t.Errorf("value = %q, want %q", val, tt.expectedValue)
			}

			lines := f.lines
			if len(lines) > 0 && lines[0].InlineComment != tt.expectedInline {
				t.Errorf("inline comment = %q, want %q", lines[0].InlineComment, tt.expectedInline)
			}
		})
	}
}

func TestQuotedValueUnquotedOnParse(t *testing.T) {
	t.Run("load quoted value returns unquoted", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.env")
		if err := os.WriteFile(testFile, []byte(`KEY="ttt"`+"\n"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		f, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		val, ok := f.Get("KEY")
		if !ok {
			t.Fatal("KEY not found")
		}
		if val != "ttt" {
			t.Errorf("Get(KEY) = %q, want ttt", val)
		}
	})

	t.Run("quoted value encrypt and decrypt yields unquoted", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.env")
		if err := os.WriteFile(testFile, []byte(`KEY="ttt"`+"\n"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		f, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		mk := deriveMasterKeyFromIdentity(t)
		env := crypto.NewEnvelope(mk)
		val, _ := f.Get("KEY")
		encrypted, err := env.Encrypt([]byte(val), "KEY")
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}
		f.Set("KEY", encrypted.String())
		if err := f.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		f2, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() after save error = %v", err)
		}
		decrypted, err := f2.DecryptAll(env)
		if err != nil {
			t.Fatalf("DecryptAll() error = %v", err)
		}
		if decrypted["KEY"] != "ttt" {
			t.Errorf("decrypted KEY = %q, want ttt", decrypted["KEY"])
		}
	})
}

func TestFileDelete(t *testing.T) {
	t.Run("delete existing key", func(t *testing.T) {
		f := New("test.env")
		f.Set("KEY1", "value1")
		f.Set("KEY2", "value2")
		f.Set("KEY3", "value3")

		deleted := f.Delete("KEY2")
		if !deleted {
			t.Error("Delete(KEY2) = false, want true")
		}

		if _, ok := f.Get("KEY2"); ok {
			t.Error("KEY2 should be deleted")
		}

		keys := f.Keys()
		if len(keys) != 2 {
			t.Errorf("Keys() len = %d, want 2", len(keys))
		}
	})

	t.Run("delete non-existing key returns false", func(t *testing.T) {
		f := New("test.env")
		f.Set("KEY1", "value1")

		deleted := f.Delete("MISSING")
		if deleted {
			t.Error("Delete(MISSING) = true, want false")
		}
	})

	t.Run("delete preserves other keys", func(t *testing.T) {
		f := New("test.env")
		f.Set("A", "a")
		f.Set("B", "b")
		f.Set("C", "c")

		f.Delete("B")

		valA, _ := f.Get("A")
		valC, _ := f.Get("C")
		if valA != "a" || valC != "c" {
			t.Error("Delete should not affect other keys")
		}
	})

	t.Run("delete and save roundtrip", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "delete.env")

		f := New(testFile)
		f.Set("KEEP", "keep-value")
		f.Set("DELETE", "delete-value")
		if err := f.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		f.Delete("DELETE")
		if err := f.Save(); err != nil {
			t.Fatalf("Save() after delete error = %v", err)
		}

		loaded, err := Load(testFile)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if _, ok := loaded.Get("DELETE"); ok {
			t.Error("DELETE should not exist in loaded file")
		}
		if val, _ := loaded.Get("KEEP"); val != "keep-value" {
			t.Errorf("KEEP = %q, want keep-value", val)
		}
	})
}

func TestCommentedAssignments(t *testing.T) {
	t.Run("parse commented assignments", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.env")

		content := `# AWS_KEY=AKIAIOSFODNN7EXAMPLE
# Regular comment
# DB_PASSWORD=secret123
ACTIVE_KEY=value
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		f, err := Load(testFile)
		if err != nil {
			t.Fatal(err)
		}

		var commented []*Line
		for _, line := range f.lines {
			if line.Type == LineTypeCommentedAssignment {
				commented = append(commented, line)
			}
		}
		if len(commented) != 2 {
			t.Fatalf("expected 2 commented assignments, got %d", len(commented))
		}

		if commented[0].Key != "AWS_KEY" {
			t.Errorf("first key = %q, want AWS_KEY", commented[0].Key)
		}
		if commented[1].Key != "DB_PASSWORD" {
			t.Errorf("second key = %q, want DB_PASSWORD", commented[1].Key)
		}
	})

	t.Run("parse quoted values in comments", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.env")

		content := `# TOKEN="ghp_1234567890abcdefghijklmnopqrstuvwxyz12"
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		f, err := Load(testFile)
		if err != nil {
			t.Fatal(err)
		}

		var commented []*Line
		for _, line := range f.lines {
			if line.Type == LineTypeCommentedAssignment {
				commented = append(commented, line)
			}
		}
		if len(commented) != 1 {
			t.Fatalf("expected 1 commented assignment, got %d", len(commented))
		}

		if commented[0].Value != "ghp_1234567890abcdefghijklmnopqrstuvwxyz12" {
			t.Errorf("value = %q, want unquoted value", commented[0].Value)
		}
	})
}

func TestLineTypes(t *testing.T) {
	t.Run("correct line types", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.env")

		content := `# Comment
# KEY=value
KEY=value

invalid line
`
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		f, err := Load(testFile)
		if err != nil {
			t.Fatal(err)
		}

		lines := f.lines
		if len(lines) != 5 {
			t.Fatalf("expected 5 lines, got %d", len(lines))
		}

		if lines[0].Type != LineTypeComment {
			t.Errorf("line 0 type = %v, want LineTypeComment", lines[0].Type)
		}
		if lines[1].Type != LineTypeCommentedAssignment {
			t.Errorf("line 1 type = %v, want LineTypeCommentedAssignment", lines[1].Type)
		}
		if lines[2].Type != LineTypeVariable {
			t.Errorf("line 2 type = %v, want LineTypeVariable", lines[2].Type)
		}
		if lines[3].Type != LineTypeEmpty {
			t.Errorf("line 3 type = %v, want LineTypeEmpty", lines[3].Type)
		}
		if lines[4].Type != LineTypeInvalid {
			t.Errorf("line 4 type = %v, want LineTypeInvalid", lines[4].Type)
		}
	})
}
