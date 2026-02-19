package crypto

import (
	"crypto/rand"
	"fmt"
)

type DEK struct {
	key []byte
}

func GenerateDEK() (*DEK, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}
	return &DEK{key: key}, nil
}

func DEKFromBytes(key []byte) *DEK {
	return &DEK{key: key}
}

func (d *DEK) Bytes() []byte {
	return d.key
}

func (d *DEK) Encrypt(plaintext []byte, keyName string) ([]byte, []byte, error) {
	ciphertext, err := EncryptAESGCM(d.key, plaintext, []byte(keyName))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	aesgcm, err := NewAESGCM(d.key)
	if err != nil {
		return nil, nil, err
	}
	nonce := ciphertext[:aesgcm.NonceSize()]
	return nonce, ciphertext, nil
}

func (d *DEK) Decrypt(ciphertext []byte, keyName string) ([]byte, error) {
	plaintext, err := DecryptAESGCM(d.key, ciphertext, []byte(keyName))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return plaintext, nil
}
