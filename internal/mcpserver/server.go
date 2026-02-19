package mcpserver

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xmazu/openenvx/internal/audit"
	"github.com/xmazu/openenvx/internal/envelope"
	"github.com/xmazu/openenvx/internal/runenv"
	"github.com/xmazu/openenvx/internal/workspace"
)

func Run(ctx context.Context) error {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "openenvx",
		Version: "0.1.0",
	}, nil)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "create_secret",
		Description: "Create or update a secret in the project's .env file. Use when the agent decides to store a new secret (e.g. API key, token). Finds .env in the given directory or parent directories. Requires the project private key to be available (OPENENVX_PRIVATE_KEY or openenvx key add). Returns ok and path only; never returns or echoes the secret value.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Key     string `json:"key" jsonschema:"environment variable name (e.g. API_KEY, DATABASE_URL)"`
		Value   string `json:"value" jsonschema:"secret value to store (will be encrypted)"`
		Workdir string `json:"workdir" jsonschema:"directory to search for .env (default: current)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		if args.Key == "" {
			return errorResult("key is required"), nil, nil
		}
		envPath, err := runenv.FindEnvInParents(args.Workdir, runenv.MaxEnvSearchDepth)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		if err := runenv.SetSecret(envPath, args.Key, args.Value); err != nil {
			return errorResult(err.Error()), nil, nil
		}
		return successResult(map[string]any{"ok": true, "path": envPath}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_keys",
		Description: "List key names for all encrypted variables in the project's .env. Returns keys and envPath only; never returns values. Use to see what secrets exist so you can suggest openenvx run --redact -- ... or ask the user to add a key.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Workdir string `json:"workdir" jsonschema:"directory to search for .env (default: current)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		envPath, err := runenv.FindEnvInParents(args.Workdir, runenv.MaxEnvSearchDepth)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		keys, err := runenv.ListEncryptedKeys(envPath)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		return successResult(map[string]any{"keys": keys, "envPath": envPath}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "get_masked",
		Description: "Get a masked version of a secret value (e.g. ****WXYZ) and its length. Use to verify a key exists without seeing the value. Requires the project private key. Never returns plaintext.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Workdir string `json:"workdir" jsonschema:"directory to search for .env (default: current)"`
		Key     string `json:"key" jsonschema:"environment variable name (e.g. API_KEY)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		if args.Key == "" {
			return errorResult("key is required"), nil, nil
		}
		envPath, err := runenv.FindEnvInParents(args.Workdir, runenv.MaxEnvSearchDepth)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		masked, valueLength, err := runenv.GetMaskedSecret(envPath, args.Key)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		return successResult(map[string]any{
			"key":          args.Key,
			"masked_value": masked,
			"value_length": valueLength,
		}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "key_exists",
		Description: "Check if an encrypted key exists in the .env file. Lightweight operation - no decryption required. Returns exists boolean. Use for quick existence checks before creating envelopes or running commands.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Workdir string `json:"workdir" jsonschema:"directory to search for .env (default: current)"`
		Key     string `json:"key" jsonschema:"environment variable name (e.g. API_KEY)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		if args.Key == "" {
			return errorResult("key is required"), nil, nil
		}
		envPath, err := runenv.FindEnvInParents(args.Workdir, runenv.MaxEnvSearchDepth)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		exists, err := runenv.KeyExists(envPath, args.Key)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		return successResult(map[string]any{"key": args.Key, "exists": exists}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "delete_secret",
		Description: "Delete a secret from the project's .env file. Use when cleaning up temporary secrets or removing outdated credentials. Returns ok and path. Warning: this operation is irreversible.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Key     string `json:"key" jsonschema:"environment variable name to delete (e.g. API_KEY)"`
		Workdir string `json:"workdir" jsonschema:"directory to search for .env (default: current)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		if args.Key == "" {
			return errorResult("key is required"), nil, nil
		}
		envPath, err := runenv.FindEnvInParents(args.Workdir, runenv.MaxEnvSearchDepth)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		deleted, err := runenv.DeleteSecret(envPath, args.Key)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}
		if !deleted {
			return errorResult(fmt.Sprintf("key %q not found in .env", args.Key)), nil, nil
		}
		_ = audit.Log(args.Workdir, audit.OpDelete, audit.WithScope([]string{args.Key}))
		return successResult(map[string]any{"ok": true, "deleted": true, "key": args.Key, "path": envPath}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "envelope_create",
		Description: "Create a self-contained envelope containing encrypted secrets. The envelope can be unwrapped without the master private key. Use this to give an agent time-limited access to specific secrets. Security: keep TTL short, scope minimal. Returns the envelope string and metadata.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Keys     []string `json:"keys" jsonschema:"required,list of secret keys to include in envelope"`
		TtlHours int      `json:"ttl_hours" jsonschema:"time-to-live in hours (default: 1)"`
		Workdir  string   `json:"workdir" jsonschema:"directory to search for .env (default: current)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		if len(args.Keys) == 0 {
			return errorResult("keys is required"), nil, nil
		}

		envPath, err := runenv.FindEnvInParents(args.Workdir, runenv.MaxEnvSearchDepth)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}

		ttl := time.Hour
		if args.TtlHours > 0 {
			ttl = time.Duration(args.TtlHours) * time.Hour
		}

		wsRoot, err := workspace.FindRoot(filepath.Dir(envPath))
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}

		secrets, err := runenv.LoadDecryptedEnv(envPath, wsRoot)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}

		filtered := make(map[string]string)
		for _, key := range args.Keys {
			val, ok := secrets[key]
			if !ok {
				return errorResult(fmt.Sprintf("key %q not found in .env", key)), nil, nil
			}
			filtered[key] = val
		}

		env, err := envelope.Create(filtered, args.Keys, ttl)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}

		envStr, err := env.String()
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}

		_ = audit.Log(args.Workdir, audit.OpEnvelopeCreate,
			audit.WithScope(args.Keys),
			audit.WithSessionID(env.SessionID),
			audit.WithTTL(fmt.Sprintf("%dh", args.TtlHours)),
		)

		return successResult(map[string]any{
			"envelope":   envStr,
			"expires_at": env.ExpiresAt.Format(time.RFC3339),
			"scope":      args.Keys,
			"session_id": env.SessionID,
		}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "envelope_run",
		Description: "Unwrap an envelope and run a command with secrets injected into the environment. The envelope is decrypted, secrets are injected, and the command is executed. Use redact_output to hide secrets in command output. Returns stdout and stderr (redacted if requested).",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Envelope     string   `json:"envelope" jsonschema:"envelope string (envelope:v1:...)"`
		Command      string   `json:"command" jsonschema:"command to run"`
		CommandArgs  []string `json:"command_args" jsonschema:"command arguments"`
		RedactOutput bool     `json:"redact_output" jsonschema:"redact secrets in output"`
		Workdir      string   `json:"workdir" jsonschema:"working directory for command"`
	}) (*mcpsdk.CallToolResult, any, error) {
		if args.Envelope == "" {
			return errorResult("envelope is required"), nil, nil
		}
		if args.Command == "" {
			return errorResult("command is required"), nil, nil
		}
		if !strings.HasPrefix(args.Envelope, "envelope:") {
			return errorResult("invalid envelope format"), nil, nil
		}

		env, err := envelope.Parse(args.Envelope)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}

		info := env.Inspect()
		if info.Status == "expired" {
			return errorResult(fmt.Sprintf("envelope expired at %s", info.ExpiresAt.Format(time.RFC3339))), nil, nil
		}

		secrets, err := env.Unwrap()
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}

		_ = audit.Log(args.Workdir, audit.OpEnvelopeRun,
			audit.WithSessionID(env.SessionID),
			audit.WithCommand(args.Command),
		)

		var result *runenv.RunResult
		if args.RedactOutput {
			result, err = runenv.RunWithEnvCapturedRedacted(secrets, args.Workdir, args.Command, args.CommandArgs)
		} else {
			result, err = runenv.RunWithEnvCaptured(secrets, args.Workdir, args.Command, args.CommandArgs)
		}

		out := map[string]any{
			"exit_code":  result.ExitCode,
			"stdout":     result.Stdout,
			"stderr":     result.Stderr,
			"success":    err == nil,
			"redacted":   args.RedactOutput,
			"session_id": env.SessionID,
		}
		if err != nil {
			out["error"] = err.Error()
		}
		return successResult(out), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "envelope_inspect",
		Description: "Inspect envelope metadata without decrypting secrets. Returns scope, expiration, and status. Safe to call - never exposes secret values.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Envelope string `json:"envelope" jsonschema:"envelope string (envelope:v1:...)"`
		Workdir  string `json:"workdir" jsonschema:"working directory for audit log"`
	}) (*mcpsdk.CallToolResult, any, error) {
		if args.Envelope == "" {
			return errorResult("envelope is required"), nil, nil
		}
		if !strings.HasPrefix(args.Envelope, "envelope:") {
			return errorResult("invalid envelope format"), nil, nil
		}

		info, err := envelope.ParseAndInspect(args.Envelope)
		if err != nil {
			return errorResult(err.Error()), nil, nil
		}

		_ = audit.Log(args.Workdir, audit.OpEnvelopeInspect,
			audit.WithSessionID(info.SessionID),
		)

		return successResult(map[string]any{
			"scope":         info.Scope,
			"expires_at":    info.ExpiresAt.Format(time.RFC3339),
			"status":        info.Status,
			"keys_included": info.KeysIncluded,
			"session_id":    info.SessionID,
		}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "audit_recent",
		Description: "Show recent audit log entries. Use to track envelope usage and secret operations. Returns list of operations with timestamps.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Count   int    `json:"count" jsonschema:"number of entries to return (default: 10)"`
		Workdir string `json:"workdir" jsonschema:"directory to search for audit log (default: current)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		count := args.Count
		if count <= 0 {
			count = 10
		}

		entries, err := audit.Show(args.Workdir, count)
		if err != nil {
			if err == audit.ErrNoAuditLog {
				return successResult(map[string]any{"entries": []any{}, "message": "No audit log found"}), nil, nil
			}
			return errorResult(err.Error()), nil, nil
		}

		return successResult(map[string]any{"entries": entries}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "audit_verify",
		Description: "Verify the integrity of the audit log chain. Checks that each entry's prev_hash matches the hash of the previous entry. Reports any breaks in the chain.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Workdir string `json:"workdir" jsonschema:"directory to search for audit log (default: current)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		result, err := audit.Verify(args.Workdir)
		if err != nil {
			if err == audit.ErrNoAuditLog {
				return successResult(map[string]any{"verified": false, "message": "No audit log found"}), nil, nil
			}
			return errorResult(err.Error()), nil, nil
		}

		msg := "Audit log chain integrity verified"
		if len(result.Breaks) > 0 {
			msg = "Chain breaks detected - log may have been tampered with"
		}
		return successResult(map[string]any{
			"verified":      len(result.Breaks) == 0,
			"total_entries": result.TotalEntries,
			"breaks":        result.Breaks,
			"message":       msg,
		}), nil, nil
	})

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "envelope_list",
		Description: "List recent envelope operations from the audit log. Shows envelope_create and envelope_run entries with session IDs and timestamps.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, args struct {
		Count   int    `json:"count" jsonschema:"number of entries to return (default: 20)"`
		Workdir string `json:"workdir" jsonschema:"directory to search for audit log (default: current)"`
	}) (*mcpsdk.CallToolResult, any, error) {
		count := args.Count
		if count <= 0 {
			count = 20
		}

		entries, err := audit.Show(args.Workdir, count)
		if err != nil {
			if err == audit.ErrNoAuditLog {
				return successResult(map[string]any{"envelopes": []any{}, "message": "No audit log found"}), nil, nil
			}
			return errorResult(err.Error()), nil, nil
		}

		var envelopeEntries []audit.EntrySummary
		for _, e := range entries {
			if e.Op == string(audit.OpEnvelopeCreate) || e.Op == string(audit.OpEnvelopeRun) {
				envelopeEntries = append(envelopeEntries, e)
			}
		}

		return successResult(map[string]any{
			"envelopes": envelopeEntries,
			"count":     len(envelopeEntries),
		}), nil, nil
	})

	return server.Run(ctx, &mcpsdk.StdioTransport{})
}
