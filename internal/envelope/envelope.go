package envelope

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xmazu/openenvx/internal/crypto"
)

const (
	prefix          = "envelope:v1:"
	hkdfInfo        = "openenvx-envelope-v1"
	unwrapSecretLen = 16
	sessionKeyLen   = 32
)

var (
	ErrInvalidFormat   = errors.New("invalid envelope format")
	ErrExpired         = errors.New("envelope has expired")
	ErrInvalidVersion  = errors.New("unsupported envelope version")
	ErrDecryption      = errors.New("failed to decrypt envelope")
	ErrScopeKeyMissing = errors.New("requested key not in envelope scope")
)

type Envelope struct {
	SessionID         string            `json:"session_id"`
	CreatedAt         time.Time         `json:"created_at"`
	ExpiresAt         time.Time         `json:"expires_at"`
	Scope             []string          `json:"scope"`
	UnwrapSecret      []byte            `json:"unwrap_secret"`
	WrappedSessionKey []byte            `json:"wrapped_session_key"`
	EncryptedSecrets  map[string][]byte `json:"encrypted_secrets"`
}

type Info struct {
	SessionID    string    `json:"session_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        []string  `json:"scope"`
	Status       string    `json:"status"`
	KeysIncluded int       `json:"keys_included"`
}

func Create(secrets map[string]string, scope []string, ttl time.Duration) (*Envelope, error) {
	if len(secrets) == 0 {
		return nil, errors.New("no secrets provided")
	}
	if len(scope) == 0 {
		return nil, errors.New("scope is required")
	}

	for _, key := range scope {
		if _, ok := secrets[key]; !ok {
			return nil, fmt.Errorf("%w: %q", ErrScopeKeyMissing, key)
		}
	}

	sessionKey := make([]byte, sessionKeyLen)
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, fmt.Errorf("generate session key: %w", err)
	}

	unwrapSecret := make([]byte, unwrapSecretLen)
	if _, err := rand.Read(unwrapSecret); err != nil {
		return nil, fmt.Errorf("generate unwrap secret: %w", err)
	}

	wrapKey := deriveWrapKey(unwrapSecret)

	wrappedSessionKey, err := wrapSessionKey(sessionKey, wrapKey)
	if err != nil {
		return nil, fmt.Errorf("wrap session key: %w", err)
	}

	encryptedSecrets := make(map[string][]byte, len(scope))
	for _, key := range scope {
		ciphertext, err := encryptWithSessionKey(sessionKey, []byte(secrets[key]), key)
		if err != nil {
			return nil, fmt.Errorf("encrypt %q: %w", key, err)
		}
		encryptedSecrets[key] = ciphertext
	}

	now := time.Now().UTC()
	return &Envelope{
		SessionID:         uuid.New().String(),
		CreatedAt:         now,
		ExpiresAt:         now.Add(ttl),
		Scope:             scope,
		UnwrapSecret:      unwrapSecret,
		WrappedSessionKey: wrappedSessionKey,
		EncryptedSecrets:  encryptedSecrets,
	}, nil
}

func (e *Envelope) Unwrap() (map[string]string, error) {
	if time.Now().UTC().After(e.ExpiresAt) {
		return nil, ErrExpired
	}

	wrapKey := deriveWrapKey(e.UnwrapSecret)

	sessionKey, err := unwrapSessionKey(e.WrappedSessionKey, wrapKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryption, err)
	}

	secrets := make(map[string]string, len(e.EncryptedSecrets))
	for key, ciphertext := range e.EncryptedSecrets {
		plaintext, err := decryptWithSessionKey(sessionKey, ciphertext, key)
		if err != nil {
			return nil, fmt.Errorf("decrypt %q: %w", key, err)
		}
		secrets[key] = string(plaintext)
	}

	return secrets, nil
}

func (e *Envelope) Inspect() *Info {
	status := "valid"
	if time.Now().UTC().After(e.ExpiresAt) {
		status = "expired"
	}

	return &Info{
		SessionID:    e.SessionID,
		CreatedAt:    e.CreatedAt,
		ExpiresAt:    e.ExpiresAt,
		Scope:        e.Scope,
		Status:       status,
		KeysIncluded: len(e.EncryptedSecrets),
	}
}

func (e *Envelope) String() (string, error) {
	payload := map[string]interface{}{
		"session_id":          e.SessionID,
		"created_at":          e.CreatedAt.Format(time.RFC3339),
		"expires_at":          e.ExpiresAt.Format(time.RFC3339),
		"scope":               e.Scope,
		"unwrap_secret":       base64.StdEncoding.EncodeToString(e.UnwrapSecret),
		"wrapped_session_key": base64.StdEncoding.EncodeToString(e.WrappedSessionKey),
		"encrypted_secrets":   make(map[string]string),
	}

	es := payload["encrypted_secrets"].(map[string]string)
	for key, ct := range e.EncryptedSecrets {
		es[key] = base64.StdEncoding.EncodeToString(ct)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal envelope: %w", err)
	}

	return prefix + base64.StdEncoding.EncodeToString(b), nil
}

func Parse(s string) (*Envelope, error) {
	if !strings.HasPrefix(s, prefix) {
		return nil, ErrInvalidFormat
	}

	b64payload := s[len(prefix):]
	payload, err := base64.StdEncoding.DecodeString(b64payload)
	if err != nil {
		return nil, fmt.Errorf("%w: base64 decode: %v", ErrInvalidFormat, err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("%w: json parse: %v", ErrInvalidFormat, err)
	}

	sessionID, _ := raw["session_id"].(string)
	createdAtStr, _ := raw["created_at"].(string)
	expiresAtStr, _ := raw["expires_at"].(string)

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("%w: parse created_at: %v", ErrInvalidFormat, err)
	}
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("%w: parse expires_at: %v", ErrInvalidFormat, err)
	}

	var scope []string
	if s, ok := raw["scope"].([]interface{}); ok {
		for _, v := range s {
			if str, ok := v.(string); ok {
				scope = append(scope, str)
			}
		}
	}

	unwrapSecretB64, _ := raw["unwrap_secret"].(string)
	unwrapSecret, err := base64.StdEncoding.DecodeString(unwrapSecretB64)
	if err != nil {
		return nil, fmt.Errorf("%w: decode unwrap_secret: %v", ErrInvalidFormat, err)
	}

	wrappedSessionKeyB64, _ := raw["wrapped_session_key"].(string)
	wrappedSessionKey, err := base64.StdEncoding.DecodeString(wrappedSessionKeyB64)
	if err != nil {
		return nil, fmt.Errorf("%w: decode wrapped_session_key: %v", ErrInvalidFormat, err)
	}

	encryptedSecretsRaw, _ := raw["encrypted_secrets"].(map[string]interface{})
	encryptedSecrets := make(map[string][]byte)
	for key, val := range encryptedSecretsRaw {
		if b64, ok := val.(string); ok {
			ct, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				continue
			}
			encryptedSecrets[key] = ct
		}
	}

	return &Envelope{
		SessionID:         sessionID,
		CreatedAt:         createdAt,
		ExpiresAt:         expiresAt,
		Scope:             scope,
		UnwrapSecret:      unwrapSecret,
		WrappedSessionKey: wrappedSessionKey,
		EncryptedSecrets:  encryptedSecrets,
	}, nil
}

func ParseAndInspect(s string) (*Info, error) {
	e, err := Parse(s)
	if err != nil {
		return nil, err
	}
	return e.Inspect(), nil
}

func deriveWrapKey(unwrapSecret []byte) []byte {
	h := sha256.New()
	h.Write(unwrapSecret)
	h.Write([]byte(hkdfInfo))
	return h.Sum(nil)[:sessionKeyLen]
}

func wrapSessionKey(sessionKey, wrapKey []byte) ([]byte, error) {
	return crypto.EncryptAESGCM(wrapKey, sessionKey, nil)
}

func unwrapSessionKey(wrappedKey, wrapKey []byte) ([]byte, error) {
	return crypto.DecryptAESGCM(wrapKey, wrappedKey, nil)
}

func encryptWithSessionKey(sessionKey, plaintext []byte, keyName string) ([]byte, error) {
	return crypto.EncryptAESGCM(sessionKey, plaintext, []byte(keyName))
}

func decryptWithSessionKey(sessionKey, ciphertext []byte, keyName string) ([]byte, error) {
	return crypto.DecryptAESGCM(sessionKey, ciphertext, []byte(keyName))
}
