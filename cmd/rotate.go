package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/runenv"
	"github.com/xmazu/openenvx/internal/workspace"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Re-encrypt all secrets with new DEKs",
	Long: `Decrypt all variables with the current master key, re-encrypt each with a new
data encryption key (DEK). Use for deterministic, verifiable rotation; schedule
via cron or CI. Does not change the age key-only rotates DEKs. Auto-detects
monorepo workspace and rotates all .env files when -f is not specified.
With --dry-run, no file is written.`,
	RunE: runRotate,
}

var rotateFile string
var rotateDryRun bool

func init() {
	rotateCmd.Flags().StringVarP(&rotateFile, "file", "f", "", "Path to .env file (auto-detects workspace if empty)")
	rotateCmd.Flags().BoolVar(&rotateDryRun, "dry-run", false, "Do not write; only report what would be done")
	rootCmd.AddCommand(rotateCmd)
}

func runRotate(cmd *cobra.Command, args []string) error {
	if rotateFile != "" {
		return runRotateSingle(cmd)
	}

	wsRoot, err := workspace.FindRoot(".")
	if err != nil {
		return fmt.Errorf("detect workspace: %w", err)
	}

	if workspace.IsWorkspace(wsRoot) {
		return runRotateWorkspace(cmd, wsRoot)
	}

	rotateFile = ".env"
	return runRotateSingle(cmd)
}

func runRotateSingle(cmd *cobra.Command) error {
	envFilePath, err := runenv.ResolveEnvPath(rotateFile, "")
	if err != nil {
		return fmt.Errorf("resolve .env path: %w", err)
	}

	wsRoot, err := workspace.FindRoot(filepath.Dir(envFilePath))
	if err != nil {
		return fmt.Errorf("find workspace: %w", err)
	}

	envFile, err := envfile.Load(envFilePath)
	if err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	masterKey, err := runenv.GetMasterKeyForWorkspace(wsRoot)
	if err != nil {
		return err
	}
	envelope := crypto.NewEnvelope(masterKey)

	decrypted, err := envFile.DecryptAll(envelope)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}
	if len(decrypted) == 0 {
		if rotateDryRun {
			fmt.Fprintln(cmd.OutOrStdout(), "dry-run: no encrypted variables to rotate")
			return nil
		}
		if err := envFile.Save(); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "rotated (no variables to re-encrypt)")
		return nil
	}

	for key, plain := range decrypted {
		encrypted, err := envelope.Encrypt([]byte(plain), key)
		if err != nil {
			return fmt.Errorf("re-encrypt %s: %w", key, err)
		}
		envFile.Set(key, encrypted.String())
	}

	if rotateDryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "dry-run: would re-encrypt %d variable(s)\n", len(decrypted))
		return nil
	}

	if err := envFile.Save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Rotated %d variable(s) at %s\n", len(decrypted), time.Now().UTC().Format(time.RFC3339))
	return nil
}

func runRotateWorkspace(cmd *cobra.Command, wsRoot string) error {
	files, err := workspace.ListEnvFiles(wsRoot)
	if err != nil {
		return fmt.Errorf("list .env files: %w", err)
	}
	if len(files) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No .env files found in workspace")
		return nil
	}

	marker := workspace.FindMarker(wsRoot)
	fmt.Fprintf(cmd.OutOrStdout(), "Workspace: %s (%s)\n\n", wsRoot, workspace.FormatMarkerForDisplay(marker))

	masterKey, err := runenv.GetMasterKeyForWorkspace(wsRoot)
	if err != nil {
		return err
	}
	envelope := crypto.NewEnvelope(masterKey)

	var rotated, skipped, failed int
	for _, envPath := range files {
		relPath, _ := filepath.Rel(wsRoot, envPath)
		fmt.Fprintf(cmd.OutOrStdout(), "Rotating %s...\n", relPath)

		envFile, err := envfile.Load(envPath)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  Error: load: %v\n", err)
			failed++
			continue
		}

		decrypted, err := envFile.DecryptAll(envelope)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  Error: decrypt: %v\n", err)
			failed++
			continue
		}

		if len(decrypted) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  No encrypted variables\n")
			skipped++
			continue
		}

		for key, plain := range decrypted {
			encrypted, err := envelope.Encrypt([]byte(plain), key)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "  Error: re-encrypt %s: %v\n", key, err)
				failed++
				continue
			}
			envFile.Set(key, encrypted.String())
		}

		if rotateDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "  dry-run: would re-encrypt %d variable(s)\n", len(decrypted))
			rotated++
			continue
		}

		if err := envFile.Save(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  Error: save: %v\n", err)
			failed++
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  Rotated %d variable(s)\n", len(decrypted))
		rotated++
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nSummary: %d rotated, %d skipped, %d failed\n", rotated, skipped, failed)
	if failed > 0 {
		return fmt.Errorf("%d file(s) failed", failed)
	}
	return nil
}
