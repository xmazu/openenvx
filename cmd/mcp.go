package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/mcpserver"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run the MCP server (stdio) for AI/IDE integration",
	Long:  `Run the Model Context Protocol server on stdio. Exposes create_secret (store secrets), list_keys (key names only), and get_masked (masked value for one key). Never returns plaintext secret values. Run with env via: openenvx run --redact -- <command>.`,
	RunE:  runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) error {
	return mcpserver.Run(context.Background())
}
