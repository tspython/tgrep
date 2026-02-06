package app

import "strings"

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncateRight(s string, width int) string {
	if width <= 1 || len(s) <= width {
		return s
	}
	return s[:width-1] + "..."
}

func truncateLeft(s string, width int) string {
	if width <= 1 || len(s) <= width {
		return s
	}
	return "..." + s[len(s)-width+3:]
}
