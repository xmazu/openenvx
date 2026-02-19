package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xmazu/openenvx/internal/config"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/workspace"
)

func TestRunInit(t *testing.T) {
	t.Run("no .env creates single .env and .openenvx.yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := t.TempDir()
		t.Setenv(config.ConfigDirEnv, configDir)
		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		err := runInit(nil, nil)
		if err != nil {
			t.Fatalf("runInit() error = %v", err)
		}

		envPath := filepath.Join(tmpDir, ".env")
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			t.Fatal(".env should be created")
		}

		workspacePath := filepath.Join(tmpDir, ".openenvx.yaml")
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			t.Fatal(".openenvx.yaml should be created for single file")
		}

		wc, err := workspace.ReadWorkspaceFile(tmpDir)
		if err != nil {
			t.Fatalf("ReadWorkspaceFile() error = %v", err)
		}
		if wc.PublicKey == "" {
			t.Error(".openenvx.yaml should have public key")
		}
	})

	t.Run("single .env file gets encrypted and .openenvx.yaml created", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := t.TempDir()
		t.Setenv(config.ConfigDirEnv, configDir)
		envPath := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envPath, []byte("API_KEY=secret123\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		err := runInit(nil, nil)
		if err != nil {
			t.Fatalf("runInit() error = %v", err)
		}

		workspacePath := filepath.Join(tmpDir, ".openenvx.yaml")
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			t.Fatal(".openenvx.yaml should be created for single file")
		}

		wc, err := workspace.ReadWorkspaceFile(tmpDir)
		if err != nil {
			t.Fatalf("ReadWorkspaceFile() error = %v", err)
		}
		if wc.PublicKey == "" {
			t.Error(".openenvx.yaml should have public key")
		}

		envFile, err := envfile.Load(envPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		val, ok := envFile.Get("API_KEY")
		if !ok {
			t.Fatal("API_KEY not found")
		}
		if !strings.HasPrefix(val, "envx:") {
			t.Error("API_KEY should be encrypted")
		}
	})

	t.Run("multiple .env files creates .openenvx.yaml, no headers", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := t.TempDir()
		t.Setenv(config.ConfigDirEnv, configDir)

		if err := os.MkdirAll(filepath.Join(tmpDir, "apps", "web"), 0755); err != nil {
			t.Fatal(err)
		}

		rootEnv := filepath.Join(tmpDir, ".env")
		webEnv := filepath.Join(tmpDir, "apps", "web", ".env")

		if err := os.WriteFile(rootEnv, []byte("ROOT_KEY=value1\n"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(webEnv, []byte("WEB_KEY=value2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		err := runInit(nil, nil)
		if err != nil {
			t.Fatalf("runInit() error = %v", err)
		}

		workspacePath := filepath.Join(tmpDir, ".openenvx.yaml")
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			t.Fatal(".openenvx.yaml should be created for multiple files")
		}

		wc, err := workspace.ReadWorkspaceFile(tmpDir)
		if err != nil {
			t.Fatalf("ReadWorkspaceFile() error = %v", err)
		}
		if wc.PublicKey == "" {
			t.Error(".openenvx.yaml should have public key")
		}

		rootFile, err := envfile.Load(rootEnv)
		if err != nil {
			t.Fatalf("Load root .env error = %v", err)
		}
		if _, ok := rootFile.Get("ROOT_KEY"); !ok {
			t.Error("root .env should have ROOT_KEY")
		}

		webFile, err := envfile.Load(webEnv)
		if err != nil {
			t.Fatalf("Load web .env error = %v", err)
		}
		if _, ok := webFile.Get("WEB_KEY"); !ok {
			t.Error("web .env should have WEB_KEY")
		}
	})

	t.Run("already initialized returns error", func(t *testing.T) {
		tmpDir := t.TempDir()

		if err := workspace.WriteWorkspaceFile(tmpDir, &workspace.WorkspaceConfig{PublicKey: "age1test"}); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		err := runInit(nil, nil)
		if err == nil {
			t.Error("runInit() should error when already initialized")
		}
		if !strings.Contains(err.Error(), "already initialized") {
			t.Errorf("error should mention 'already initialized', got: %v", err)
		}
	})

	t.Run("warns about commented secrets", func(t *testing.T) {
		tmpDir := t.TempDir()
		configDir := t.TempDir()
		t.Setenv(config.ConfigDirEnv, configDir)
		envPath := filepath.Join(tmpDir, ".env")
		content := `# AWS_KEY=AKIAIOSFODNN7EXAMPLE
# DB_PASSWORD=supersecret123
API_KEY=active_key
`
		if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(origDir)

		err := runInit(nil, nil)
		if err != nil {
			t.Fatalf("runInit() error = %v", err)
		}

		envFile, err := envfile.Load(envPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		val, ok := envFile.Get("API_KEY")
		if !ok {
			t.Fatal("API_KEY not found")
		}
		if !strings.HasPrefix(val, "envx:") {
			t.Error("API_KEY should be encrypted")
		}
	})
}
