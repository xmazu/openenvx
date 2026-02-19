package runenv

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/workspace"
)

func TestFindEnvInParents(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("finds .env in given dir", func(t *testing.T) {
		envPath := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envPath, []byte("KEY=value\n"), 0600); err != nil {
			t.Fatalf("write .env: %v", err)
		}
		got, err := FindEnvInParents(tmpDir, 5)
		if err != nil {
			t.Fatalf("FindEnvInParents() error = %v", err)
		}
		if got != envPath {
			t.Errorf("FindEnvInParents() = %q, want %q", got, envPath)
		}
	})

	t.Run("finds .env in parent", func(t *testing.T) {
		envPath := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envPath, []byte("KEY=value\n"), 0600); err != nil {
			t.Fatalf("write .env: %v", err)
		}
		sub := filepath.Join(tmpDir, "a", "b")
		if err := os.MkdirAll(sub, 0700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		got, err := FindEnvInParents(sub, 5)
		if err != nil {
			t.Fatalf("FindEnvInParents() error = %v", err)
		}
		if got != envPath {
			t.Errorf("FindEnvInParents() = %q, want %q", got, envPath)
		}
	})

	t.Run("returns error when no .env in depth", func(t *testing.T) {
		deep := filepath.Join(tmpDir, "a", "b", "c")
		if err := os.MkdirAll(deep, 0700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		_, err := FindEnvInParents(deep, 2)
		if err == nil {
			t.Error("FindEnvInParents() should error when .env not found within depth")
		}
	})
}

func TestResolveEnvPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		path     string
		workdir  string
		wantDir  string
		wantFile string
	}{
		{
			name:     "empty path uses .env in current dir",
			path:     "",
			workdir:  "",
			wantFile: ".env",
		},
		{
			name:     "path .env with workdir",
			path:     ".env",
			workdir:  tmpDir,
			wantDir:  tmpDir,
			wantFile: ".env",
		},
		{
			name:     "path with subdir and workdir joins",
			path:     "sub/env.env",
			workdir:  tmpDir,
			wantDir:  filepath.Join(tmpDir, "sub"),
			wantFile: "env.env",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveEnvPath(tt.path, tt.workdir)
			if err != nil {
				t.Fatalf("ResolveEnvPath() error = %v", err)
			}
			if tt.wantDir != "" {
				dir := filepath.Dir(got)
				if dir != tt.wantDir {
					t.Errorf("ResolveEnvPath() dir = %v, want %v", dir, tt.wantDir)
				}
			}
			base := filepath.Base(got)
			wantBase := tt.wantFile
			if tt.wantFile != "" && filepath.Base(tt.wantFile) != "" {
				wantBase = filepath.Base(tt.wantFile)
			}
			if base != wantBase && tt.wantFile != "" {
				t.Errorf("ResolveEnvPath() base = %v, want %v", base, wantBase)
			}
			if !filepath.IsAbs(got) {
				t.Errorf("ResolveEnvPath() should return absolute path, got %v", got)
			}
		})
	}
}

