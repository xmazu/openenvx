package workspace

import (
	"fmt"
	"strings"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
)

type DecryptConfig struct {
	Envelope *crypto.Envelope
}

func DecryptEnvFile(path string, cfg *DecryptConfig) (*envfile.File, *FileResult) {
	result := &FileResult{}

	envFile, err := envfile.Load(path)
	if err != nil {
		result.Err = fmt.Errorf("load %s: %w", path, err)
		return envFile, result
	}

	isEncrypted, err := IsEnvFileEncrypted(path)
	if err != nil {
		result.Err = fmt.Errorf("check encryption status: %w", err)
		return envFile, result
	}
	if !isEncrypted {
		result.AlreadyEncrypted = false
		return envFile, result
	}

	for _, key := range envFile.Keys() {
		val, ok := envFile.Get(key)
		if !ok {
			continue
		}

		if !strings.HasPrefix(val, crypto.EncryptedValuePrefix) {
			result.Skipped++
			continue
		}

		ev, err := crypto.ParseEncryptedValue(val)
		if err != nil {
			result.Err = fmt.Errorf("parse encrypted value for %s: %w", key, err)
			return envFile, result
		}

		plaintext, err := cfg.Envelope.Decrypt(ev, key)
		if err != nil {
			result.Err = fmt.Errorf("decrypt %s: %w", key, err)
			return envFile, result
		}

		envFile.Set(key, string(plaintext))
		result.Encrypted++ // Reuse field to count decrypted variables
	}

	return envFile, result
}
