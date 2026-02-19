package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/workspace"
)

func TestRunRotate(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	identity, err := crypto.GenerateAgeKeyPair()
	if err != nil {
		t.Fatalf("GenerateAgeKeyPair() error = %v", err)
	}

	wc := &workspace.WorkspaceConfig{PublicKey: identity.Recipient().String()}
	if err := workspace.WriteWorkspaceFile(tmpDir, wc); err != nil {
		t.Fatalf("WriteWorkspaceFile() error = %v", err)
	}

	ef := envfile.New(envPath)
	strategy := crypto.NewAsymmetricStrategy(identity)
	masterKey, err := strategy.GetMasterKey()
	if err != nil {
		t.Fatalf("GetMasterKey() error = %v", err)
	}
	env := crypto.NewEnvelope(masterKey)
	enc, _ := env.Encrypt([]byte("secret1"), "KEY1")
	ef.Set("KEY1", enc.String())
	if err := ef.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	os.Setenv("OPENENVX_PRIVATE_KEY", identity.String())
	defer os.Unsetenv("OPENENVX_PRIVATE_KEY")

	prev, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(prev)

	rotateFile = ""
	rotateDryRun = true
	c := &cobra.Command{}
	c.SetOut(os.Stdout)
	if err := runRotate(c, nil); err != nil {
		t.Fatalf("runRotate() dry-run error = %v", err)
	}

	rotateDryRun = false
	if err := runRotate(c, nil); err != nil {
		t.Fatalf("runRotate() error = %v", err)
	}

	loaded, err := envfile.Load(envPath)
	if err != nil {
		t.Fatalf("Load() after rotate: %v", err)
	}
	val, ok := loaded.Get("KEY1")
	if !ok {
		t.Fatal("KEY1 not found")
	}
	if val == "" {
		t.Error("KEY1 should have a value after rotate")
	}
}
