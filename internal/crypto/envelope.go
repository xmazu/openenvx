package crypto

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const EncryptedValuePrefix = "envx:"

type EncryptedValue struct {
	WrappedDEK string
	Ciphertext string
}

func (e *EncryptedValue) String() string {
	return fmt.Sprintf("%s%s:%s", EncryptedValuePrefix, e.WrappedDEK, e.Ciphertext)
}

func ParseEncryptedValue(s string) (*EncryptedValue, error) {
	if !strings.HasPrefix(s, EncryptedValuePrefix) {
		return nil, fmt.Errorf("not an encrypted value")
	}
	rest := s[len(EncryptedValuePrefix):]

	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid encrypted value format")
	}

	return &EncryptedValue{
		WrappedDEK: parts[0],
		Ciphertext: parts[1],
	}, nil
}

type Envelope struct {
	masterKey *MasterKey
}

func NewEnvelope(masterKey *MasterKey) *Envelope {
	return &Envelope{masterKey: masterKey}
}

func (e *Envelope) Encrypt(plaintext []byte, keyName string) (*EncryptedValue, error) {
	dek, err := GenerateDEK()
	if err != nil {
		return nil, err
	}

	_, ciphertext, err := dek.Encrypt(plaintext, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with DEK: %w", err)
	}

	wrappedDEK, err := e.wrapDEK(dek)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap DEK: %w", err)
	}

	return &EncryptedValue{
		WrappedDEK: base64.StdEncoding.EncodeToString(wrappedDEK),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

func (e *Envelope) Decrypt(ev *EncryptedValue, keyName string) ([]byte, error) {
	wrappedDEK, err := base64.StdEncoding.DecodeString(ev.WrappedDEK)
	if err != nil {
		return nil, fmt.Errorf("failed to decode wrapped DEK: %w", err)
	}

	dek, err := e.unwrapDEK(wrappedDEK)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap DEK: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ev.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	plaintext, err := dek.Decrypt(ciphertext, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

func (e *Envelope) wrapDEK(dek *DEK) ([]byte, error) {
	return EncryptAESGCM(e.masterKey.Bytes(), dek.Bytes(), nil)
}

func (e *Envelope) unwrapDEK(wrapped []byte) (*DEK, error) {
	key, err := DecryptAESGCM(e.masterKey.Bytes(), wrapped, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap DEK: %w", err)
	}
	return DEKFromBytes(key), nil
}
