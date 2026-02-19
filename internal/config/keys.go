package config

import (
	"fmt"
	"path/filepath"

	"github.com/xmazu/openenvx/internal/storage"
)

const KeysFileName = "keys.yaml"

type WorkspaceKey struct {
	Public  string `yaml:"public"`
	Private string `yaml:"private"`
}

type KeysFile struct {
	Workspaces map[string]WorkspaceKey `yaml:"workspaces"`
	file       *storage.YAMLFile
}

func KeysPath() string {
	return filepath.Join(ConfigDir(), KeysFileName)
}

func LoadKeysFile() (*KeysFile, error) {
	path := KeysPath()
	file := storage.NewYAMLFile(path)

	kf := &KeysFile{
		Workspaces: make(map[string]WorkspaceKey),
		file:       file,
	}

	if err := file.LoadOrCreate(kf); err != nil {
		return nil, err
	}

	if kf.Workspaces == nil {
		kf.Workspaces = make(map[string]WorkspaceKey)
	}

	return kf, nil
}

func (k *KeysFile) Save() error {
	return k.file.Save(k)
}

func (k *KeysFile) Get(workspacePath string) (WorkspaceKey, bool) {
	key, ok := k.Workspaces[workspacePath]
	return key, ok
}

func (k *KeysFile) Set(workspacePath string, publicKey, privateKey string) error {
	if workspacePath == "" {
		return fmt.Errorf("workspace path must not be empty")
	}
	if publicKey == "" || privateKey == "" {
		return fmt.Errorf("public and private keys must not be empty")
	}

	k.Workspaces[workspacePath] = WorkspaceKey{
		Public:  publicKey,
		Private: privateKey,
	}

	return k.Save()
}

func (k *KeysFile) List() []string {
	paths := make([]string, 0, len(k.Workspaces))
	for path := range k.Workspaces {
		paths = append(paths, path)
	}
	return paths
}
