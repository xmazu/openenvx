package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/runenv"
	"github.com/xmazu/openenvx/internal/watch"
	"github.com/xmazu/openenvx/internal/workspace"
)

var runCmd = &cobra.Command{
	Use:   "run -- [command]",
	Short: "Run a command with decrypted environment variables",
	Long: `Decrypt all variables and run the specified command with them as environment variables.
Use -f multiple times to compose .env files (later overrides with --overload).
Use --env KEY=value to add or override.
Use --redact so command output is redacted (secrets replaced with [REDACTED:KEY]).
Auto-detects monorepos and loads all .env files.
Auto-restarts dev servers when .env changes (disable with --no-watch).`,
	RunE: runRun,
}

var runFiles []string
var runOverload bool
var runEnv []string
var runStrict bool
var runQuiet bool
var runRedact bool
var runNoWatch bool

func init() {
	runCmd.Flags().StringSliceVarP(&runFiles, "file", "f", []string{}, "Path(s) to .env file (can be repeated, auto-detected if empty)")
	runCmd.Flags().BoolVar(&runOverload, "overload", false, "Let later files and --env override earlier values")
	runCmd.Flags().StringSliceVarP(&runEnv, "env", "e", nil, "Environment override KEY=value (can be repeated)")
	runCmd.Flags().BoolVar(&runStrict, "strict", false, "Fail if any env file is missing or decryption fails")
	runCmd.Flags().BoolVar(&runQuiet, "quiet", false, "Suppress non-error output")
	runCmd.Flags().BoolVar(&runRedact, "redact", false, "Redact secret values in command output with [REDACTED:KEY] (for agent use)")
	runCmd.Flags().BoolVar(&runNoWatch, "no-watch", false, "Disable auto-restart on .env changes")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified. Use: openenvx run -- your-command")
	}

	files := runFiles
	if len(files) == 0 {
		files = resolveEnvFiles()
	}

	merged, err := runenv.LoadDecryptedEnvFromFiles(files, runOverload, runStrict)
	if err != nil {
		return err
	}
	if err := runenv.MergeOverlayEnv(merged, runEnv, runOverload); err != nil {
		return err
	}

	command := args[0]
	var cmdArgs []string
	if len(args) > 1 {
		cmdArgs = args[1:]
	}

	shouldWatch := !runNoWatch && runenv.IsDevServerCommand(command)

	if !shouldWatch {
		return runOnce(merged, command, cmdArgs)
	}

	return runWithWatch(merged, files, command, cmdArgs)
}

func resolveEnvFiles() []string {
	wsRoot, err := workspace.FindRoot(".")
	if err != nil {
		return []string{".env"}
	}
	if !workspace.IsWorkspace(wsRoot) {
		return []string{".env"}
	}

	envFiles, err := workspace.ListEnvFiles(wsRoot)
	if err != nil || len(envFiles) == 0 {
		return []string{".env"}
	}

	cwd, _ := os.Getwd()
	files := make([]string, 0, len(envFiles))
	for _, f := range envFiles {
		rel, err := filepath.Rel(cwd, f)
		if err != nil {
			files = append(files, f)
		} else {
			files = append(files, rel)
		}
	}
	return files
}

func runOnce(envMap map[string]string, command string, args []string) error {
	var exitCode int
	var err error
	if runRedact {
		exitCode, err = runenv.RunWithEnvRedactedFromMap(envMap, "", command, args)
	} else {
		exitCode, err = runenv.RunWithEnvFromMap(envMap, "", command, args)
	}
	if err != nil {
		if exitCode >= 0 {
			os.Exit(exitCode)
		}
		return err
	}
	return nil
}

func runWithWatch(envMap map[string]string, files []string, command string, args []string) error {
	fw, err := watch.NewFileWatcher()
	if err != nil {
		return fmt.Errorf("create file watcher: %w", err)
	}
	defer fw.Close()

	for _, f := range files {
		absPath, err := filepath.Abs(f)
		if err != nil {
			continue
		}
		if err := fw.Add(absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not watch %s: %v\n", f, err)
		}
	}

	changes := fw.Start()

	runner := &runenv.ProcessRunner{
		Command: command,
		Args:    args,
		Env:     envMap,
		Redact:  runRedact,
	}

	if err := runner.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	for {
		select {
		case sig := <-sigCh:
			if runner.Running() {
				_ = runner.Stop()
			}
			if sig == syscall.SIGTERM {
				os.Exit(143)
			}
			os.Exit(130) // typical exit for SIGINT (Ctrl+C)

		case <-changes:
			fmt.Fprintf(os.Stdout, "⚡ .env changed (%d files watched), restarting...\n", len(fw.Files()))

			if runner.Running() {
				if err := runner.Stop(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: stop error: %v\n", err)
				}
			}

			newEnv, err := runenv.LoadDecryptedEnvFromFiles(files, runOverload, runStrict)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reloading .env: %v\n", err)
				continue
			}
			if err := runenv.MergeOverlayEnv(newEnv, runEnv, runOverload); err != nil {
				fmt.Fprintf(os.Stderr, "Error merging overlay: %v\n", err)
				continue
			}

			runner.Env = newEnv
			if err := runner.Start(); err != nil {
				return fmt.Errorf("restart command: %w", err)
			}

		case err := <-waitProcess(runner):
			if err != nil {
				if runner.ExitCode() >= 0 {
					os.Exit(runner.ExitCode())
				}
				return err
			}
			return nil
		}
	}
}

func waitProcess(runner *runenv.ProcessRunner) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- runner.Wait()
	}()
	return ch
}
