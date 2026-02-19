package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"os"

	"filippo.io/age"

	"github.com/xmazu/openenvx/internal/config"
)

type AsymmetricStrategy struct {
	publicKey string
	identity  *age.X25519Identity
	cachedKey *MasterKey
}

func NewAsymmetricStrategy(identity *age.X25519Identity) *AsymmetricStrategy {
	var publicKey string
	if identity != nil {
		publicKey = identity.Recipient().String()
	}
	return &AsymmetricStrategy{
		publicKey: publicKey,
		identity:  identity,
	}
}

func (a *AsymmetricStrategy) GetMasterKey() (*MasterKey, error) {
	if a.cachedKey != nil {
		return a.cachedKey, nil
	}

	if a.identity == nil {
		return nil, fmt.Errorf("private key not available - set OPENENVX_PRIVATE_KEY environment variable")
	}

	identityBytes := []byte(a.identity.String())
	h := hmac.New(sha256.New, identityBytes)
	h.Write([]byte("openenvx-master-key-derivation"))
	key := h.Sum(nil)[:masterKeySize]

	a.cachedKey = &MasterKey{key: key}
	return a.cachedKey, nil
}

func GenerateAgeKeyPair() (*age.X25519Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate age identity: %w", err)
	}
	return identity, nil
}

func ParseAgeIdentity(identityStr string) (*age.X25519Identity, error) {
	identity, err := age.ParseX25519Identity(identityStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse age identity: %w", err)
	}
	return identity, nil
}

func GetPrivateKey(workspacePath string) (*age.X25519Identity, bool) {
	if keyStr := os.Getenv("OPENENVX_PRIVATE_KEY"); keyStr != "" {
		identity, err := ParseAgeIdentity(keyStr)
		if err == nil {
			return identity, true
		}
	}

	if workspacePath != "" {
		keysFile, err := config.LoadKeysFile()
		if err == nil {
			if key, ok := keysFile.Get(workspacePath); ok && key.Private != "" {
				identity, err := ParseAgeIdentity(key.Private)
				if err == nil {
					return identity, true
				}
			}
		}
	}

	return nil, false
}
