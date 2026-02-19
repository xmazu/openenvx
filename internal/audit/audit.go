package audit

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	auditDir  = ".envx"
	auditFile = "audit.logl"
)

var (
	ErrNoAuditLog = errors.New("no audit log found")
	mu            sync.Mutex
)

type Op string

const (
	OpEnvelopeCreate  Op = "envelope_create"
	OpEnvelopeRun     Op = "envelope_run"
	OpEnvelopeInspect Op = "envelope_inspect"
	OpSet             Op = "set"
	OpGet             Op = "get"
	OpDelete          Op = "delete"
	OpRotate          Op = "rotate"
	OpMCPCall         Op = "mcp_call"
)

type Entry struct {
	Timestamp time.Time `json:"ts"`
	Op        Op        `json:"op"`
	Scope     []string  `json:"scope,omitempty"`
	SessionID string    `json:"sid,omitempty"`
	TTL       string    `json:"ttl,omitempty"`
	Command   string    `json:"cmd,omitempty"`
	ExitCode  int       `json:"exit,omitempty"`
	Tool      string    `json:"tool,omitempty"`
	PrevHash  string    `json:"prev_hash"`
}

type EntrySummary struct {
	Timestamp string   `json:"ts"`
	Op        string   `json:"op"`
	Scope     []string `json:"scope,omitempty"`
	SessionID string   `json:"sid,omitempty"`
	TTL       string   `json:"ttl,omitempty"`
	Command   string   `json:"cmd,omitempty"`
	ExitCode  int      `json:"exit,omitempty"`
	Tool      string   `json:"tool,omitempty"`
}

func auditPath(workdir string) string {
	if workdir == "" {
		workdir, _ = os.Getwd()
	}
	return filepath.Join(workdir, auditDir, auditFile)
}

func ensureDir(workdir string) error {
	path := auditPath(workdir)
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}

func lastHash(workdir string) string {
	path := auditPath(workdir)
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	var lastLine string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if lastLine == "" {
		return ""
	}

	hash := sha256.Sum256([]byte(lastLine))
	return hex.EncodeToString(hash[:])
}

func Log(workdir string, op Op, opts ...Option) error {
	mu.Lock()
	defer mu.Unlock()

	if err := ensureDir(workdir); err != nil {
		return fmt.Errorf("ensure audit dir: %w", err)
	}

	entry := &Entry{
		Timestamp: time.Now().UTC(),
		Op:        op,
		PrevHash:  lastHash(workdir),
	}

	for _, opt := range opts {
		opt(entry)
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}

	path := auditPath(workdir)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, string(b)); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}

type Option func(*Entry)

func WithScope(keys []string) Option {
	return func(e *Entry) {
		e.Scope = keys
	}
}

func WithSessionID(id string) Option {
	return func(e *Entry) {
		e.SessionID = id
	}
}

func WithTTL(ttl string) Option {
	return func(e *Entry) {
		e.TTL = ttl
	}
}

func WithCommand(cmd string) Option {
	return func(e *Entry) {
		e.Command = cmd
	}
}

func Show(workdir string, lastN int) ([]EntrySummary, error) {
	path := auditPath(workdir)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoAuditLog
		}
		return nil, fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()

	var entries []EntrySummary
	scanner := bufio.NewScanner(f)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if lastN > 0 && len(lines) > lastN {
		lines = lines[len(lines)-lastN:]
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		entries = append(entries, EntrySummary{
			Timestamp: e.Timestamp.Format(time.RFC3339),
			Op:        string(e.Op),
			Scope:     e.Scope,
			SessionID: e.SessionID,
			TTL:       e.TTL,
			Command:   e.Command,
			ExitCode:  e.ExitCode,
			Tool:      e.Tool,
		})
	}

	return entries, nil
}

type VerifyResult struct {
	TotalEntries  int
	Breaks        []int
	FirstModified int
}

func Verify(workdir string) (*VerifyResult, error) {
	path := auditPath(workdir)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoAuditLog
		}
		return nil, fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()

	result := &VerifyResult{}
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	result.TotalEntries = len(lines)

	if len(lines) == 0 {
		return result, nil
	}

	var firstEntry Entry
	if err := json.Unmarshal([]byte(lines[0]), &firstEntry); err != nil {
		result.FirstModified = 1
		return result, nil
	}

	if firstEntry.PrevHash != "" {
		result.Breaks = append(result.Breaks, 1)
	}

	for i := 1; i < len(lines); i++ {
		prevHash := sha256.Sum256([]byte(lines[i-1]))
		prevHashStr := hex.EncodeToString(prevHash[:])

		var entry Entry
		if err := json.Unmarshal([]byte(lines[i]), &entry); err != nil {
			result.Breaks = append(result.Breaks, i+1)
			continue
		}

		if entry.PrevHash != prevHashStr {
			result.Breaks = append(result.Breaks, i+1)
		}
	}

	return result, nil
}
