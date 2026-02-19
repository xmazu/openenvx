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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize workspace with encryption",
	Long: `Initialize a new OpenEnvX workspace.

Finds the workspace root (monorepo markers or current directory), generates a new
keypair, and encrypts all .env files with the workspace key.

The public key is stored in .openenvx.yaml at the workspace root.
The private key is stored in ~/.config/openenvx/keys.yaml and should be shared with your
team via a password manager (e.g. 1Password). New teammates run 'openenvx key add'
to add the private key.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	wsRoot, err := workspace.FindRoot(".")
	if err != nil {
		return fmt.Errorf("find workspace root: %w", err)
	}

	if workspace.IsInitialized(wsRoot) {
		marker := workspace.FindMarker(wsRoot)
		markerStr := workspace.FormatMarkerForDisplay(marker)
		return fmt.Errorf("Workspace already initialized at %s (%s).\n\nUse 'openenvx key add' to add your private key.\nUse 'openenvx migrate' to encrypt new .env files.", wsRoot, markerStr)
	}

	if err := workspace.ErrEncryptedEnvWithoutOpenenvx(wsRoot); err != nil {
		return err
	}

	identity, err := crypto.GenerateAgeKeyPair()
	if err != nil {
		return fmt.Errorf("generate keypair: %w", err)
	}
	publicKey := identity.Recipient().String()
	privateKey := identity.String()

	keysFile, err := config.LoadKeysFile()
	if err != nil {
		return fmt.Errorf("load keys file: %w", err)
	}
	if err := keysFile.Set(wsRoot, publicKey, privateKey); err != nil {
		return fmt.Errorf("save private key: %w", err)
	}

	files, err := workspace.ListEnvFiles(wsRoot)
	if err != nil {
		return fmt.Errorf("list .env files: %w", err)
	}

	var createdEmptyEnvPath string
	if len(files) == 0 {
		envPath := filepath.Join(wsRoot, ".env")
		envFile := envfile.New(envPath)
		if err := envFile.Save(); err != nil {
			return fmt.Errorf("create .env: %w", err)
		}
		createdEmptyEnvPath = envPath
		files = []string{envPath}
		fmt.Fprintf(os.Stderr, "%s Created %s\n", tui.Success("✓"), envPath)
	}

	masterKey, err := crypto.NewAsymmetricStrategy(identity).GetMasterKey()
	if err != nil {
		return fmt.Errorf("derive master key: %w", err)
	}

	cfg := &workspace.EncryptConfig{
		PublicKey:         publicKey,
		NoEncryptDetector: envfile.NewNoEncryptDetector(),
		Envelope:          crypto.NewEnvelope(masterKey),
	}

	var totalEncrypted, totalSkipped int
	for _, f := range files {
		rel, _ := filepath.Rel(wsRoot, f)

		if f == createdEmptyEnvPath {
			fmt.Fprintf(os.Stderr, "  %s %s (created)\n", tui.Muted("•"), rel)
			continue
		}

		envFile, result := workspace.EncryptEnvFile(f, cfg)
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", tui.Error("✗"), rel, result.Err)
			continue
		}

		if result.AlreadyEncrypted {
			fmt.Fprintf(os.Stderr, "  %s %s (already encrypted)\n", tui.Muted("•"), rel)
			totalSkipped++
			continue
		}

		if err := envFile.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", tui.Error("✗"), rel, err)
			continue
		}

		fmt.Fprintf(os.Stderr, "  %s %s encrypted\n", tui.Success("✓"), rel)
		totalEncrypted++

		printCommentedSecretWarnings(result.CommentedSecrets)
	}

	if err := workspace.WriteWorkspaceFile(wsRoot, &workspace.WorkspaceConfig{PublicKey: publicKey}); err != nil {
		return fmt.Errorf("write .openenvx.yaml: %w", err)
	}
	fmt.Fprintf(os.Stderr, "  %s Created .openenvx.yaml\n", tui.Success("✓"))

	marker := workspace.FindMarker(wsRoot)
	markerStr := workspace.FormatMarkerForDisplay(marker)
	fmt.Fprintf(os.Stderr, "\n%s Initialized workspace at %s (%s)\n", tui.Success("✓"), wsRoot, markerStr)
	fmt.Fprintf(os.Stderr, "%s %d .env file(s) encrypted", tui.Muted("•"), totalEncrypted)
	if totalSkipped > 0 {
		fmt.Fprintf(os.Stderr, ", %d skipped", totalSkipped)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s Share the private key with your team (e.g. 1Password). New teammates: %s\n", tui.Muted("Tip:"), tui.Label("openenvx key add"))

	if runCmd := workspace.SuggestDevRunCommand(wsRoot); runCmd != "" {
		fmt.Fprintf(os.Stderr, "%s Run your dev server with decrypted env: %s\n", tui.Muted("Tip:"), tui.Label("openenvx run -- "+runCmd))
	}

	return nil
}

func printCommentedSecretWarnings(secrets []envfile.CommentedSecret) {
	filtered := envfile.FilterSecrets(secrets)
	for _, cs := range filtered {
		masked := maskValue(cs.Line.Value)
		fmt.Fprintf(os.Stderr, "    %s Line %d: %s=%s \n", tui.Warning("⚠"), cs.Line.Num, cs.Line.Key, masked)
	}
}

func maskValue(value string) string {
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}
