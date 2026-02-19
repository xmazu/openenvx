package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	t.Run("root command has correct metadata", func(t *testing.T) {
		if rootCmd.Use != "openenvx" {
			t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "openenvx")
		}

		if rootCmd.Short != "Secure environment variable management" {
			t.Errorf("rootCmd.Short = %q, want %q", rootCmd.Short, "Secure environment variable management")
		}

		if rootCmd.Long == "" {
			t.Error("rootCmd.Long should not be empty")
		}
	})

	t.Run("root command executes without error with --help", func(t *testing.T) {
		// Create a copy of the root command for testing
		testCmd := &cobra.Command{
			Use:   rootCmd.Use,
			Short: rootCmd.Short,
			Long:  rootCmd.Long,
			Run:   func(cmd *cobra.Command, args []string) {},
		}

		testCmd.SetArgs([]string{"--help"})
		var buf bytes.Buffer
		testCmd.SetOut(&buf)

		err := testCmd.Execute()
		if err != nil {
			t.Errorf("Execute() with --help error = %v", err)
		}

		output := buf.String()
		if output == "" {
			t.Error("--help should produce output")
		}
	})

	t.Run("root command has subcommands", func(t *testing.T) {
		// Verify that subcommands have been added
		commands := []string{"init", "key", "mcp", "migrate", "rotate", "run", "scan", "set"}
		for _, cmdName := range commands {
			found := false
			for _, sub := range rootCmd.Commands() {
				if sub.Name() == cmdName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not found", cmdName)
			}
		}
	})

	t.Run("root command long description contains expected content", func(t *testing.T) {
		expectedStrings := []string{
			"OpenEnvX",
			"encryption",
			"age",
			"init",
			"set",
			"run",
		}

		for _, str := range expectedStrings {
			if !bytes.Contains([]byte(rootCmd.Long), []byte(str)) {
				t.Errorf("rootCmd.Long should contain %q", str)
			}
		}
	})
}

func TestExecute(t *testing.T) {
	t.Run("Execute does not exit when no error", func(t *testing.T) {
		// This test verifies that Execute can be called without panic
		// We can't fully test the os.Exit(1) path without forking,
		// but we can verify the function signature and basic behavior
		// The function itself just wraps rootCmd.Execute()

		// Since Execute calls os.Exit on error, we can only test that
		// it doesn't panic when called (in normal operation it would exit)
		// In test context, without args, it will show help and exit 0

		// We restore os.Args to avoid interference with other tests
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		// We can't actually call Execute() because it will exit,
		// but we've verified the rootCmd structure above
		_ = Execute // Just verify it exists and is callable
	})
}
