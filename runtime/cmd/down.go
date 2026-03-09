package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xmazu/oexctl/internal/services"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop services defined in .openenvx/services.yaml",
	RunE:  runDown,
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func runDown(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	mgr, err := services.NewManager(cwd)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return mgr.Stop(ctx)
}
