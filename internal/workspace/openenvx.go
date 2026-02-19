package workspace

import (
	"fmt"
	"path/filepath"

	"github.com/xmazu/openenvx/internal/storage"
)

const WorkspaceFileName = ".openenvx.yaml"

type ScanConfig struct {
	ExcludeFiles []string `yaml:"exclude_files"`
}

type WorkspaceConfig struct {
	PublicKey string      `yaml:"public_key"`
	Scan      *ScanConfig `yaml:"scan"`
}

func ReadWorkspaceFile(dir string) (*WorkspaceConfig, error) {
	path := filepath.Join(dir, WorkspaceFileName)
	file := storage.NewYAMLFile(path)

	var cfg WorkspaceConfig
	if err := file.Load(&cfg); err != nil {
		return nil, fmt.Errorf("no .openenvx.yaml file in %s", dir)
	}
	return &cfg, nil
}

func WriteWorkspaceFile(dir string, cfg *WorkspaceConfig) error {
	path := filepath.Join(dir, WorkspaceFileName)
	file := storage.NewYAMLFile(path)
	return file.SaveWithPerm(cfg, 0644)
}

func WorkspaceFileExists(dir string) bool {
	path := filepath.Join(dir, WorkspaceFileName)
	return storage.NewYAMLFile(path).Exists()
}
