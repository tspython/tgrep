package search

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var (
	terminalFormatter = formatters.Get("terminal16m")
	highlightStyle    = styles.Get("github-dark")
)

func init() {
	if terminalFormatter == nil {
		terminalFormatter = formatters.Get("terminal")
	}
	if highlightStyle == nil {
		highlightStyle = styles.Fallback
	}
}

func highlightForTerminal(filePath, line string) string {
	if terminalFormatter == nil || line == "" {
		return line
	}

	lexer := lexers.Match(filePath)
	if lexer == nil {
		lexer = lexers.Analyse(line)
	}
	if lexer == nil {
		return line
	}

	iterator, err := lexer.Tokenise(nil, line)
	if err != nil {
		return line
	}

	var buf bytes.Buffer
	if err := terminalFormatter.Format(&buf, highlightStyle, iterator); err != nil {
		return line
	}

	return strings.TrimRight(buf.String(), "\r\n")
}
