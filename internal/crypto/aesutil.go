package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

var (
	ErrInvalidKeySize  = errors.New("invalid key size: must be 16, 24, or 32 bytes")
	ErrCiphertextShort = errors.New("ciphertext too short")
	ErrInvalidNonce    = errors.New("invalid nonce")
)

type AESGCM struct {
	key   []byte
	block cipher.Block
	gcm   cipher.AEAD
}

func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &AESGCM{
		key:   key,
		block: block,
		gcm:   gcm,
	}, nil
}

func (a *AESGCM) NonceSize() int {
	return a.gcm.NonceSize()
}

func (a *AESGCM) Seal(plaintext, additionalData []byte) ([]byte, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := a.gcm.Seal(nonce, nonce, plaintext, additionalData)
	return ciphertext, nil
}

func (a *AESGCM) Open(ciphertext, additionalData []byte) ([]byte, error) {
	nonceSize := a.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextShort
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := a.gcm.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

func EncryptAESGCM(key, plaintext, additionalData []byte) ([]byte, error) {
	aesgcm, err := NewAESGCM(key)
	if err != nil {
		return nil, err
	}
	return aesgcm.Seal(plaintext, additionalData)
}

func DecryptAESGCM(key, ciphertext, additionalData []byte) ([]byte, error) {
	aesgcm, err := NewAESGCM(key)
	if err != nil {
		return nil, err
	}
	return aesgcm.Open(ciphertext, additionalData)
}
