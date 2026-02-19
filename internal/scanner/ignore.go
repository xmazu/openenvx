package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type ignoreRule struct {
	pattern string // for doublestar; forward slashes
	dirOnly bool   // trailing slash: match directories only
	anchor  bool   // leading slash: match only at scan root
}

type IgnoreMatcher struct {
	rules []ignoreRule
}

func parseIgnoreFile(path string) ([]ignoreRule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var rules []ignoreRule
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		dirOnly := strings.HasSuffix(line, "/")
		if dirOnly {
			line = strings.TrimSuffix(line, "/")
		}

		anchor := strings.HasPrefix(line, "/")
		if anchor {
			line = strings.TrimPrefix(line, "/")
		}

		if line == "" {
			continue
		}

		line = filepath.ToSlash(line)

		rules = append(rules, ignoreRule{pattern: line, dirOnly: dirOnly, anchor: anchor})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return rules, nil
}

func LoadGitignore(root string) (*IgnoreMatcher, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve scan root: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat scan root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scan root is not a directory: %s", absRoot)
	}

	path := filepath.Join(absRoot, ".gitignore")
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat .gitignore: %w", err)
	}

	rules, err := parseIgnoreFile(path)
	if err != nil {
		return nil, err
	}

	if len(rules) == 0 {
		return nil, nil
	}

	return &IgnoreMatcher{rules: rules}, nil
}

func (m *IgnoreMatcher) ShouldIgnore(relPath string, isDir bool) bool {
	if m == nil || len(m.rules) == 0 {
		return false
	}

	relPath = filepath.ToSlash(relPath)
	relPath = strings.TrimPrefix(relPath, "/")

	for _, r := range m.rules {

		if r.anchor {
			if relPath == r.pattern {
				return true
			}
			if isDir && relPath+"/" == r.pattern+"/" {
				return true
			}
			if strings.HasPrefix(relPath, r.pattern+"/") {
				return true
			}
			continue
		}

		if r.dirOnly && !isDir {
			under := strings.TrimPrefix(r.pattern, "**/")
			if under != "" && (strings.Contains(relPath, "/"+under+"/") || strings.HasPrefix(relPath, under+"/")) {
				return true
			}
			continue
		}

		matched, err := doublestar.Match(r.pattern, relPath)
		if err != nil {
			continue
		}
		if matched {
			return true
		}

		if r.dirOnly && isDir && (relPath == r.pattern || strings.HasPrefix(relPath, r.pattern+"/")) {
			return true
		}
	}

	return false
}

func DefaultScanExcludedFiles() []string {
	return []string{
		".openenvx.yaml",
		"go.mod",
		"go.sum",
		"package-lock.json",
		"pnpm-lock.yaml",
		"pnpm-lock.yml",
		"yarn.lock",
		"bun.lock",
		"bun.lockb",
		"npm-shrinkwrap.json",
	}
}

var binaryExtensions = map[string]bool{
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".ico":   true,
	".svg":   true,
	".webp":  true,
	".bmp":   true,
	".tiff":  true,
	".tif":   true,
	".pdf":   true,
	".zip":   true,
	".gz":    true,
	".tar":   true,
	".rar":   true,
	".7z":    true,
	".mp3":   true,
	".mp4":   true,
	".wav":   true,
	".avi":   true,
	".mov":   true,
	".mkv":   true,
	".woff":  true,
	".woff2": true,
	".ttf":   true,
	".eot":   true,
	".otf":   true,
}

func IsBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return binaryExtensions[ext]
}

func FileMatchesExclude(relPath string, patterns []string) bool {
	for _, p := range patterns {
		if p == "" {
			continue
		}
		patternPath := filepath.ToSlash(p)
		relPathClean := filepath.ToSlash(relPath)

		matched, err := doublestar.Match(patternPath, relPathClean)
		if err == nil && matched {
			return true
		}

		base := filepath.Base(relPathClean)
		if patternPath == base {
			return true
		}
		matched, err = filepath.Match(patternPath, base)
		if err == nil && matched {
			return true
		}
	}
	return false
}
