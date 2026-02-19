package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

var MarkerFiles = []string{
	"pnpm-workspace.yaml",
	"pnpm-lock.yaml",
	"turbo.json",
	"lerna.json",
	"go.work",
	"settings.gradle",
	"settings.gradle.kts",
	"gradlew",
	"gradlew.bat",
	".git",
}

func FindRoot(dir string) (string, error) {
	original, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}
	dir = original

	for {
		for _, marker := range MarkerFiles {
			path := filepath.Join(dir, marker)
			if _, err := os.Stat(path); err == nil {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return original, nil
		}
		dir = parent
	}
}

func IsInitialized(root string) bool {
	return WorkspaceFileExists(root)
}

func IsWorkspace(dir string) bool {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	for _, marker := range MarkerFiles {
		path := filepath.Join(dir, marker)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func FindMarker(root string) string {
	for _, marker := range MarkerFiles {
		path := filepath.Join(root, marker)
		if _, err := os.Stat(path); err == nil {
			return marker
		}
	}
	return ""
}

func FormatMarkerForDisplay(marker string) string {
	if marker == "" {
		return "unknown"
	}
	if marker == ".git" {
		return "git repository"
	}
	return marker
}
