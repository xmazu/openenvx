package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/xmazu/openenvx/internal/crypto"
	"github.com/xmazu/openenvx/internal/scanner"
	"github.com/xmazu/openenvx/internal/tui"
	"github.com/xmazu/openenvx/internal/workspace"
	"golang.org/x/sync/errgroup"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for secrets in code files",
	Long:  `Scan the current directory (or specified path) for potential secrets and API keys. If --path is not specified and a workspace is detected, defaults to workspace root.`,
	RunE:  runScan,
}

var scanPath string
var scanExclude []string
var scanExcludeFiles []string
var scanIgnoreMismatch string
var scanOutput string

func init() {
	scanCmd.Flags().StringVarP(&scanPath, "path", "p", "", "Path to scan (default: workspace root if detected, else current directory)")
	scanCmd.Flags().StringSliceVarP(&scanExclude, "exclude", "e", []string{".git", "node_modules", "vendor"}, "Directories to exclude")
	scanCmd.Flags().StringSliceVarP(&scanExcludeFiles, "exclude-files", "E", nil, "Additional file names or glob patterns to exclude (matched against base name). Defaults from .openenvx.yaml scan.exclude_files, or go.mod, go.sum if no .openenvx.yaml")
	scanCmd.Flags().StringVar(&scanIgnoreMismatch, "ignore-mismatch", "off", "Deprecated: no longer has any effect since only .gitignore is used for exclusions")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "Output format: default (human), json (machine-readable for editors)")
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	matches := []scanner.Match{}
	var errOut io.Writer = os.Stderr
	if cmd != nil {
		errOut = cmd.ErrOrStderr()
	}

	effectivePath := scanPath
	if effectivePath == "" {
		wsRoot, err := workspace.FindRoot(".")
		if err == nil && workspace.IsWorkspace(wsRoot) {
			effectivePath = wsRoot
		} else {
			effectivePath = "."
		}
	}

	scanRoot, err := filepath.Abs(effectivePath)
	if err != nil {
		return fmt.Errorf("resolve scan path: %w", err)
	}

	effectiveExcludeFiles := effectiveScanExcludeFiles(scanRoot)
	effectiveExcludeFiles = append(effectiveExcludeFiles, scanExcludeFiles...)

	switch scanIgnoreMismatch {
	case "off", "warn", "error":
	default:
		return fmt.Errorf("invalid --ignore-mismatch %q: must be off, warn, or error", scanIgnoreMismatch)
	}

	gitignoreMatcher, err := scanner.LoadGitignore(scanRoot)
	if err != nil {
		return fmt.Errorf("load .gitignore: %w", err)
	}

	var filesToScan []string
	err = filepath.WalkDir(effectivePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(errOut, "Warning: cannot access %s: %v\n", path, err)
			return nil
		}

		absPath, absErr := filepath.Abs(path)
		if absErr != nil {
			absPath = path
		}
		relPath, relErr := filepath.Rel(scanRoot, absPath)
		if relErr != nil {
			relPath = path
		}
		relPath = filepath.ToSlash(relPath)

		if d.IsDir() {
			for _, exclude := range scanExclude {
				if strings.Contains(path, exclude) {
					return filepath.SkipDir
				}
			}
			if gitignoreMatcher != nil && gitignoreMatcher.ShouldIgnore(relPath, true) {
				return filepath.SkipDir
			}
			return nil
		}

		if gitignoreMatcher != nil && gitignoreMatcher.ShouldIgnore(relPath, false) {
			return nil
		}

		if scanner.FileMatchesExclude(relPath, effectiveExcludeFiles) {
			return nil
		}

		if scanner.IsBinaryFile(path) {
			return nil
		}

		filesToScan = append(filesToScan, path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan: %w", err)
	}

	numWorkers := runtime.NumCPU()
	if numWorkers < 2 {
		numWorkers = 2
	}

	var mu sync.Mutex
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(numWorkers)

	for _, path := range filesToScan {
		path := path
		select {
		case <-ctx.Done():
			break
		default:
		}

		g.Go(func() error {
			fileMatches, scanErr := scanFile(path)
			if scanErr != nil {
				fmt.Fprintf(errOut, "Warning: cannot scan %s: %v\n", path, scanErr)
				return nil
			}
			mu.Lock()
			matches = append(matches, fileMatches...)
			mu.Unlock()
			return nil
		})
	}

	if waitErr := g.Wait(); waitErr != nil {
		return fmt.Errorf("scan error: %w", waitErr)
	}

	matches = dedupeMatches(matches)

	if scanOutput == "json" {
		return outputScanJSON(matches, scanRoot, cmd.OutOrStdout())
	}

	if len(matches) == 0 {
		return nil
	}

	fileEntries := reduceMatchesByFile(matches)

	high := 0
	medium := 0
	low := 0
	for _, m := range fileEntries {
		switch m.Pattern.Severity {
		case "high":
			high++
		case "medium":
			medium++
		case "low":
			low++
		}
	}

	fmt.Printf("%s %d files with potential secrets:\n", tui.Header("Found"), len(fileEntries))
	fmt.Printf("  %s %d files | %s %d files | %s %d files\n", tui.SeverityHigh.Render("High:"), high, tui.SeverityMedium.Render("Medium:"), medium, tui.SeverityLow.Render("Low:"), low)
	fmt.Println()

	for _, match := range fileEntries {
		var sevStyle lipgloss.Style
		switch match.Pattern.Severity {
		case "high":
			sevStyle = tui.SeverityCritical
		case "medium":
			sevStyle = tui.SeverityMedium
		default:
			sevStyle = tui.SeverityLow
		}

		fmt.Printf("%s %s\n", sevStyle.Render("["+strings.ToUpper(match.Pattern.Severity)+"]"), tui.Label(match.Pattern.Name))
		fmt.Printf("  %s %s:%d\n", tui.Muted("File:"), match.File, match.Line)
		fmt.Printf("  %s %s\n", tui.Muted("Match:"), truncateString(match.Match, 50))
		fmt.Printf("  %s %s\n", tui.Muted("Context:"), truncateString(match.Context, 80))
		fmt.Println()
	}

	if high > 0 {
		return fmt.Errorf("found %d high-severity secrets", high)
	}

	return nil
}

