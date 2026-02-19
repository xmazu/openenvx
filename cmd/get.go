package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/runenv"
	"github.com/xmazu/openenvx/internal/workspace"
)

var getCmd = &cobra.Command{
	Use:   "get [KEY]",
	Short: "Get decrypted environment variable(s)",
	Long: `Get one or all decrypted environment variables from the .env file.
Without KEY, outputs all variables as JSON. With KEY, outputs the single value (for scripts: $(openenvx get KEY)).
Use --format shell or --format eval for shell-friendly output.
Use --masked to get a masked version of the value without seeing plaintext.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGet,
}

var getFile string
var getFormat string
var getMasked bool

func init() {
	getCmd.Flags().StringVarP(&getFile, "file", "f", ".env", "Path to .env file")
	getCmd.Flags().StringVar(&getFormat, "format", "json", "Output format: json, shell, or eval (default json when no KEY; raw value when KEY given)")
	getCmd.Flags().BoolVar(&getMasked, "masked", false, "Show masked value instead of plaintext (requires KEY)")
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	envFilePath, err := runenv.ResolveEnvPath(getFile, "")
	if err != nil {
		return fmt.Errorf("resolve .env path: %w", err)
	}

	wsRoot, err := workspace.FindRoot(filepath.Dir(envFilePath))
	if err != nil {
		return fmt.Errorf("find workspace: %w", err)
	}

	if len(args) == 1 {
		key := args[0]

		if getMasked {
			masked, valueLength, err := runenv.GetMaskedSecret(envFilePath, key)
			if err != nil {
				return err
			}
			output := map[string]interface{}{
				"key":          key,
				"masked_value": masked,
				"value_length": valueLength,
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			enc.SetEscapeHTML(false)
			return enc.Encode(output)
		}

		decrypted, err := runenv.LoadDecryptedEnv(envFilePath, wsRoot)
		if err != nil {
			return err
		}

		value, ok := decrypted[key]
		if !ok {
			return fmt.Errorf("key %q not found", key)
		}
		switch getFormat {
		case "shell":
			fmt.Print(shellEscape(key) + "=" + shellEscape(value))
			return nil
		case "eval":
			fmt.Print(shellEscape(key) + "=" + evalQuoted(value))
			return nil
		default:
			fmt.Print(value)
			return nil
		}
	}

	decrypted, err := runenv.LoadDecryptedEnv(envFilePath, wsRoot)
	if err != nil {
		return err
	}

	switch getFormat {
	case "shell":
		keys := make([]string, 0, len(decrypted))
		for k := range decrypted {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for i, k := range keys {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString(shellEscape(k))
			b.WriteString("=")
			b.WriteString(shellEscape(decrypted[k]))
		}
		fmt.Print(b.String())
		return nil
	case "eval":
		keys := make([]string, 0, len(decrypted))
		for k := range decrypted {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Println(shellEscape(k) + "=" + evalQuoted(decrypted[k]))
		}
		return nil
	default:
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		return enc.Encode(decrypted)
	}
}

func shellEscape(s string) string {
	if strings.ContainsAny(s, " \t\n\"'") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

func evalQuoted(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}