func TestGetEnvelopes(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("returns publicKey and encrypted entries for requested keys", func(t *testing.T) {
		path := filepath.Join(tmpDir, ".env")
		makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"API_KEY": "secret1", "DB_URL": "secret2"})
		pub, entries, err := GetEnvelopes(path, []string{"API_KEY", "DB_URL"})
		if err != nil {
			t.Fatalf("GetEnvelopes() error = %v", err)
		}
		if pub == "" || len(pub) < 10 {
			t.Errorf("GetEnvelopes() publicKey = %q, want non-empty age key", pub)
		}
		if len(entries) != 2 {
			t.Errorf("GetEnvelopes() entries len = %d, want 2", len(entries))
		}
		for _, k := range []string{"API_KEY", "DB_URL"} {
			val, ok := entries[k]
			if !ok {
				t.Errorf("GetEnvelopes() missing key %q", k)
				continue
			}
			if _, err := crypto.ParseEncryptedValue(val); err != nil {
				t.Errorf("GetEnvelopes() entry %q not encrypted: %v", k, err)
			}
		}
	})

	t.Run("missing key not in entries", func(t *testing.T) {
		path := filepath.Join(tmpDir, "partial.env")
		makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"A": "a"})
		_, entries, err := GetEnvelopes(path, []string{"A", "MISSING"})
		if err != nil {
			t.Fatalf("GetEnvelopes() error = %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("GetEnvelopes() entries len = %d, want 1", len(entries))
		}
		if _, ok := entries["MISSING"]; ok {
			t.Error("GetEnvelopes() should not include missing key")
		}
	})

	t.Run("non-encrypted value skipped", func(t *testing.T) {
		path := filepath.Join(tmpDir, "mixed.env")
		makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"ENC": "secret"})
		ef, _ := envfile.Load(path)
		ef.Set("PLAIN", "plaintext-value")
		ef.Save()
		_, entries, err := GetEnvelopes(path, nil)
		if err != nil {
			t.Fatalf("GetEnvelopes() error = %v", err)
		}
		if _, ok := entries["PLAIN"]; ok {
			t.Error("GetEnvelopes() should skip non-encrypted value")
		}
		if entries["ENC"] == "" {
			t.Error("GetEnvelopes() should include encrypted ENC")
		}
	})

	t.Run("missing public key returns error", func(t *testing.T) {
		noPubDir := filepath.Join(tmpDir, "nopubdir")
		if err := os.MkdirAll(noPubDir, 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(noPubDir, "nopub.env")
		ef := envfile.New(path)
		ef.Set("KEY", "envx:foo:bar")
		if err := ef.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		_, _, err := GetEnvelopes(path, []string{"KEY"})
		if err == nil {
			t.Error("GetEnvelopes() should error when public key missing")
		}
	})

	t.Run("empty keys returns all encrypted entries", func(t *testing.T) {
		path := filepath.Join(tmpDir, "all.env")
		makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"X": "x", "Y": "y"})
		_, entries, err := GetEnvelopes(path, nil)
		if err != nil {
			t.Fatalf("GetEnvelopes() error = %v", err)
		}
		if len(entries) != 2 {
			t.Errorf("GetEnvelopes(keys=nil) entries len = %d, want 2", len(entries))
		}
	})
}

func makeEncryptedEnvFile(t *testing.T, wsRoot, path string, vars map[string]string) string {
	t.Helper()
	identity, err := crypto.GenerateAgeKeyPair()
	if err != nil {
		t.Fatalf("GenerateAgeKeyPair() error = %v", err)
	}

	wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
	if err := workspace.WriteWorkspaceFile(wsRoot, wc); err != nil {
		t.Fatalf("WriteWorkspaceFile: %v", err)
	}

	ef := envfile.New(path)
	strategy := crypto.NewAsymmetricStrategy(identity)
	masterKey, err := strategy.GetMasterKey()
	if err != nil {
		t.Fatalf("GetMasterKey() error = %v", err)
	}
	env := crypto.NewEnvelope(masterKey)
	for k, v := range vars {
		enc, _ := env.Encrypt([]byte(v), k)
		ef.Set(k, enc.String())
	}
	if err := ef.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	return identity.String()
}

func TestLoadDecryptedEnv(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("missing file returns error", func(t *testing.T) {
		path := filepath.Join(tmpDir, "missing.env")
		_, err := LoadDecryptedEnv(path, tmpDir)
		if err == nil {
			t.Error("LoadDecryptedEnv() should error on missing file")
		}
	})

	t.Run("file without private key returns error", func(t *testing.T) {
		path := filepath.Join(tmpDir, "nokey.env")
		makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"X": "y"})
		os.Unsetenv("OPENENVX_PRIVATE_KEY")
		_, err := LoadDecryptedEnv(path, tmpDir)
		if err == nil {
			t.Error("LoadDecryptedEnv() should error when private key not available")
		}
	})

	t.Run("success with OPENENVX_PRIVATE_KEY", func(t *testing.T) {
		path := filepath.Join(tmpDir, "good.env")
		keyStr := makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"TEST_VAR": "secret-val", "OTHER": "other-val"})
		os.Setenv("OPENENVX_PRIVATE_KEY", keyStr)
		defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

		decrypted, err := LoadDecryptedEnv(path, tmpDir)
		if err != nil {
			t.Fatalf("LoadDecryptedEnv() error = %v", err)
		}
		if decrypted["TEST_VAR"] != "secret-val" {
			t.Errorf("TEST_VAR = %q, want %q", decrypted["TEST_VAR"], "secret-val")
		}
		if decrypted["OTHER"] != "other-val" {
			t.Errorf("OTHER = %q, want %q", decrypted["OTHER"], "other-val")
		}
	})
}

