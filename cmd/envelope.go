package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/audit"
	"github.com/xmazu/openenvx/internal/envelope"
	"github.com/xmazu/openenvx/internal/runenv"
	"github.com/xmazu/openenvx/internal/workspace"
)

var envelopeCmd = &cobra.Command{
	Use:   "envelope",
	Short: "Manage cryptographic envelopes for agents",
	Long: `Create and use self-contained envelopes for agent workflows.

An envelope is a sealed package containing encrypted secrets that can be
unwrapped without the master private key. Security comes from short TTL
and minimal scope.

Examples:
  openenvx envelope create --scope=DATABASE_URL,API_KEY --ttl=1h
  openenvx envelope run envelope:v1:... -- npm test
  openenvx envelope inspect envelope:v1:...`,
}

var envelopeCreateCmd = &cobra.Command{
	Use:   "create --scope=KEY1,KEY2 --ttl=1h",
	Short: "Create a self-contained envelope",
	Long: `Create an envelope containing the specified secrets.

The envelope is self-contained and can be passed to agents.
They can decrypt and use the secrets without needing the master private key.

Security: Keep TTL short, scope minimal. Envelope possession = access.`,
	RunE: runEnvelopeCreate,
}

var envelopeRunCmd = &cobra.Command{
	Use:   "run <envelope> -- <command>",
	Short: "Run a command with envelope secrets",
	Long: `Unwrap the envelope and run a command with secrets injected.

The envelope is decrypted, secrets are injected into the environment,
and the command is executed. Output can be redacted for safety.`,
	RunE: runEnvelopeRun,
}

var envelopeInspectCmd = &cobra.Command{
	Use:   "inspect <envelope>",
	Short: "Inspect envelope metadata without decrypting",
	Long: `Show envelope metadata: scope, expiration, status.

Does not decrypt or expose secret values.`,
	RunE: runEnvelopeInspect,
}

var envelopeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent envelopes (from audit log)",
	Long: `Show recent envelope operations from the audit log.

Lists envelope_create and envelope_run entries.`,
	RunE: runEnvelopeList,
}

var (
	envelopeScope      []string
	envelopeTTL        string
	envelopeOutputFile string
	envelopeRedact     bool
	envelopeWorkdir    string
)

func init() {
	envelopeCreateCmd.Flags().StringSliceVar(&envelopeScope, "scope", nil, "Keys to include (required, comma-separated)")
	envelopeCreateCmd.Flags().StringVar(&envelopeTTL, "ttl", "1h", "Time-to-live (e.g., 30m, 1h, 24h)")
	envelopeCreateCmd.Flags().StringVar(&envelopeOutputFile, "output-file", "", "Write envelope to file instead of stdout")
	envelopeCreateCmd.Flags().StringVarP(&envelopeWorkdir, "workdir", "w", "", "Working directory (default: current)")
	envelopeCreateCmd.MarkFlagRequired("scope")

	envelopeRunCmd.Flags().BoolVar(&envelopeRedact, "redact", false, "Redact secret values in output")
	envelopeRunCmd.Flags().StringVarP(&envelopeWorkdir, "workdir", "w", "", "Working directory (default: current)")

	envelopeInspectCmd.Flags().StringVarP(&envelopeWorkdir, "workdir", "w", "", "Working directory (default: current)")

	envelopeListCmd.Flags().StringVarP(&envelopeWorkdir, "workdir", "w", "", "Working directory (default: current)")

	envelopeCmd.AddCommand(envelopeCreateCmd)
	envelopeCmd.AddCommand(envelopeRunCmd)
	envelopeCmd.AddCommand(envelopeInspectCmd)
	envelopeCmd.AddCommand(envelopeListCmd)

	rootCmd.AddCommand(envelopeCmd)
}