type scanJSONMatch struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Pattern  string `json:"pattern"`
	Severity string `json:"severity"`
	Match    string `json:"match"`
	Context  string `json:"context"`
}

func outputScanJSON(matches []scanner.Match, scanRoot string, w io.Writer) error {
	list := make([]scanJSONMatch, 0, len(matches))
	for _, m := range matches {
		abs := m.File
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(scanRoot, abs)
		}
		if a, err := filepath.Abs(abs); err == nil {
			abs = a
		}
		list = append(list, scanJSONMatch{
			File:     filepath.ToSlash(abs),
			Line:     m.Line,
			Column:   m.Column,
			Pattern:  m.Pattern.Name,
			Severity: m.Pattern.Severity,
			Match:    truncateString(m.Match, 50),
			Context:  truncateString(m.Context, 80),
		})
	}
	out := struct {
		Matches []scanJSONMatch `json:"matches"`
	}{Matches: list}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode scan JSON: %w", err)
	}
	if f, ok := w.(interface{ Sync() error }); ok {
		_ = f.Sync()
	}
	return nil
}

func isEnvFile(path string) bool {
	base := filepath.Base(path)
	return base == ".env" || strings.HasPrefix(base, ".env.")
}

func shouldSuppressEnvMatch(path, line string) bool {
	if !isEnvFile(path) {
		return false
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}
	idx := strings.Index(line, "=")
	if idx <= 0 {
		return false
	}
	value := strings.TrimSpace(line[idx+1:])
	if i := strings.Index(value, " #"); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}
	if !strings.HasPrefix(value, crypto.EncryptedValuePrefix) {
		return false
	}
	_, err := crypto.ParseEncryptedValue(value)
	return err == nil
}

func scanFile(path string) ([]scanner.Match, error) {
	matches := []scanner.Match{}

	file, err := os.Open(path)
	if err != nil {
		return matches, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return matches, err
	}
	if stat.Size() > 10*1024*1024 {
		return matches, nil
	}

	scannerBuf := bufio.NewScanner(file)
	scannerBuf.Buffer(make([]byte, 1024*1024), 1024*1024)
	lineNum := 0

	for scannerBuf.Scan() {
		lineNum++
		line := scannerBuf.Text()
		if len(line) > 10000 {
			continue
		}

		for _, pattern := range scanner.Patterns {
			if locs := pattern.Regex.FindAllStringIndex(line, -1); locs != nil {
				for _, loc := range locs {
					matchStr := line[loc[0]:loc[1]]

					if shouldSuppressEnvMatch(path, line) {
						continue
					}

					match := scanner.Match{
						Pattern: pattern,
						File:    path,
						Line:    lineNum,
						Column:  loc[0],
						Match:   matchStr,
						Context: line,
					}
					matches = append(matches, match)
				}
			}
		}
	}

	if err := scannerBuf.Err(); err != nil {
		if strings.Contains(err.Error(), "token too long") {
			return matches, nil
		}
		return matches, err
	}
	return matches, nil
}

func severityRank(s string) int {
	switch s {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func matchesOverlap(a, b scanner.Match) bool {
	if a.File != b.File {
		return false
	}
	if a.Line != b.Line {
		return false
	}
	aStart := a.Column
	aEnd := a.Column + len(a.Match)
	bStart := b.Column
	bEnd := b.Column + len(b.Match)
	return aStart < bEnd && bStart < aEnd
}

func dedupeMatches(matches []scanner.Match) []scanner.Match {
	if len(matches) == 0 {
		return matches
	}
	result := make([]scanner.Match, 0, len(matches))
	for _, m := range matches {
		overlaps := false
		for i := range result {
			if matchesOverlap(m, result[i]) {
				overlaps = true
				if severityRank(m.Pattern.Severity) > severityRank(result[i].Pattern.Severity) {
					result[i] = m
				}
				break
			}
		}
		if !overlaps {
			result = append(result, m)
		}
	}
	return result
}

func reduceMatchesByFile(matches []scanner.Match) []scanner.Match {
	byFile := make(map[string][]scanner.Match)
	for _, m := range matches {
		abs, err := filepath.Abs(m.File)
		if err != nil {
			abs = m.File
		}
		key := filepath.ToSlash(abs)
		byFile[key] = append(byFile[key], m)
	}
	result := make([]scanner.Match, 0, len(byFile))
	for _, fileMatches := range byFile {
		if len(fileMatches) == 0 {
			continue
		}
		maxRank := 0
		var best scanner.Match
		for _, m := range fileMatches {
			r := severityRank(m.Pattern.Severity)
			if r > maxRank {
				maxRank = r
				best = m
			}
		}
		result = append(result, best)
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func effectiveScanExcludeFiles(scanRoot string) []string {
	out := append([]string{}, scanner.DefaultScanExcludedFiles()...)
	wc, err := workspace.ReadWorkspaceFile(scanRoot)
	if err == nil && wc.Scan != nil && len(wc.Scan.ExcludeFiles) > 0 {
		out = append(out, wc.Scan.ExcludeFiles...)
	}
	return out
}
