package audit

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	tmp := t.TempDir()

	err := Log(tmp, OpSet, WithScope([]string{"TEST_KEY"}))
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	path := auditPath(tmp)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("audit file not created: %v", err)
	}

	entries, err := Show(tmp, 10)
	if err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("entries count = %d, want 1", len(entries))
	}

	if entries[0].Op != string(OpSet) {
		t.Errorf("op = %q, want %q", entries[0].Op, OpSet)
	}

	if len(entries[0].Scope) != 1 || entries[0].Scope[0] != "TEST_KEY" {
		t.Errorf("scope = %v, want [TEST_KEY]", entries[0].Scope)
	}
}

func TestLog_MultipleEntries(t *testing.T) {
	tmp := t.TempDir()

	for i := 0; i < 5; i++ {
		err := Log(tmp, OpSet, WithScope([]string{"KEY"}))
		if err != nil {
			t.Fatalf("Log %d failed: %v", i, err)
		}
	}

	entries, err := Show(tmp, 0)
	if err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("entries count = %d, want 5", len(entries))
	}
}

func TestLog_AllOptions(t *testing.T) {
	tmp := t.TempDir()

	err := Log(tmp, OpEnvelopeCreate,
		WithScope([]string{"KEY1", "KEY2"}),
		WithSessionID("test-session-id"),
		WithTTL("1h"),
	)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	entries, err := Show(tmp, 10)
	if err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("entries count = %d, want 1", len(entries))
	}

	e := entries[0]
	if e.Op != string(OpEnvelopeCreate) {
		t.Errorf("op = %q, want %q", e.Op, OpEnvelopeCreate)
	}
	if len(e.Scope) != 2 {
		t.Errorf("scope length = %d, want 2", len(e.Scope))
	}
	if e.SessionID != "test-session-id" {
		t.Errorf("session_id = %q, want test-session-id", e.SessionID)
	}
	if e.TTL != "1h" {
		t.Errorf("ttl = %q, want 1h", e.TTL)
	}
}

func TestShow_LastN(t *testing.T) {
	tmp := t.TempDir()

	for i := 0; i < 10; i++ {
		err := Log(tmp, OpSet)
		if err != nil {
			t.Fatalf("Log %d failed: %v", i, err)
		}
		time.Sleep(1 * time.Millisecond)
	}

	entries, err := Show(tmp, 3)
	if err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("entries count = %d, want 3", len(entries))
	}
}

func TestShow_NoLog(t *testing.T) {
	tmp := t.TempDir()

	_, err := Show(tmp, 10)
	if err != ErrNoAuditLog {
		t.Errorf("error = %v, want ErrNoAuditLog", err)
	}
}

func TestVerify(t *testing.T) {
	tmp := t.TempDir()

	for i := 0; i < 5; i++ {
		err := Log(tmp, OpSet)
		if err != nil {
			t.Fatalf("Log %d failed: %v", i, err)
		}
	}

	result, err := Verify(tmp)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if result.TotalEntries != 5 {
		t.Errorf("total entries = %d, want 5", result.TotalEntries)
	}
	if len(result.Breaks) != 0 {
		t.Errorf("breaks = %v, want empty", result.Breaks)
	}
}

func TestVerify_NoLog(t *testing.T) {
	tmp := t.TempDir()

	_, err := Verify(tmp)
	if err != ErrNoAuditLog {
		t.Errorf("error = %v, want ErrNoAuditLog", err)
	}
}

func TestVerify_Tampered(t *testing.T) {
	tmp := t.TempDir()

	for i := 0; i < 3; i++ {
		err := Log(tmp, OpSet)
		if err != nil {
			t.Fatalf("Log %d failed: %v", i, err)
		}
	}

	path := auditPath(tmp)
	f, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open audit log: %v", err)
	}
	defer f.Close()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}

	tampered := string(content) + `{"ts":"2026-01-01T00:00:00Z","op":"set","prev_hash":"tampered"}` + "\n"
	if err := os.WriteFile(path, []byte(tampered), 0644); err != nil {
		t.Fatalf("write tampered log: %v", err)
	}

	result, err := Verify(tmp)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if len(result.Breaks) == 0 {
		t.Error("expected breaks in tampered log")
	}
}

func TestVerify_DeletedMiddle(t *testing.T) {
	tmp := t.TempDir()

	for i := 0; i < 5; i++ {
		err := Log(tmp, OpSet)
		if err != nil {
			t.Fatalf("Log %d failed: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	path := auditPath(tmp)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	for i, line := range lines {
		if i == 2 {
			continue
		}
		if line != "" {
			newLines = append(newLines, line)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create audit log: %v", err)
	}
	for _, line := range newLines {
		fmt.Fprintln(f, line)
	}
	f.Close()

	result, err := Verify(tmp)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if len(result.Breaks) == 0 {
		t.Error("expected breaks after deletion")
	}
}

func TestChainIntegrity(t *testing.T) {
	tmp := t.TempDir()

	err := Log(tmp, OpEnvelopeCreate, WithSessionID("session-1"), WithTTL("1h"))
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	err = Log(tmp, OpEnvelopeRun, WithSessionID("session-1"), WithCommand("npm test"))
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	result, err := Verify(tmp)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if result.TotalEntries != 2 {
		t.Errorf("total entries = %d, want 2", result.TotalEntries)
	}

	entries, err := Show(tmp, 0)
	if err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("entries count = %d, want 2", len(entries))
	}

	if entries[0].Op != string(OpEnvelopeCreate) {
		t.Errorf("first entry op = %q, want %q", entries[0].Op, OpEnvelopeCreate)
	}
	if entries[1].Op != string(OpEnvelopeRun) {
		t.Errorf("second entry op = %q, want %q", entries[1].Op, OpEnvelopeRun)
	}
}
