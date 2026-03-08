package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oexctl",
	Short: "Local-first dev runtime - proxy, services, env; CLI and agent-friendly",
	Long:  `OpenEnvX Runtime: local-first control plane for development. Proxy (*.localhost), services, env; use from the CLI or let an AI agent drive it via MCP.`,
}

func init() {
	rootCmd.AddCommand(proxyCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
