package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "openenvx",
	Short: "Secure environment variable management",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `OpenEnvX - Production-grade, zero-config CLI for local env vars with age (X25519) encryption.
Opinionated: one workflow. Share .env in the repo; share the private key with your team (e.g. 1Password)-teammates run ` + "`openenvx key add`" + ` then ` + "`openenvx run -- ...`" + `.

ENCRYPTION:

  - Public key in repo: anyone can add secrets (encrypt).
  - Private key with you and your team: share via password manager; only key holder can decrypt.
  - Same .env file everywhere; no extra config.

EXAMPLES:

  openenvx init
  openenvx set DATABASE_URL=postgres://user:pass@localhost/db
  openenvx run -- node server.js

  # New teammate: paste key from 1Password, then run
  openenvx key add
  openenvx run -- npm start

  # In CI: set OPENENVX_PRIVATE_KEY and run the same command.

Get started: openenvx init --help`,
}

func init() {
	// Cobra adds --version when Version is set; use a clear template
	rootCmd.SetVersionTemplate("openenvx version {{.Version}}\n")
}

// SetVersion sets the version string shown by --version (e.g. from ldflags).
func SetVersion(v string) { rootCmd.Version = v }

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
