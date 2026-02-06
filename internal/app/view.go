package app

import (
	"fmt"
	"strings"

	"tgrep/internal/domain"

	"github.com/charmbracelet/lipgloss"
)

const (
	vercelBg           = "#000000"
	vercelSurface      = "#111111"
	vercelSurfaceAlt   = "#0A0A0A"
	vercelBorder       = "#2A2A2A"
	vercelBorderStrong = "#3A3A3A"
	vercelText         = "#FAFAFA"
	vercelMuted        = "#A1A1AA"
	vercelAccent       = "#0070F3"
	vercelAccentSoft   = "#0A2540"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Starting tgrep..."
	}

	page := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelBg)).
		Foreground(lipgloss.Color(vercelText)).
		Padding(0, 1)

	header := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelText)).
		Padding(0, 1).
		Bold(true).
		Width(max(10, m.width-4)).
		Render("tgrep  -  fast local search")

	queryTitle := lipgloss.NewStyle().Foreground(lipgloss.Color(vercelMuted)).Bold(true).Render("Query")
	queryValue := m.query
	if queryValue == "" {
		queryValue = "Type a regex or plain text and press Enter"
	}
	queryBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(vercelBorderStrong)).
		Background(lipgloss.Color(vercelSurfaceAlt)).
		Foreground(lipgloss.Color(vercelText)).
		Padding(0, 1).
		Width(max(10, m.width-8)).
		Render(queryValue + "_")

	status := fmt.Sprintf("cwd: .   results: %d", len(m.results))
	if m.searching {
		status = "searching recursively in cwd..."
	}
	statusBar := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelMuted)).
		Padding(0, 1).
		Width(max(10, m.width-4)).
		Render(status)

	resultsPanel := m.renderResultsPanel()
	previewPanel := m.renderPreviewPanel()
	body := lipgloss.JoinHorizontal(lipgloss.Top, resultsPanel, previewPanel)

	footerText := "Enter search | Up/Down move | Esc quit"
	if m.err != nil {
		footerText = "error: " + m.err.Error()
	}
	footer := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelMuted)).
		Padding(0, 1).
		Width(max(10, m.width-4)).
		Render(footerText)

	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		queryTitle,
		queryBox,
		statusBar,
		body,
		footer,
	)

	return page.Render(layout)
}

func (m Model) renderResultsPanel() string {
	panelWidth := max(38, (m.width-6)*62/100)
	height := m.resultsPanelHeight()

	head := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelText)).
		Bold(true).
		Padding(0, 1).
		Width(panelWidth - 2).
		Render(m.formatHeader(panelWidth - 2))

	rows := make([]string, 0, height)
	if len(m.results) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color(vercelMuted)).Render("No matches yet"))
	} else {
		start := min(m.listOffset, len(m.results)-1)
		end := min(start+height-1, len(m.results))
		for i := start; i < end; i++ {
			row := m.formatRow(m.results[i], panelWidth-2)
			if i == m.selected {
				row = lipgloss.NewStyle().
					Background(lipgloss.Color(vercelAccentSoft)).
					Foreground(lipgloss.Color(vercelText)).
					BorderLeft(true).
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color(vercelAccent)).
					Padding(0, 1).
					Width(panelWidth - 2).
					Render(row)
			} else {
				bg := vercelSurfaceAlt
				if (i-start)%2 == 1 {
					bg = vercelSurface
				}
				row = lipgloss.NewStyle().
					Background(lipgloss.Color(bg)).
					Foreground(lipgloss.Color(vercelText)).
					Padding(0, 1).
					Width(panelWidth - 2).
					Render(row)
			}
			rows = append(rows, row)
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, append([]string{head}, rows...)...)
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(vercelBorder)).
		Width(panelWidth).
		Height(max(6, m.height-10)).
		Render(body)
}

func (m Model) renderPreviewPanel() string {
	panelWidth := max(24, (m.width-6)-(m.width-6)*62/100)
	title := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelText)).
		Bold(true).
		Padding(0, 1).
		Width(panelWidth - 2).
		Render("Preview")

	content := m.preview
	if content == "" {
		content = "Select a result to preview context"
	}
	previewBody := lipgloss.NewStyle().
		Foreground(lipgloss.Color(vercelText)).
		Background(lipgloss.Color(vercelSurfaceAlt)).
		Padding(1, 1).
		Width(panelWidth - 2).
		Height(max(5, m.height-13)).
		Render(content)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(vercelBorder)).
		Width(panelWidth).
		Height(max(6, m.height-10)).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, previewBody))
}

func (m Model) formatHeader(width int) string {
	fileW := max(10, width*36/100)
	lineW := 6
	colW := 5
	snippetW := max(10, width-fileW-lineW-colW-3)
	return padRight("File", fileW) + " " + padRight("Line", lineW) + " " + padRight("Col", colW) + " " + padRight("Snippet", snippetW)
}

func (m Model) formatRow(r domain.SearchResult, width int) string {
	fileW := max(10, width*36/100)
	lineW := 6
	colW := 5
	snippetW := max(10, width-fileW-lineW-colW-3)

	file := truncateLeft(r.File, fileW)
	line := fmt.Sprintf("%d", r.Line)
	col := fmt.Sprintf("%d", r.Column)
	snippet := strings.TrimSpace(r.Content)
	if snippet == "" {
		snippet = " "
	}
	snippet = truncateRight(snippet, snippetW)

	return padRight(file, fileW) + " " + padRight(line, lineW) + " " + padRight(col, colW) + " " + padRight(snippet, snippetW)
}

func (m Model) resultsPanelHeight() int {
	h := m.height - 13
	if h < 6 {
		return 6
	}
	return h
}
