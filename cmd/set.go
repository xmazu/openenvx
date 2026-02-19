package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/runenv"
	"github.com/xmazu/openenvx/internal/tui"
	"github.com/xmazu/openenvx/internal/workspace"
)

var setCmd = &cobra.Command{
	Use:   "set KEY",
	Short: "Set an environment variable",
	Long: `Store an environment variable in .env file.
Values that don't appear to be secrets (URLs without credentials, booleans, numbers, localhost)
are stored as plaintext; other values are encrypted.`,
	Args: cobra.ExactArgs(1),
	RunE: runSet,
}

var setFile string

func init() {
	setCmd.Flags().StringVarP(&setFile, "file", "f", ".env", "Path to .env file")
	rootCmd.AddCommand(setCmd)
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	if strings.Contains(key, "=") {
		return fmt.Errorf("invalid key: use openenvx set KEY (value is read after the command starts)")
	}

	value, err := readSetValue(key)
	if err != nil {
		return err
	}

	envFile, err := envfile.Load(setFile)
	if err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	detector := envfile.NewNoEncryptDetector()
	if detector.ShouldSkip(key, value) {
		envFile.Set(key, value)
	} else {
		absPath, err := filepath.Abs(setFile)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
		wsRoot, err := workspace.FindRoot(filepath.Dir(absPath))
		if err != nil {
			return fmt.Errorf("find workspace: %w", err)
		}

		masterKey, err := runenv.GetMasterKeyForWorkspace(wsRoot)
		if err != nil {
			return err
		}

		env := crypto.NewEnvelope(masterKey)

		encrypted, err := env.Encrypt([]byte(value), key)
		if err != nil {
			return fmt.Errorf("failed to encrypt: %w", err)
		}

		envFile.Set(key, encrypted.String())
	}

	if err := envFile.Save(); err != nil {
		return fmt.Errorf("failed to save .env file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s %s set\n", tui.Success("✓"), tui.Label(key))
	return nil
}

func readSetValue(key string) (string, error) {
	return tui.HiddenInput(fmt.Sprintf("Value for %s", key))
}
