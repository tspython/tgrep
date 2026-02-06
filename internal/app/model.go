package app

import (
	"tgrep/internal/domain"
	"tgrep/internal/search"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	query         string
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

func NewModel() Model {
	return Model{query: "", results: []domain.SearchResult{}}
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
		case "enter":
			if m.query != "" && !m.searching {
				m.searching = true
				m.err = nil
				return m, performSearch(m.query)
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
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
			}
		case "space":
			m.query += " "
		default:
			if msg.Type == tea.KeyRunes {
				m.query += string(msg.Runes)
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
	visibleRows := m.resultsPanelHeight() - 1
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

func performSearch(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := search.Files(query)
		if err != nil {
			return errMsg{error: err}
		}
		return searchFinishedMsg(results)
	}
}
