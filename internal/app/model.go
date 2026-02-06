package app

import (
	"github.com/tspython/tgrep/internal/domain"
	"github.com/tspython/tgrep/internal/search"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	queryFocus    inputFocus
	query         string
	fileQuery     string
	results       []domain.SearchResult
	selected      int
	listOffset    int
	searching     bool
	err           error
	preview       string
	width, height int
}

type searchFinishedMsg []domain.SearchResult

type errMsg struct{ error }

type inputFocus int

const (
	focusQuery inputFocus = iota
	focusFiles
)

func NewModel() Model {
	return Model{
		queryFocus: focusQuery,
		query:      "",
		fileQuery:  "*",
		results:    []domain.SearchResult{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			m.toggleFocus()
		case "enter":
			if m.query != "" && !m.searching {
				m.searching = true
				m.err = nil
				return m, performSearch(m.query, m.fileQuery)
			}
		case "up":
			if m.selected > 0 {
				m.selected--
				m.keepSelectionVisible()
				m.refreshPreview()
			}
		case "down":
			if m.selected < len(m.results)-1 {
				m.selected++
				m.keepSelectionVisible()
				m.refreshPreview()
			}
		case "backspace":
			m.backspaceFocusedInput()
		case "space":
			m.appendToFocusedInput(" ")
		case "ctrl+u":
			m.clearFocusedInput()
		default:
			if msg.Type == tea.KeyRunes {
				m.appendToFocusedInput(string(msg.Runes))
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.keepSelectionVisible()
	case searchFinishedMsg:
		m.searching = false
		m.results = []domain.SearchResult(msg)
		m.selected = 0
		m.listOffset = 0
		m.refreshPreview()
	case errMsg:
		m.err = msg
		m.searching = false
	}

	return m, nil
}

func (m *Model) keepSelectionVisible() {
	visibleRows := m.resultsPanelHeight() - 3
	if visibleRows < 1 {
		visibleRows = 1
	}

	if m.selected < m.listOffset {
		m.listOffset = m.selected
	}
	if m.selected >= m.listOffset+visibleRows {
		m.listOffset = m.selected - visibleRows + 1
	}
	if m.listOffset < 0 {
		m.listOffset = 0
	}
}

func (m *Model) refreshPreview() {
	if len(m.results) == 0 || m.selected < 0 || m.selected >= len(m.results) {
		m.preview = "Select a result to preview context"
		return
	}
	m.preview = search.Preview(m.results[m.selected], 2)
}

func (m *Model) toggleFocus() {
	if m.queryFocus == focusQuery {
		m.queryFocus = focusFiles
		return
	}
	m.queryFocus = focusQuery
}

func (m *Model) appendToFocusedInput(text string) {
	if m.queryFocus == focusFiles {
		m.fileQuery += text
		return
	}
	m.query += text
}

func (m *Model) backspaceFocusedInput() {
	if m.queryFocus == focusFiles {
		if len(m.fileQuery) > 0 {
			m.fileQuery = m.fileQuery[:len(m.fileQuery)-1]
		}
		return
	}
	if len(m.query) > 0 {
		m.query = m.query[:len(m.query)-1]
	}
}

func (m *Model) clearFocusedInput() {
	if m.queryFocus == focusFiles {
		m.fileQuery = ""
		return
	}
	m.query = ""
}

func performSearch(query, fileQuery string) tea.Cmd {
	return func() tea.Msg {
		results, err := search.Files(query, fileQuery)
		if err != nil {
			return errMsg{error: err}
		}
		return searchFinishedMsg(results)
	}
}