func TestSetSecret(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	keyStr := makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"EXISTING": "old"})
	os.Setenv("OPENENVX_PRIVATE_KEY", keyStr)
	defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

	if err := SetSecret(path, "NEW_KEY", "new-secret-value"); err != nil {
		t.Fatalf("SetSecret() error = %v", err)
	}
	decrypted, err := LoadDecryptedEnv(path, tmpDir)
	if err != nil {
		t.Fatalf("LoadDecryptedEnv() after SetSecret: %v", err)
	}
	if decrypted["EXISTING"] != "old" {
		t.Errorf("EXISTING = %q, want old", decrypted["EXISTING"])
	}
	if decrypted["NEW_KEY"] != "new-secret-value" {
		t.Errorf("NEW_KEY = %q, want new-secret-value", decrypted["NEW_KEY"])
	}
}

func TestRunWithEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows due to command differences")
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "run.env")
	keyStr := makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"ECHO_VAR": "ok"})
	os.Setenv("OPENENVX_PRIVATE_KEY", keyStr)
	defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

	decrypted, err := LoadDecryptedEnv(path, tmpDir)
	if err != nil {
		t.Fatalf("LoadDecryptedEnv() error = %v", err)
	}

	t.Run("runs command with decrypted env and exit 0", func(t *testing.T) {
		exitCode, err := RunWithEnvFromMap(decrypted, "", "sh", []string{"-c", "echo $ECHO_VAR"})
		if err != nil {
			t.Fatalf("RunWithEnvFromMap() error = %v", err)
		}
		if exitCode != 0 {
			t.Errorf("RunWithEnvFromMap() exitCode = %d, want 0", exitCode)
		}
	})

	t.Run("non-zero exit returns exit code and error", func(t *testing.T) {
		exitCode, err := RunWithEnvFromMap(decrypted, "", "sh", []string{"-c", "exit 3"})
		if err == nil {
			t.Error("RunWithEnvFromMap() should return error on non-zero exit")
		}
		if exitCode != 3 {
			t.Errorf("RunWithEnvFromMap() exitCode = %d, want 3", exitCode)
		}
	})
}

