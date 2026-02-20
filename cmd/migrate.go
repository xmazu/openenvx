package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/config"
	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/tui"
	"github.com/xmazu/openenvx/internal/workspace"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Encrypt plaintext .env files with workspace key",
	Long: `Find all plaintext .env files in the workspace and encrypt them
with the existing workspace key.

Use this when you've added new .env files to an already-initialized workspace.`,
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	wsRoot, err := workspace.FindRoot(".")
	if err != nil {
		return fmt.Errorf("find workspace root: %w", err)
	}

	wsKey, err := workspace.GetWorkspacePublicKey(wsRoot)
	if err != nil {
		return fmt.Errorf("get workspace key: %w", err)
	}
	if wsKey == "" {
		return fmt.Errorf("No workspace key found. Run 'openenvx init' first.")
	}

	keysFile, err := config.LoadKeysFile()
	if err != nil {
		return fmt.Errorf("load keys file: %w", err)
	}

	key, found := keysFile.Get(wsRoot)
	if !found || key.Private == "" {
		return fmt.Errorf("Private key not found for this workspace. Run 'openenvx key add' to add it.")
	}

	identity, err := crypto.ParseAgeIdentity(key.Private)
	if err != nil {
		return fmt.Errorf("parse private key: %w", err)
	}

	marker := workspace.FindMarker(wsRoot)
	fmt.Fprintln(os.Stdout, tui.Header(fmt.Sprintf("Workspace: %s (%s)", wsRoot, workspace.FormatMarkerForDisplay(marker))))

	masterKey, err := crypto.NewAsymmetricStrategy(identity).GetMasterKey()
	if err != nil {
		return fmt.Errorf("derive master key: %w", err)
	}

	cfg := &workspace.EncryptConfig{
		PublicKey:         wsKey,
		NoEncryptDetector: envfile.NewNoEncryptDetector(),
		Envelope:          crypto.NewEnvelope(masterKey),
	}

	files, err := workspace.ListEnvFiles(wsRoot)
	if err != nil {
		return fmt.Errorf("list .env files: %w", err)
	}

	var totalEncrypted, totalSkipped, totalMismatched int
	for _, f := range files {
		rel, _ := filepath.Rel(wsRoot, f)

		isEncrypted, err := workspace.IsEnvFileEncrypted(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", tui.Error("✗"), rel, err)
			continue
		}

		if isEncrypted {
			fmt.Fprintf(os.Stdout, "  %s %s (already encrypted)\n", tui.Muted("•"), rel)
			totalSkipped++
			continue
		}

		envFile, result := workspace.EncryptEnvFile(f, cfg)
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", tui.Error("✗"), rel, result.Err)
			continue
		}

		if err := envFile.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", tui.Error("✗"), rel, err)
			continue
		}

		fmt.Fprintf(os.Stdout, "  %s %s encrypted\n", tui.Success("✓"), rel)
		totalEncrypted++

		printCommentedSecretWarnings(result.CommentedSecrets)
	}

	fmt.Fprintf(os.Stdout, "\n%s Migrated %d .env file(s)\n", tui.Success("Done."), totalEncrypted)
	if totalSkipped > 0 {
		fmt.Fprintf(os.Stdout, "%s %d file(s) already encrypted (skipped)\n", tui.Muted("•"), totalSkipped)
	}
	if totalMismatched > 0 {
		fmt.Fprintf(os.Stderr, "%s %d file(s) have different keys (skipped, manual intervention needed)\n", tui.Warning("Note:"), totalMismatched)
	}

	return nil
}
