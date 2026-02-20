package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xmazu/openenvx/internal/config"
	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/tui"
	"github.com/xmazu/openenvx/internal/workspace"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage private keys in OpenEnvX global config",
	Long: `Manage private keys stored in the OpenEnvX config directory (~/.config/openenvx/keys.yaml).
Keys are stored by workspace path, so each workspace has its own unique keypair.`,
}

var (
	keyAddFile     string
	keyAddEnv      bool
	keyShowWorkdir string
)

var keyAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a private key to the OpenEnvX config",
	Long: `Add a private key to the global key store for the current workspace.
The key is stored under the workspace path, so openenvx run will find it
when running from this workspace.

Use this when a teammate shared the project key with you (e.g. from 1Password).

Input (one of):
  - interactive: run without flags and paste when prompted (input is hidden)
  - --file path to file containing OPENENVX_PRIVATE_KEY=... or the raw key
  - --env to read from OPENENVX_PRIVATE_KEY environment variable`,
	RunE: runKeyAdd,
}

var keyShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show private key for the current repository",
	Long: `Find the workspace root, then print the matching private key from
the global key store. Use when you need the key for this repo
(e.g. export OPENENVX_PRIVATE_KEY=$(openenvx key show)).`,
	RunE: runKeyShow,
}

func init() {
	rootCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(keyAddCmd)
	keyCmd.AddCommand(keyShowCmd)

	keyAddCmd.Flags().StringVarP(&keyAddFile, "file", "f", "", "Read private key from file")
	keyAddCmd.Flags().BoolVar(&keyAddEnv, "env", false, "Read private key from OPENENVX_PRIVATE_KEY environment variable")
	keyShowCmd.Flags().StringVarP(&keyShowWorkdir, "dir", "C", "", "Directory to search for workspace (default: current)")
}

func runKeyAdd(cmd *cobra.Command, args []string) error {
	var raw string
	var err error

	switch {
	case keyAddEnv:
		keyStr := os.Getenv("OPENENVX_PRIVATE_KEY")
		if keyStr == "" {
			return fmt.Errorf("OPENENVX_PRIVATE_KEY is not set")
		}
		raw = keyStr
	case keyAddFile != "":
		data, err := os.ReadFile(keyAddFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		raw = string(data)
	default:
		raw, err = tui.PlaintextInput("Paste private key")
		if err != nil {
			return fmt.Errorf("failed to read key: %w", err)
		}
	}

	keyStr := parsePrivateKeyFromInput(raw)
	if keyStr == "" {
		return fmt.Errorf("no valid private key found in input")
	}

	identity, err := crypto.ParseAgeIdentity(keyStr)
	if err != nil {
		return fmt.Errorf("invalid age private key: %w", err)
	}

	publicKey := identity.Recipient().String()

	wsRoot, err := workspace.FindRoot(".")
	if err != nil {
		return fmt.Errorf("find workspace root: %w", err)
	}

	if workspace.IsWorkspace(wsRoot) {
		cwd, _ := os.Getwd()
		absCwd, _ := filepath.Abs(cwd)
		if wsRoot != absCwd {
			marker := workspace.FindMarker(wsRoot)
			fmt.Fprintf(os.Stdout, "%s %s (%s)\n", tui.Header("Workspace detected at"), wsRoot, workspace.FormatMarkerForDisplay(marker))
		}

		wsKey, _ := workspace.GetWorkspacePublicKey(wsRoot)
		if wsKey != "" && wsKey != publicKey {
			fmt.Fprintf(os.Stderr, "%s key belongs to %s, but workspace uses %s\n", tui.Warning("Warning:"), tui.FormatKeyDisplay(publicKey), tui.FormatKeyDisplay(wsKey))
		}
	}

	keysFile, err := config.LoadKeysFile()
	if err != nil {
		return fmt.Errorf("load keys file: %w", err)
	}
	if err := keysFile.Set(wsRoot, publicKey, keyStr); err != nil {
		return fmt.Errorf("failed to save key: %w", err)
	}

	fmt.Fprintf(os.Stdout, "%s Key stored for workspace: %s\n", tui.Success("✓"), wsRoot)
	fmt.Fprintf(os.Stdout, "%s Public key: %s\n", tui.Muted("  "), tui.FormatKeyDisplay(publicKey))
	return nil
}

func parsePrivateKeyFromInput(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && strings.HasPrefix(line, "AGE-") {
			return line
		}
	}
	return ""
}

func runKeyShow(cmd *cobra.Command, args []string) error {
	wsRoot, err := workspace.FindRoot(keyShowWorkdir)
	if err != nil {
		return fmt.Errorf("find workspace root: %w", err)
	}

	wsKey, err := workspace.GetWorkspacePublicKey(wsRoot)
	if err != nil {
		return fmt.Errorf("get workspace public key: %w", err)
	}
	if wsKey == "" {
		return fmt.Errorf("workspace at %s has no public key", wsRoot)
	}

	keysFile, err := config.LoadKeysFile()
	if err != nil {
		return fmt.Errorf("load keys file: %w", err)
	}

	key, found := keysFile.Get(wsRoot)
	if !found {
		return fmt.Errorf("private key not found for this workspace - add it with 'openenvx key add'")
	}
	fmt.Println(key.Private)
	return nil
}