func TestMaskSecretValue(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"", ""},
		{"a", "*"},
		{"ab", "**"},
		{"abc", "***"},
		{"abcd", "****"},
		{"abcde", "***de"},
		{"abcdefgh", "******gh"},
		{"abcdefghi", "*****fghi"},
		{"sk_live_abc123xyz", "*************3xyz"},
	}
	for _, tt := range tests {
		got := MaskSecretValue(tt.value)
		if got != tt.want {
			t.Errorf("MaskSecretValue(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestListEncryptedKeys(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	_ = makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"API_KEY": "secret1", "DB_URL": "secret2"})

	keys, err := ListEncryptedKeys(path)
	if err != nil {
		t.Fatalf("ListEncryptedKeys() error = %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("ListEncryptedKeys() len = %d, want 2", len(keys))
	}
	if keys[0] != "API_KEY" || keys[1] != "DB_URL" {
		t.Errorf("ListEncryptedKeys() = %v", keys)
	}
}

func TestGetMaskedSecret(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	keyStr := makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"API_KEY": "sk_live_abc123"})
	os.Setenv("OPENENVX_PRIVATE_KEY", keyStr)
	defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

	masked, length, err := GetMaskedSecret(path, "API_KEY")
	if err != nil {
		t.Fatalf("GetMaskedSecret() error = %v", err)
	}
	if length != 14 {
		t.Errorf("GetMaskedSecret() valueLength = %d, want 14", length)
	}
	if masked != "**********c123" {
		t.Errorf("GetMaskedSecret() masked_value = %q, want **********c123", masked)
	}

	_, _, err = GetMaskedSecret(path, "MISSING")
	if err == nil {
		t.Error("GetMaskedSecret(MISSING) should error")
	}
}

func TestRunWithEnvRedacted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows due to command differences")
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	secret := "my-secret-value"
	keyStr := makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"REDACT_KEY": secret})
	os.Setenv("OPENENVX_PRIVATE_KEY", keyStr)
	defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	decrypted, err := LoadDecryptedEnv(path, tmpDir)
	if err != nil {
		t.Fatalf("LoadDecryptedEnv() error = %v", err)
	}
	exitCode, err := RunWithEnvRedactedFromMap(decrypted, "", "sh", []string{"-c", "echo $REDACT_KEY"})
	_ = wOut.Close()
	_ = wErr.Close()
	var outBuf, errBuf strings.Builder
	_, _ = io.Copy(&outBuf, rOut)
	_, _ = io.Copy(&errBuf, rErr)

	if err != nil {
		t.Fatalf("RunWithEnvRedactedFromMap() error = %v", err)
	}
	if exitCode != 0 {
		t.Errorf("RunWithEnvRedactedFromMap() exitCode = %d, want 0", exitCode)
	}

	out := outBuf.String()
	if !strings.Contains(out, "[REDACTED:REDACT_KEY]") {
		t.Errorf("stdout should contain [REDACTED:REDACT_KEY], got %q", out)
	}
	if strings.Contains(out, secret) {
		t.Errorf("stdout must not contain secret %q", secret)
	}
}

