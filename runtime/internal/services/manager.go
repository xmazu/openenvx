package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct {
	runtime Runtime
	root    string
}

func NewManager(root string) (*Manager, error) {
	rt := NewDocker()
	if !rt.IsAvailable() {
		return nil, fmt.Errorf("docker compose not available")
	}
	return &Manager{runtime: rt, root: root}, nil
}

func (m *Manager) composeFile() string {
	return filepath.Join(m.root, ".openenvx", "services.yaml")
}

func (m *Manager) hasConfig() bool {
	_, err := os.Stat(m.composeFile())
	return err == nil
}

func (m *Manager) Start(ctx context.Context) error {
	if !m.hasConfig() {
		return fmt.Errorf("no services.yaml found in .openenvx/")
	}
	return m.runtime.Start(ctx, m.composeFile())
}

func (m *Manager) Stop(ctx context.Context) error {
	if !m.hasConfig() {
		return nil
	}
	return m.runtime.Stop(ctx, m.composeFile())
}

func (m *Manager) Status(ctx context.Context) ([]Status, error) {
	if !m.hasConfig() {
		return []Status{}, nil
	}
	return m.runtime.Status(ctx, m.composeFile())
}