func runEnvelopeCreate(cmd *cobra.Command, args []string) error {
	envPath, err := runenv.FindEnvInParents(envelopeWorkdir, runenv.MaxEnvSearchDepth)
	if err != nil {
		return fmt.Errorf("find .env: %w", err)
	}

	wsRoot, err := workspace.FindRoot(filepath.Dir(envPath))
	if err != nil {
		return fmt.Errorf("find workspace: %w", err)
	}

	ttl, err := parseTTL(envelopeTTL)
	if err != nil {
		return fmt.Errorf("parse ttl: %w", err)
	}

	secrets, err := runenv.LoadDecryptedEnv(envPath, wsRoot)
	if err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	filtered := make(map[string]string)
	for _, key := range envelopeScope {
		val, ok := secrets[key]
		if !ok {
			return fmt.Errorf("key %q not found in .env", key)
		}
		filtered[key] = val
	}

	env, err := envelope.Create(filtered, envelopeScope, ttl)
	if err != nil {
		return fmt.Errorf("create envelope: %w", err)
	}

	envStr, err := env.String()
	if err != nil {
		return fmt.Errorf("serialize envelope: %w", err)
	}

	if err := audit.Log(envelopeWorkdir, audit.OpEnvelopeCreate,
		audit.WithScope(envelopeScope),
		audit.WithSessionID(env.SessionID),
		audit.WithTTL(envelopeTTL),
	); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write audit log: %v\n", err)
	}

	if envelopeOutputFile != "" {
		return os.WriteFile(envelopeOutputFile, []byte(envStr), 0600)
	}

	fmt.Println(envStr)
	return nil
}

func runEnvelopeRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: openenvx envelope run <envelope> -- <command>")
	}

	envStr := args[0]
	if !strings.HasPrefix(envStr, "envelope:") {
		return fmt.Errorf("invalid envelope format")
	}

	var command string
	var cmdArgs []string
	if len(args) > 1 {
		command = args[1]
	}
	if len(args) > 2 {
		cmdArgs = args[2:]
	}
	if command == "" {
		return fmt.Errorf("no command specified")
	}

	env, err := envelope.Parse(envStr)
	if err != nil {
		return fmt.Errorf("parse envelope: %w", err)
	}

	info := env.Inspect()
	if info.Status == "expired" {
		return fmt.Errorf("envelope expired at %s", info.ExpiresAt.Format(time.RFC3339))
	}

	secrets, err := env.Unwrap()
	if err != nil {
		return fmt.Errorf("unwrap envelope: %w", err)
	}

	if err := audit.Log(envelopeWorkdir, audit.OpEnvelopeRun,
		audit.WithSessionID(env.SessionID),
		audit.WithCommand(command),
	); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write audit log: %v\n", err)
	}

	var exitCode int
	if envelopeRedact {
		exitCode, err = runenv.RunWithEnvRedactedFromMap(secrets, envelopeWorkdir, command, cmdArgs)
	} else {
		exitCode, err = runenv.RunWithEnvFromMap(secrets, envelopeWorkdir, command, cmdArgs)
	}

	if err != nil {
		if exitCode >= 0 {
			os.Exit(exitCode)
		}
		return err
	}

	return nil
}

func runEnvelopeInspect(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: openenvx envelope inspect <envelope>")
	}

	envStr := args[0]
	if !strings.HasPrefix(envStr, "envelope:") {
		return fmt.Errorf("invalid envelope format")
	}

	info, err := envelope.ParseAndInspect(envStr)
	if err != nil {
		return fmt.Errorf("inspect envelope: %w", err)
	}

	out := map[string]interface{}{
		"scope":         info.Scope,
		"expires_at":    info.ExpiresAt.Format(time.RFC3339),
		"status":        info.Status,
		"keys_included": info.KeysIncluded,
		"session_id":    info.SessionID,
	}

	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))

	if err := audit.Log(envelopeWorkdir, audit.OpEnvelopeInspect,
		audit.WithSessionID(info.SessionID),
	); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write audit log: %v\n", err)
	}

	return nil
}

func runEnvelopeList(cmd *cobra.Command, args []string) error {
	entries, err := audit.Show(envelopeWorkdir, 20)
	if err != nil {
		if err == audit.ErrNoAuditLog {
			fmt.Println("No audit log found. Create an envelope first.")
			return nil
		}
		return fmt.Errorf("read audit log: %w", err)
	}

	envelopeEntries := make([]audit.EntrySummary, 0)
	for _, e := range entries {
		if e.Op == string(audit.OpEnvelopeCreate) || e.Op == string(audit.OpEnvelopeRun) {
			envelopeEntries = append(envelopeEntries, e)
		}
	}

	if len(envelopeEntries) == 0 {
		fmt.Println("No envelope operations found in audit log.")
		return nil
	}

	b, _ := json.MarshalIndent(envelopeEntries, "", "  ")
	fmt.Println(string(b))

	return nil
}

func parseTTL(s string) (time.Duration, error) {
	if s == "" {
		return time.Hour, nil
	}
	return time.ParseDuration(s)
}
