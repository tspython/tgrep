package search

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/tspython/tgrep/internal/domain"
)

func Files(query, fileQuery string) ([]domain.SearchResult, error) {
	results := make([]domain.SearchResult, 0)
	pattern, err := regexp.Compile(query)
	useRegex := err == nil
	filters, err := parseFileFilters(fileQuery)
	if err != nil {
		return nil, err
	}

	walkErr := filepath.Walk(".", func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() && strings.HasPrefix(path, ".") && path != "." {
			return filepath.SkipDir
		}
		if info.IsDir() || shouldSkipFile(path) {
			return nil
		}
		if !matchesFileFilters(path, filters) {
			return nil
		}

		fileMatches, fileErr := inFile(path, query, pattern, useRegex)
		if fileErr != nil {
			return nil
		}

		results = append(results, fileMatches...)
		return nil
	})

	return results, walkErr
}

func parseFileFilters(fileQuery string) ([]string, error) {
	raw := strings.TrimSpace(fileQuery)
	if raw == "" || raw == "*" {
		return nil, nil
	}

	tokens := strings.Split(raw, ",")
	if len(tokens) == 1 {
		tokens = strings.Fields(raw)
	}
	filters := make([]string, 0, len(tokens))

	for _, token := range tokens {
		filter := filepath.ToSlash(strings.TrimSpace(token))
		if filter == "" {
			continue
		}
		if filter == "*" {
			return nil, nil
		}
		if _, err := path.Match(filter, "x"); err != nil {
			return nil, fmt.Errorf("invalid file filter %q: %w", filter, err)
		}
		filters = append(filters, filter)
	}

	return filters, nil
}

func matchesFileFilters(filePath string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}

	relPath := filepath.ToSlash(strings.TrimPrefix(filePath, "./"))
	base := filepath.Base(filePath)

	for _, filter := range filters {
		var (
			match bool
			err   error
		)

		if strings.Contains(filter, "/") {
			match, err = path.Match(filter, relPath)
		} else {
			match, err = path.Match(filter, base)
		}

		if err == nil && match {
			return true
		}
	}

	return false
}

func Preview(result domain.SearchResult, context int) string {
	file, err := os.Open(result.File)
	if err != nil {
		return "Could not open preview"
	}
	defer file.Close()

	start := max(1, result.Line-context)
	end := result.Line + context

	lines := make([]string, 0, end-start+1)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum < start {
			continue
		}
		if lineNum > end {
			break
		}

		prefix := "  "
		if lineNum == result.Line {
			prefix = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%4d | %s", prefix, lineNum, scanner.Text()))
	}

	if len(lines) == 0 {
		return "No preview available"
	}
	return strings.Join(lines, "\n")
}

func shouldSkipFile(path string) bool {
	skipDirs := []string{".git", "node_modules", "vendor", ".vscode", "target", "dist"}
	for _, dir := range skipDirs {
		if strings.Contains(path, dir+string(os.PathSeparator)) {
			return true
		}
	}

	skipExts := []string{".exe", ".dll", ".so", ".dylib", ".bin", ".png", ".jpg", ".jpeg", ".gif", ".zip", ".tar", ".gz"}
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(skipExts, ext)
}

func inFile(filePath, query string, pattern *regexp.Regexp, useRegex bool) ([]domain.SearchResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	results := make([]domain.SearchResult, 0)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	queryLower := strings.ToLower(query)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if useRegex {
			matches := pattern.FindAllStringIndex(line, -1)
			for _, match := range matches {
				results = append(results, domain.SearchResult{
					File:    filePath,
					Line:    lineNum,
					Column:  match[0] + 1,
					Content: line,
				})
			}
			continue
		}

		lower := strings.ToLower(line)
		idx := 0
		for {
			found := strings.Index(lower[idx:], queryLower)
			if found < 0 {
				break
			}

			column := idx + found + 1
			results = append(results, domain.SearchResult{
				File:    filePath,
				Line:    lineNum,
				Column:  column,
				Content: line,
			})

			idx += found + 1
			if idx >= len(lower) {
				break
			}
		}
	}

	return results, scanner.Err()
}
