package search

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/tspython/tgrep/internal/domain"
)

func Files(query, fileQuery string) ([]domain.SearchResult, error) {
	filters := parseFileFilters(fileQuery)

	args := []string{
		"--json",
		"--line-number",
		"--column",
		"--color", "never",
		"--no-heading",
		"--with-filename",
		"--smart-case",
	}

	if isValidRegex(query) {
		args = append(args, query)
	} else {
		args = append(args, "-F", query)
	}

	for _, filter := range filters {
		args = append(args, "-g", filter)
	}

	args = append(args, ".")

	stdout, stderr, err := runRipgrep(args)
	if err != nil {
		return nil, formatRipgrepError(err, stderr)
	}

	return parseRipgrepOutput(stdout)
}

func parseFileFilters(fileQuery string) []string {
	raw := strings.TrimSpace(fileQuery)
	if raw == "" || raw == "*" {
		return nil
	}

	tokens := strings.Split(raw, ",")
	if len(tokens) == 1 {
		tokens = strings.Fields(raw)
	}

	filters := make([]string, 0, len(tokens))
	for _, token := range tokens {
		filter := strings.TrimSpace(token)
		if filter == "" || filter == "*" {
			continue
		}
		filters = append(filters, filter)
	}

	return filters
}

func isValidRegex(query string) bool {
	_, err := regexp.Compile(query)
	return err == nil
}

func runRipgrep(args []string) ([]byte, []byte, error) {
	cmd := exec.Command("rg", args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil, nil
		}
	}

	return stdout.Bytes(), stderr.Bytes(), err
}

func formatRipgrepError(err error, stderr []byte) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("ripgrep not found: install 'rg' and make sure it is in PATH")
	}

	message := strings.TrimSpace(string(stderr))
	if message == "" {
		message = err.Error()
	}

	return fmt.Errorf("ripgrep failed: %s", message)
}

func parseRipgrepOutput(output []byte) ([]domain.SearchResult, error) {
	results := make([]domain.SearchResult, 0)

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event rgEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		if event.Type != "match" {
			continue
		}

		if event.Data.Path.Text == "" {
			continue
		}

		content := strings.TrimRight(event.Data.Lines.Text, "\r\n")
		for _, submatch := range event.Data.Submatches {
			results = append(results, domain.SearchResult{
				File:    event.Data.Path.Text,
				Line:    event.Data.LineNumber,
				Column:  submatch.Start + 1,
				Content: content,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
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

		highlighted := highlightForTerminal(result.File, scanner.Text())
		lines = append(lines, fmt.Sprintf("%s%4d | %s", prefix, lineNum, highlighted))
	}

	if len(lines) == 0 {
		return "No preview available"
	}
	return strings.Join(lines, "\n")
}

type rgEvent struct {
	Type string      `json:"type"`
	Data rgEventData `json:"data"`
}

type rgEventData struct {
	Path       rgText       `json:"path"`
	Lines      rgText       `json:"lines"`
	LineNumber int          `json:"line_number"`
	Submatches []rgSubmatch `json:"submatches"`
}

type rgSubmatch struct {
	Start int `json:"start"`
}

type rgText struct {
	Text string `json:"text"`
}
