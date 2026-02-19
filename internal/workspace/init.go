package workspace

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/envfile"
	"github.com/xmazu/openenvx/internal/scanner"
)

type EncryptConfig struct {
	PublicKey         string
	NoEncryptDetector *envfile.NoEncryptDetector
	Envelope          *crypto.Envelope
}

type FileResult struct {
	Encrypted        int
	Skipped          int
	AlreadyEncrypted bool
	CommentedSecrets []envfile.CommentedSecret
	Err              error
}

type CommentedSecret = envfile.CommentedSecret

func IsEnvFileEncrypted(path string) (bool, error) {
	envFile, err := envfile.Load(path)
	if err != nil {
		return false, err
	}
	for _, key := range envFile.Keys() {
		val, _ := envFile.Get(key)
		if strings.HasPrefix(val, crypto.EncryptedValuePrefix) {
			return true, nil
		}
	}
	return false, nil
}

func ErrEncryptedEnvWithoutOpenenvx(root string) error {
	if WorkspaceFileExists(root) {
		return nil
	}
	files, err := ListEnvFiles(root)
	if err != nil {
		return err
	}
	var encrypted []string
	for _, f := range files {
		ok, err := IsEnvFileEncrypted(f)
		if err != nil {
			return err
		}
		if ok {
			rel, _ := filepath.Rel(root, f)
			if rel == "" {
				rel = f
			}
			encrypted = append(encrypted, rel)
		}
	}
	if len(encrypted) == 0 {
		return nil
	}
	return fmt.Errorf("encrypted .env file(s) found but no .openenvx.yaml at workspace root: %v. Restore .openenvx.yaml or remove encrypted values before running init", encrypted)
}

func EncryptEnvFile(path string, cfg *EncryptConfig) (*envfile.File, *FileResult) {
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
	if isEncrypted {
		result.AlreadyEncrypted = true
		return envFile, result
	}

	result.CommentedSecrets = envFile.DetectCommentedSecrets(cfg.NoEncryptDetector, scanner.Patterns)

	vars := make(map[string]string)
	for _, key := range envFile.Keys() {
		val, ok := envFile.Get(key)
		if ok && val != "" {
			vars[key] = val
		}
	}

	for key, val := range vars {
		if cfg.NoEncryptDetector.ShouldSkip(key, val) {
			envFile.Set(key, val)
			result.Skipped++
		} else {
			enc, err := cfg.Envelope.Encrypt([]byte(val), key)
			if err != nil {
				result.Err = fmt.Errorf("encrypt %s in %s: %w", key, path, err)
				return envFile, result
			}
			envFile.Set(key, enc.String())
			result.Encrypted++
		}
	}

	return envFile, result
}
