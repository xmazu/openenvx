package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/audit"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View and verify audit log",
	Long: `View the audit log and verify chain integrity.

The audit log tracks all envelope and secret operations.
Each entry links to the previous entry, forming a tamper-evident chain.`,
}

var auditShowCmd = &cobra.Command{
	Use:   "show [--last=N]",
	Short: "Show audit log entries",
	Long: `Display recent audit log entries.

Shows operation type, timestamp, scope, and session ID.`,
	RunE: runAuditShow,
}

var auditVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify audit log chain integrity",
	Long: `Verify that the audit log chain is intact.

Checks that each entry's prev_hash matches the hash of the previous entry.
Reports any breaks in the chain.`,
	RunE: runAuditVerify,
}

var (
	auditLastN   int
	auditWorkdir string
)

func init() {
	auditShowCmd.Flags().IntVarP(&auditLastN, "last", "n", 10, "Number of entries to show")
	auditShowCmd.Flags().StringVarP(&auditWorkdir, "workdir", "w", "", "Working directory (default: current)")

	auditVerifyCmd.Flags().StringVarP(&auditWorkdir, "workdir", "w", "", "Working directory (default: current)")

	auditCmd.AddCommand(auditShowCmd)
	auditCmd.AddCommand(auditVerifyCmd)

	rootCmd.AddCommand(auditCmd)
}

func runAuditShow(cmd *cobra.Command, args []string) error {
	entries, err := audit.Show(auditWorkdir, auditLastN)
	if err != nil {
		if err == audit.ErrNoAuditLog {
			fmt.Println("No audit log found. Operations will be logged when you create envelopes or use secrets.")
			return nil
		}
		return fmt.Errorf("read audit log: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No entries in audit log.")
		return nil
	}

	b, _ := json.MarshalIndent(entries, "", "  ")
	fmt.Println(string(b))

	return nil
}

func runAuditVerify(cmd *cobra.Command, args []string) error {
	result, err := audit.Verify(auditWorkdir)
	if err != nil {
		if err == audit.ErrNoAuditLog {
			fmt.Println("No audit log found.")
			return nil
		}
		return fmt.Errorf("verify audit log: %w", err)
	}

	fmt.Printf("Audit log verified: %d entries\n", result.TotalEntries)

	if len(result.Breaks) == 0 {
		fmt.Println("Chain integrity: OK")
		return nil
	}

	fmt.Printf("Chain breaks detected at lines: %v\n", result.Breaks)
	fmt.Println("Warning: Log may have been tampered with.")
	return nil
}
