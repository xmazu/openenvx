package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type YAMLFile struct {
	path string
}

func NewYAMLFile(path string) *YAMLFile {
	return &YAMLFile{path: path}
}

func (y *YAMLFile) Path() string {
	return y.path
}

func (y *YAMLFile) Exists() bool {
	_, err := os.Stat(y.path)
	return err == nil
}

func (y *YAMLFile) Load(dest interface{}) error {
	data, err := os.ReadFile(y.path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", y.path)
		}
		return fmt.Errorf("read file: %w", err)
	}

	if err := yaml.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	return nil
}

func (y *YAMLFile) LoadOrCreate(dest interface{}) error {
	data, err := os.ReadFile(y.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read file: %w", err)
	}

	if err := yaml.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	return nil
}

func (y *YAMLFile) Save(data interface{}) error {
	return y.SaveWithPerm(data, 0600)
}

func (y *YAMLFile) SaveWithPerm(data interface{}, perm os.FileMode) error {
	dir := filepath.Dir(y.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	if err := os.WriteFile(y.path, out, perm); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (y *YAMLFile) Delete() error {
	if err := os.Remove(y.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}
