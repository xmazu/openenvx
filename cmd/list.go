package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/runenv"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List encrypted environment variable names",
	Long: `List the names of all encrypted environment variables in the .env file.
This shows key names only - no values are displayed. Use this to see what
secrets exist without decrypting them.

Examples:
  openenvx list                    # List keys in .env
  openenvx list --file .env.local  # List keys in specific file`,
	RunE: runList,
}

var listFile string

func init() {
	listCmd.Flags().StringVarP(&listFile, "file", "f", ".env", "Path to .env file")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	envFilePath, err := runenv.ResolveEnvPath(listFile, "")
	if err != nil {
		return fmt.Errorf("resolve .env path: %w", err)
	}

	keys, err := runenv.ListEncryptedKeys(envFilePath)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		fmt.Println("No encrypted environment variables found.")
		return nil
	}

	output := map[string]interface{}{
		"keys":  keys,
		"path":  envFilePath,
		"count": len(keys),
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(output)
}
