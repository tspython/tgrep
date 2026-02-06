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
	if m.width < 70 || m.height < 18 {
		return "Terminal too small. Resize to at least 70x18."
	}

	page := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelBg)).
		Foreground(lipgloss.Color(vercelText)).
		Padding(1, 1, 0, 1)

	innerWidth := max(30, m.width-page.GetHorizontalFrameSize())
	panelHeight := max(6, m.resultsPanelHeight())

	header := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelText)).
		Padding(0, 1).
		Bold(true).
		Width(innerWidth).
		Render("tgrep  -  fast local search")

	queryTitle := lipgloss.NewStyle().Foreground(lipgloss.Color(vercelMuted)).Bold(true).Render("Query")
	queryValue := m.query
	if queryValue == "" {
		queryValue = "Type a regex or plain text and press Enter"
	}
	queryBorder := vercelBorderStrong
	if m.queryFocus == focusQuery {
		queryBorder = vercelAccent
	}
	queryBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(queryBorder)).
		Background(lipgloss.Color(vercelSurfaceAlt)).
		Foreground(lipgloss.Color(vercelText)).
		Padding(0, 1)
	queryBox := queryBoxStyle.
		Width(max(10, innerWidth-queryBoxStyle.GetHorizontalFrameSize())).
		Render(queryValue + "_")

	filesTitle := lipgloss.NewStyle().Foreground(lipgloss.Color(vercelMuted)).Bold(true).Render("Files (glob)")
	filesValue := m.fileQuery
	if filesValue == "" {
		filesValue = "*"
	}
	filesBorder := vercelBorderStrong
	if m.queryFocus == focusFiles {
		filesBorder = vercelAccent
	}
	filesBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(filesBorder)).
		Background(lipgloss.Color(vercelSurfaceAlt)).
		Foreground(lipgloss.Color(vercelText)).
		Padding(0, 1)
	filesBox := filesBoxStyle.
		Width(max(10, innerWidth-filesBoxStyle.GetHorizontalFrameSize())).
		Render(filesValue + "_")

	status := fmt.Sprintf("cwd: .   files: %s   results: %d", filesValue, len(m.results))
	if m.searching {
		status = fmt.Sprintf("searching recursively in cwd (files: %s)...", filesValue)
	}
	statusBar := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelMuted)).
		Padding(0, 1).
		Width(innerWidth).
		Render(status)

	resultsPanelWidth := max(32, innerWidth*62/100)
	previewPanelWidth := max(24, innerWidth-resultsPanelWidth-1)
	resultsPanel := m.renderResultsPanel(resultsPanelWidth, panelHeight)
	previewPanel := m.renderPreviewPanel(previewPanelWidth, panelHeight)
	body := lipgloss.JoinHorizontal(lipgloss.Top, resultsPanel, " ", previewPanel)

	footerText := "Tab switch input | Enter search | Up/Down move | Esc quit"
	if m.err != nil {
		footerText = "error: " + m.err.Error()
	}
	footer := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelMuted)).
		Padding(0, 1).
		Width(innerWidth).
		Render(footerText)

	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		queryTitle,
		queryBox,
		filesTitle,
		filesBox,
		statusBar,
		"",
		body,
		footer,
	)

	return page.Render(layout)
}

func (m Model) renderResultsPanel(panelWidth, panelHeight int) string {
	rowWidth := max(10, panelWidth-2)
	rowCapacity := max(1, panelHeight-3)

	head := lipgloss.NewStyle().
		Background(lipgloss.Color(vercelSurface)).
		Foreground(lipgloss.Color(vercelText)).
		Bold(true).
		Padding(0, 1).
		Width(rowWidth).
		Render(m.formatHeader(rowWidth))

	rows := make([]string, 0, rowCapacity)
	if len(m.results) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color(vercelMuted)).Render("No matches yet"))
	} else {
		start := min(m.listOffset, len(m.results)-1)
		end := min(start+rowCapacity, len(m.results))
		for i := start; i < end; i++ {
			row := m.formatRow(m.results[i], rowWidth)
			if i == m.selected {
				row = lipgloss.NewStyle().
					Background(lipgloss.Color(vercelAccentSoft)).
					Foreground(lipgloss.Color(vercelText)).
					Padding(0, 1).
					Width(rowWidth).
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
					Width(rowWidth).
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
		Height(panelHeight).
		Render(body)
}

func (m Model) renderPreviewPanel(panelWidth, panelHeight int) string {
	bodyHeight := max(3, panelHeight-3)
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
		Height(bodyHeight).
		Render(content)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(vercelBorder)).
		Width(panelWidth).
		Height(panelHeight).
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
	h := m.height - 15
	if h < 6 {
		return 6
	}
	return h
}
