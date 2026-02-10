package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// ASCII art title with typing reveal and animated border
// ---------------------------------------------------------------------------

// asciiArt returns the raw lines for "SCAFFOLD" and "WIZARD" in block letters.
// Each word is 4 lines tall. The two words are separated by a blank line.
// All lines within a word have equal width (padded with spaces).
func asciiArt() []string {
	return []string{
		"▄███▄ ▄███▄  ▄█▄  █▀▀▀▀ █▀▀▀▀ ▄███▄ █     ▄██▄",
		"▀▄    █     █▀ ▀█ █▀▀   █▀▀   █   █ █     █  █",
		" ▀██▄ █     █▀▀▀█ █     █     █   █ █     █  █",
		"▀███▀ ▀███▀ ▀   ▀ ▀     ▀     ▀███▀ ▀▀▀▀▀ ▀██▀",
		"",
		"        █   █ ▀█▀ ▀▀▀█  ▄█▄  █▀▀▄ ▄██▄        ",
		"        █ █ █  █    █▀ █▀ ▀█ █▀▀▄ █  █        ",
		"        █▄█▄█  █   █▀  █▀▀▀█ █  █ █  █        ",
		"        ▀   ▀ ▀▀▀ █▀▀▀ ▀   ▀ ▀  ▀ ▀██▀        ",
	}
}

// artWidth returns the rune-width of the widest line in the ASCII art.
func artWidth() int {
	w := 0
	for _, line := range asciiArt() {
		n := runeLen(line)
		if n > w {
			w = n
		}
	}
	return w
}

func runeLen(s string) int {
	return len([]rune(s))
}

// revealColumns is the number of columns revealed per animation tick.
const revealColumns = 3

// revealTotalTicks returns how many ticks to fully reveal the art.
func revealTotalTicks() int {
	w := artWidth()
	ticks := w / revealColumns
	if w%revealColumns != 0 {
		ticks++
	}
	return ticks
}

// renderAnimatedBorder returns a single styled border line with a traveling spark.
// width is the total character width. frame drives the spark position.
func renderAnimatedBorder(width int, frame int, s styles) string {
	if width < 2 {
		return ""
	}

	panelBg := panelBackground(s)

	dimStyle := lipgloss.NewStyle().Foreground(s.soft).Background(panelBg)
	sparkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#bb9af7")).Bold(true).Background(panelBg)
	brightStyle := lipgloss.NewStyle().Foreground(s.accent).Background(panelBg)

	innerWidth := width - 2 // corners take 2 chars
	if innerWidth < 0 {
		innerWidth = 0
	}
	sparkPos := frame % (innerWidth + 2) // position along the entire width

	var b strings.Builder
	for i := 0; i < width; i++ {
		ch := "═"
		if i == 0 {
			ch = "╾"
		} else if i == width-1 {
			ch = "╼"
		}

		dist := sparkPos - i
		if dist < 0 {
			dist = -dist
		}

		switch {
		case dist == 0:
			b.WriteString(sparkStyle.Render(ch))
		case dist <= 2:
			b.WriteString(brightStyle.Render(ch))
		default:
			b.WriteString(dimStyle.Render(ch))
		}
	}
	return b.String()
}

// renderAnimatedTitle composes the full animated title block:
// border line, ASCII art with typing reveal, border line.
func (m model) renderAnimatedTitle(width int) string {
	panelBg := panelBackground(m.styles)

	art := asciiArt()
	aw := artWidth()
	frame := m.titleFrame
	revealedCols := frame * revealColumns
	if revealedCols > aw {
		revealedCols = aw
	}

	// Color palette for the art — gradient from accent to purple
	artColors := []lipgloss.Color{
		"#7aa2f7", // accent blue
		"#7aa2f7",
		"#7dcfff", // cyan
		"#7dcfff",
		"#24283b", // blank line separator — invisible
		"#bb9af7", // purple
		"#bb9af7",
		"#9d7cd8", // deeper purple
		"#9d7cd8",
	}

	flashColor := lipgloss.Color("#c0caf5") // bright white for the reveal edge
	bgStyle := lipgloss.NewStyle().Background(panelBg)

	var lines []string

	// Top border
	lines = append(lines, renderAnimatedBorder(width, frame, m.styles))

	// Render each art line with typing reveal
	for lineIdx, artLine := range art {
		runes := []rune(artLine)
		artRuneLen := len(runes)

		// Pad to artWidth
		for len(runes) < aw {
			runes = append(runes, ' ')
		}

		// Determine color for this line
		lineColor := lipgloss.Color("#7aa2f7")
		if lineIdx < len(artColors) {
			lineColor = artColors[lineIdx]
		}

		normalStyle := lipgloss.NewStyle().Foreground(lineColor).Bold(true).Background(panelBg)
		flashStyle := lipgloss.NewStyle().Foreground(flashColor).Bold(true).Background(panelBg)

		var lineBuilder strings.Builder

		// Center padding
		leftPad := (width - aw) / 2
		if leftPad < 0 {
			leftPad = 0
		}
		if leftPad > 0 {
			lineBuilder.WriteString(bgStyle.Render(strings.Repeat(" ", leftPad)))
		}

		if artLine == "" {
			// Blank separator line
			lineBuilder.WriteString(bgStyle.Render(strings.Repeat(" ", aw)))
		} else {
			for col := 0; col < artRuneLen; col++ {
				ch := string(runes[col])
				if col >= revealedCols {
					// Not yet revealed — render as space
					lineBuilder.WriteString(bgStyle.Render(" "))
				} else if col >= revealedCols-revealColumns && frame < revealTotalTicks() {
					// Flash edge — just revealed this tick
					lineBuilder.WriteString(flashStyle.Render(ch))
				} else {
					lineBuilder.WriteString(normalStyle.Render(ch))
				}
			}
			// Pad remaining after art chars
			remaining := aw - artRuneLen
			if remaining > 0 {
				lineBuilder.WriteString(bgStyle.Render(strings.Repeat(" ", remaining)))
			}
		}

		// Right padding
		rightPad := width - leftPad - aw
		if rightPad > 0 {
			lineBuilder.WriteString(bgStyle.Render(strings.Repeat(" ", rightPad)))
		}

		lines = append(lines, lineBuilder.String())
	}

	// Bottom border
	lines = append(lines, renderAnimatedBorder(width, frame+width/2, m.styles))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
