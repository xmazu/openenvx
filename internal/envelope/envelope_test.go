package envelope

import (
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	secrets := map[string]string{
		"DATABASE_URL": "postgres://localhost:5432/db",
		"API_KEY":      "sk-test-123456",
	}
	scope := []string{"DATABASE_URL", "API_KEY"}

	env, err := Create(secrets, scope, time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if env.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if len(env.Scope) != 2 {
		t.Errorf("Scope length = %d, want 2", len(env.Scope))
	}
	if len(env.UnwrapSecret) != unwrapSecretLen {
		t.Errorf("UnwrapSecret length = %d, want %d", len(env.UnwrapSecret), unwrapSecretLen)
	}
	if len(env.WrappedSessionKey) == 0 {
		t.Error("WrappedSessionKey should not be empty")
	}
	if len(env.EncryptedSecrets) != 2 {
		t.Errorf("EncryptedSecrets length = %d, want 2", len(env.EncryptedSecrets))
	}
	if env.ExpiresAt.Before(env.CreatedAt) {
		t.Error("ExpiresAt should be after CreatedAt")
	}
}

func TestCreate_EmptySecrets(t *testing.T) {
	_, err := Create(map[string]string{}, []string{"KEY"}, time.Hour)
	if err == nil {
		t.Error("Create with empty secrets should fail")
	}
}

func TestCreate_EmptyScope(t *testing.T) {
	_, err := Create(map[string]string{"KEY": "value"}, []string{}, time.Hour)
	if err == nil {
		t.Error("Create with empty scope should fail")
	}
}

func TestCreate_KeyNotInSecrets(t *testing.T) {
	_, err := Create(map[string]string{"KEY1": "value"}, []string{"KEY2"}, time.Hour)
	if err == nil {
		t.Error("Create with scope key not in secrets should fail")
	}
}

func TestUnwrap(t *testing.T) {
	secrets := map[string]string{
		"DATABASE_URL": "postgres://localhost:5432/db",
		"API_KEY":      "sk-test-123456",
	}
	scope := []string{"DATABASE_URL", "API_KEY"}

	env, err := Create(secrets, scope, time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	unwrapped, err := env.Unwrap()
	if err != nil {
		t.Fatalf("Unwrap failed: %v", err)
	}

	if unwrapped["DATABASE_URL"] != secrets["DATABASE_URL"] {
		t.Errorf("DATABASE_URL = %q, want %q", unwrapped["DATABASE_URL"], secrets["DATABASE_URL"])
	}
	if unwrapped["API_KEY"] != secrets["API_KEY"] {
		t.Errorf("API_KEY = %q, want %q", unwrapped["API_KEY"], secrets["API_KEY"])
	}
}

func TestUnwrap_Expired(t *testing.T) {
	secrets := map[string]string{"KEY": "value"}
	scope := []string{"KEY"}

	env, err := Create(secrets, scope, -time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, err = env.Unwrap()
	if err != ErrExpired {
		t.Errorf("Unwrap error = %v, want ErrExpired", err)
	}
}

func TestInspect(t *testing.T) {
	secrets := map[string]string{"KEY": "value"}
	scope := []string{"KEY"}

	env, err := Create(secrets, scope, time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	info := env.Inspect()

	if info.SessionID != env.SessionID {
		t.Errorf("SessionID = %q, want %q", info.SessionID, env.SessionID)
	}
	if info.Status != "valid" {
		t.Errorf("Status = %q, want valid", info.Status)
	}
	if len(info.Scope) != 1 {
		t.Errorf("Scope length = %d, want 1", len(info.Scope))
	}
}

func TestInspect_Expired(t *testing.T) {
	secrets := map[string]string{"KEY": "value"}
	scope := []string{"KEY"}

	env, err := Create(secrets, scope, -time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	info := env.Inspect()

	if info.Status != "expired" {
		t.Errorf("Status = %q, want expired", info.Status)
	}
}

func TestStringAndParse(t *testing.T) {
	secrets := map[string]string{
		"DATABASE_URL": "postgres://localhost:5432/db",
		"API_KEY":      "sk-test-123456",
	}
	scope := []string{"DATABASE_URL", "API_KEY"}

	env, err := Create(secrets, scope, time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	str, err := env.String()
	if err != nil {
		t.Fatalf("String failed: %v", err)
	}

	if str == "" {
		t.Fatal("String should not be empty")
	}

	parsed, err := Parse(str)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if parsed.SessionID != env.SessionID {
		t.Errorf("SessionID = %q, want %q", parsed.SessionID, env.SessionID)
	}
	if len(parsed.Scope) != len(env.Scope) {
		t.Errorf("Scope length = %d, want %d", len(parsed.Scope), len(env.Scope))
	}
	if len(parsed.EncryptedSecrets) != len(env.EncryptedSecrets) {
		t.Errorf("EncryptedSecrets length = %d, want %d", len(parsed.EncryptedSecrets), len(env.EncryptedSecrets))
	}

	unwrapped, err := parsed.Unwrap()
	if err != nil {
		t.Fatalf("Unwrap after Parse failed: %v", err)
	}

	if unwrapped["DATABASE_URL"] != secrets["DATABASE_URL"] {
		t.Errorf("DATABASE_URL after round-trip = %q, want %q", unwrapped["DATABASE_URL"], secrets["DATABASE_URL"])
	}
}

func TestParse_InvalidFormat(t *testing.T) {
	_, err := Parse("invalid")
	if err != ErrInvalidFormat {
		t.Errorf("Parse error = %v, want ErrInvalidFormat", err)
	}
}

func TestParse_InvalidBase64(t *testing.T) {
	_, err := Parse(prefix + "not-valid-base64!!!")
	if err == nil {
		t.Error("Parse with invalid base64 should fail")
	}
}

func TestRoundTrip_SpecialCharacters(t *testing.T) {
	secrets := map[string]string{
		"KEY": "value with spaces, symbols!@#$%^&*(), and unicode: 你好世界 🎉",
	}
	scope := []string{"KEY"}

	env, err := Create(secrets, scope, time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	str, err := env.String()
	if err != nil {
		t.Fatalf("String failed: %v", err)
	}

	parsed, err := Parse(str)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	unwrapped, err := parsed.Unwrap()
	if err != nil {
		t.Fatalf("Unwrap failed: %v", err)
	}

	if unwrapped["KEY"] != secrets["KEY"] {
		t.Errorf("KEY = %q, want %q", unwrapped["KEY"], secrets["KEY"])
	}
}

func TestRoundTrip_ManyKeys(t *testing.T) {
	secrets := make(map[string]string)
	scope := make([]string, 100)
	for i := 0; i < 100; i++ {
		key := "KEY_" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		secrets[key] = "value_" + key
		scope[i] = key
	}

	env, err := Create(secrets, scope, time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	str, err := env.String()
	if err != nil {
		t.Fatalf("String failed: %v", err)
	}

	parsed, err := Parse(str)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	unwrapped, err := parsed.Unwrap()
	if err != nil {
		t.Fatalf("Unwrap failed: %v", err)
	}

	if len(unwrapped) != 100 {
		t.Errorf("Unwrapped count = %d, want 100", len(unwrapped))
	}

	for key, want := range secrets {
		if got := unwrapped[key]; got != want {
			t.Errorf("unwrapped[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestParseAndInspect(t *testing.T) {
	secrets := map[string]string{"KEY": "value"}
	scope := []string{"KEY"}

	env, err := Create(secrets, scope, time.Hour)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	str, err := env.String()
	if err != nil {
		t.Fatalf("String failed: %v", err)
	}

	info, err := ParseAndInspect(str)
	if err != nil {
		t.Fatalf("ParseAndInspect failed: %v", err)
	}

	if info.SessionID != env.SessionID {
		t.Errorf("SessionID = %q, want %q", info.SessionID, env.SessionID)
	}
	if info.KeysIncluded != 1 {
		t.Errorf("KeysIncluded = %d, want 1", info.KeysIncluded)
	}
}
