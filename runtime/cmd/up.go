package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xmazu/oexctl/internal/services"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start services defined in .openenvx/services.yaml",
	RunE:  runUp,
}

func init() {
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	mgr, err := services.NewManager(cwd)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := mgr.Start(ctx); err != nil {
		return err
	}

	status, err := mgr.Status(ctx)
	if err != nil {
		return err
	}

	for _, s := range status {
		fmt.Printf("  %s: %s\n", s.Name, s.State)
	}

	return nil
}
