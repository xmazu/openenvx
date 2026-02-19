package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/config"
	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/tui"
	"github.com/xmazu/openenvx/internal/workspace"
)

var unmigrateCmd = &cobra.Command{
	Use:   "unmigrate",
	Short: "Decrypt all .env files back to plaintext",
	Long: `Find all encrypted .env files in the workspace and decrypt them
back to plaintext values.

WARNING: This will expose all secret values in plain text. Use with caution.
The .openenvx.yaml file will be preserved for reference, but can be removed manually
if you no longer want to use OpenEnvX for this workspace.`,
	RunE: runUnmigrate,
}

var unmigrateRemoveOpenenvx bool

func init() {
	unmigrateCmd.Flags().BoolVar(&unmigrateRemoveOpenenvx, "remove-openenvx", false, "Remove .openenvx.yaml file after unmigration")
	rootCmd.AddCommand(unmigrateCmd)
}

func runUnmigrate(cmd *cobra.Command, args []string) error {
	wsRoot, err := workspace.FindRoot(".")
	if err != nil {
		return fmt.Errorf("find workspace root: %w", err)
	}

	wsKey, err := workspace.GetWorkspacePublicKey(wsRoot)
	if err != nil {
		return fmt.Errorf("get workspace key: %w", err)
	}
	if wsKey == "" {
		return fmt.Errorf("No workspace key found. Nothing to unmigrate.")
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
	fmt.Fprintln(os.Stderr, tui.Header(fmt.Sprintf("Workspace: %s (%s)", wsRoot, workspace.FormatMarkerForDisplay(marker))))

	masterKey, err := crypto.NewAsymmetricStrategy(identity).GetMasterKey()
	if err != nil {
		return fmt.Errorf("derive master key: %w", err)
	}

	cfg := &workspace.DecryptConfig{
		Envelope: crypto.NewEnvelope(masterKey),
	}

	files, err := workspace.ListEnvFiles(wsRoot)
	if err != nil {
		return fmt.Errorf("list .env files: %w", err)
	}

	var totalDecrypted, totalPlaintext int
	for _, f := range files {
		rel, _ := filepath.Rel(wsRoot, f)

		isEncrypted, err := workspace.IsEnvFileEncrypted(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", tui.Error("✗"), rel, err)
			continue
		}

		if !isEncrypted {
			fmt.Fprintf(os.Stderr, "  %s %s (already plaintext)\n", tui.Muted("•"), rel)
			totalPlaintext++
			continue
		}

		envFile, result := workspace.DecryptEnvFile(f, cfg)
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", tui.Error("✗"), rel, result.Err)
			continue
		}

		if err := envFile.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", tui.Error("✗"), rel, err)
			continue
		}

		fmt.Fprintf(os.Stderr, "  %s %s decrypted\n", tui.Success("✓"), rel)
		totalDecrypted++
	}

	fmt.Fprintf(os.Stderr, "\n%s Decrypted %d .env file(s)\n", tui.Success("Done."), totalDecrypted)
	if totalPlaintext > 0 {
		fmt.Fprintf(os.Stderr, "%s %d file(s) already plaintext (skipped)\n", tui.Muted("•"), totalPlaintext)
	}

	if unmigrateRemoveOpenenvx {
		openenvxPath := filepath.Join(wsRoot, workspace.WorkspaceFileName)
		if err := os.Remove(openenvxPath); err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to remove %s: %v\n", tui.Warning("Warning:"), workspace.WorkspaceFileName, err)
		} else {
			fmt.Fprintf(os.Stderr, "%s Removed %s file\n", tui.Success("✓"), workspace.WorkspaceFileName)
		}
	} else {
		fmt.Fprintf(os.Stderr, "\n%s The %s file was preserved.\n", tui.Muted("Note:"), workspace.WorkspaceFileName)
		fmt.Fprintf(os.Stderr, "%s Run with --remove-openenvx to remove it, or delete it manually.\n", tui.Muted("Tip:"))
	}

	return nil
}
