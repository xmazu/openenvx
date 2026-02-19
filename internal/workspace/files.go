package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

var DefaultExcludeDirs = []string{
	".git",
	"node_modules",
	".cache",
	".turbo",
	".next",
}

func ListEnvFiles(root string) ([]string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("get absolute path: %w", err)
	}

	excludeSet := make(map[string]bool)
	for _, d := range DefaultExcludeDirs {
		excludeSet[d] = true
	}

	var files []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if excludeSet[d.Name()] || excludeSet[rel] {
				return filepath.SkipDir
			}
			return nil
		}

		if IsEnvFilename(d.Name()) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