func TestIsDevServerCommand(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{"next", true},
		{"next dev", true},
		{"vite", true},
		{"vite dev", true},
		{"ng serve", true},
		{"vue-cli-service serve", true},
		{"react-scripts start", true},
		{"wrangler dev", true},
		{"serve", true},
		{"nodemon", true},
		{"npm run dev", true},
		{"npm start", true},
		{"yarn dev", true},
		{"pnpm dev", true},
		{"bun dev", true},
		{"echo", false},
		{"ls", false},
		{"cat file.txt", false},
		{"python script.py", false},
		{"go run main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := IsDevServerCommand(tt.command)
			if got != tt.want {
				t.Errorf("IsDevServerCommand(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}

func TestProcessRunner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows due to command differences")
	}

	t.Run("start and wait", func(t *testing.T) {
		runner := &ProcessRunner{
			Command: "echo",
			Args:    []string{"hello"},
			Env:     map[string]string{"TEST_VAR": "test"},
		}

		if err := runner.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if !runner.Running() {
			t.Error("Running() should return true after Start()")
		}

		if err := runner.Wait(); err != nil {
			t.Fatalf("Wait() error = %v", err)
		}

		if runner.Running() {
			t.Error("Running() should return false after Wait()")
		}

		if runner.ExitCode() != 0 {
			t.Errorf("ExitCode() = %d, want 0", runner.ExitCode())
		}
	})

	t.Run("stop kills process", func(t *testing.T) {
		runner := &ProcessRunner{
			Command: "sh",
			Args:    []string{"-c", "sleep 10"},
			Env:     map[string]string{},
		}

		if err := runner.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if !runner.Running() {
			t.Error("Running() should return true after Start()")
		}

		if err := runner.Stop(); err != nil {
			t.Logf("Stop() error = %v (may be expected)", err)
		}

		if runner.Running() {
			t.Error("Running() should return false after Stop()")
		}
	})
}

func TestKeyExists(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	_ = makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"API_KEY": "secret", "DB_URL": "url"})

	t.Run("returns true for existing encrypted key", func(t *testing.T) {
		exists, err := KeyExists(path, "API_KEY")
		if err != nil {
			t.Fatalf("KeyExists() error = %v", err)
		}
		if !exists {
			t.Error("KeyExists(API_KEY) = false, want true")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		exists, err := KeyExists(path, "MISSING_KEY")
		if err != nil {
			t.Fatalf("KeyExists() error = %v", err)
		}
		if exists {
			t.Error("KeyExists(MISSING_KEY) = true, want false")
		}
	})
}

func TestDeleteSecret(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	keyStr := makeEncryptedEnvFile(t, tmpDir, path, map[string]string{"TO_DELETE": "secret", "TO_KEEP": "keep"})
	os.Setenv("OPENENVX_PRIVATE_KEY", keyStr)
	defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

	t.Run("deletes existing key", func(t *testing.T) {
		deleted, err := DeleteSecret(path, "TO_DELETE")
		if err != nil {
			t.Fatalf("DeleteSecret() error = %v", err)
		}
		if !deleted {
			t.Error("DeleteSecret(TO_DELETE) = false, want true")
		}

		decrypted, err := LoadDecryptedEnv(path, tmpDir)
		if err != nil {
			t.Fatalf("LoadDecryptedEnv() error = %v", err)
		}
		if _, ok := decrypted["TO_DELETE"]; ok {
			t.Error("TO_DELETE should be deleted")
		}
		if decrypted["TO_KEEP"] != "keep" {
			t.Error("TO_KEEP should remain")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		deleted, err := DeleteSecret(path, "MISSING")
		if err != nil {
			t.Fatalf("DeleteSecret() error = %v", err)
		}
		if deleted {
			t.Error("DeleteSecret(MISSING) = true, want false")
		}
	})
}

func TestRunWithEnvCaptured(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows due to command differences")
	}

	t.Run("captures stdout and stderr", func(t *testing.T) {
		result, err := RunWithEnvCaptured(map[string]string{}, "", "sh", []string{"-c", "echo stdout; echo stderr >&2"})
		if err != nil {
			t.Fatalf("RunWithEnvCaptured() error = %v", err)
		}
		if !strings.Contains(result.Stdout, "stdout") {
			t.Errorf("stdout = %q, want to contain 'stdout'", result.Stdout)
		}
		if !strings.Contains(result.Stderr, "stderr") {
			t.Errorf("stderr = %q, want to contain 'stderr'", result.Stderr)
		}
		if result.ExitCode != 0 {
			t.Errorf("exitCode = %d, want 0", result.ExitCode)
		}
	})

	t.Run("captures non-zero exit code", func(t *testing.T) {
		result, _ := RunWithEnvCaptured(map[string]string{}, "", "sh", []string{"-c", "exit 5"})
		if result.ExitCode != 5 {
			t.Errorf("exitCode = %d, want 5", result.ExitCode)
		}
	})

	t.Run("injects environment variables", func(t *testing.T) {
		result, err := RunWithEnvCaptured(map[string]string{"MY_VAR": "test123"}, "", "sh", []string{"-c", "echo $MY_VAR"})
		if err != nil {
			t.Fatalf("RunWithEnvCaptured() error = %v", err)
		}
		if !strings.Contains(result.Stdout, "test123") {
			t.Errorf("stdout = %q, want to contain 'test123'", result.Stdout)
		}
	})
}

func TestRunWithEnvCapturedRedacted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows due to command differences")
	}

	secret := "my-secret-value-123"
	result, err := RunWithEnvCapturedRedacted(
		map[string]string{"SECRET_KEY": secret},
		"", "sh", []string{"-c", "echo $SECRET_KEY"},
	)
	if err != nil {
		t.Fatalf("RunWithEnvCapturedRedacted() error = %v", err)
	}
	if strings.Contains(result.Stdout, secret) {
		t.Errorf("stdout should not contain secret, got %q", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "[REDACTED:SECRET_KEY]") {
		t.Errorf("stdout should contain [REDACTED:SECRET_KEY], got %q", result.Stdout)
	}
}

func TestWorkspaceFindRoot(t *testing.T) {
	t.Run("finds workspace root from subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "apps", "web")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		root, err := workspace.FindRoot(subDir)
		if err != nil {
			t.Fatalf("FindRoot() error = %v", err)
		}
		if root != subDir {
			t.Errorf("FindRoot() = %q, want %q (no markers, should return original)", root, subDir)
		}
	})
}
