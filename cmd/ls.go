package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/tui"
	"github.com/xmazu/openenvx/internal/workspace"
)

var lsCmd = &cobra.Command{
	Use:   "ls [directory]",
	Short: "List .env files in a directory tree",
	Long: `Discover and list .env and .env.* files under the given directory.
Auto-detects monorepo workspace and lists all .env files within it.
Output is a simple tree.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLs,
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

func runLs(cmd *cobra.Command, args []string) error {
	var root string

	explicitDir := len(args) == 1

	if !explicitDir {
		wsRoot, err := workspace.FindRoot(".")
		if err != nil {
			return fmt.Errorf("detect workspace: %w", err)
		}

		if workspace.IsWorkspace(wsRoot) {
			root = wsRoot
		} else {
			root = "."
		}
	} else {
		root = args[0]
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve directory: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("directory %s: %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", root)
	}

	isWorkspace := !explicitDir && workspace.IsWorkspace(root)

	var paths []string
	if isWorkspace {
		files, err := workspace.ListEnvFiles(root)
		if err != nil {
			return fmt.Errorf("list .env files: %w", err)
		}
		for _, f := range files {
			rel, _ := filepath.Rel(root, f)
			paths = append(paths, rel)
		}
	} else {
		err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if workspace.IsEnvFilename(d.Name()) {
				rel, _ := filepath.Rel(root, path)
				paths = append(paths, rel)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("walk %s: %w", root, err)
		}
	}
	sort.Strings(paths)

	if len(paths) == 0 {
		return nil
	}

	var stdout io.Writer = os.Stdout
	if cmd != nil {
		stdout = cmd.OutOrStdout()
	}

	if isWorkspace {
		marker := workspace.FindMarker(root)
		fmt.Fprintf(stdout, "%s%s (%s)\n\n", tui.Label("Workspace: "), root, workspace.FormatMarkerForDisplay(marker))
	}

	tree := workspace.BuildEnvTree(paths)
	workspace.PrintEnvTree(tree, "", true)
	return nil
}
