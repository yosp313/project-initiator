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

// animCache holds pre-computed styles for the title animation so they
// are not re-allocated on every frame render.
type animCache struct {
	dim    lipgloss.Style
	glow   [6]lipgloss.Style // gradient from spark → dim
	bg     lipgloss.Style
	flash  lipgloss.Style
	normal [9]lipgloss.Style // one per art line, pre-colored
}

func buildAnimCache(s styles) animCache {
	panelBg := s.panelBg

	// Color palette for the art — gradient from accent to purple.
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

	var normal [9]lipgloss.Style
	for i, c := range artColors {
		normal[i] = lipgloss.NewStyle().Foreground(c).Bold(true).Background(panelBg)
	}

	// 6-level glow gradient from bright spark (#bb9af7) → dim (#3b4261).
	// Intermediate hex values interpolated between the two endpoints.
	glowColors := [6]lipgloss.Color{
		"#bb9af7", // level 0 — spark center
		"#9d8ad4", // level 1
		"#7f7ab1", // level 2
		"#636a8e", // level 3
		"#4f5c78", // level 4
		"#3b4261", // level 5 — nearly dim
	}
	var glow [6]lipgloss.Style
	for i, c := range glowColors {
		glow[i] = lipgloss.NewStyle().Foreground(c).Background(panelBg)
		if i == 0 {
			glow[i] = glow[i].Bold(true)
		}
	}

	return animCache{
		dim:    lipgloss.NewStyle().Foreground(s.soft).Background(panelBg),
		glow:   glow,
		bg:     lipgloss.NewStyle().Background(panelBg),
		flash:  lipgloss.NewStyle().Foreground(lipgloss.Color("#c0caf5")).Bold(true).Background(panelBg),
		normal: normal,
	}
}

// renderAnimatedBorder returns a single styled border line with a traveling spark.
// width is the total character width. frame drives the spark position.
func renderAnimatedBorder(width int, frame int, cache animCache) string {
	if width < 2 {
		return ""
	}

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

		if dist < len(cache.glow) {
			b.WriteString(cache.glow[dist].Render(ch))
		} else {
			b.WriteString(cache.dim.Render(ch))
		}
	}
	return b.String()
}

// renderAnimatedTitle composes the full animated title block:
// border line, ASCII art with typing reveal, border line.
func (m model) renderAnimatedTitle(width int) string {
	art := asciiArt()
	aw := artWidth()
	frame := m.titleFrame
	revealedCols := frame * revealColumns
	if revealedCols > aw {
		revealedCols = aw
	}

	cache := m.animCache

	var lines []string

	// Top border
	lines = append(lines, renderAnimatedBorder(width, frame, cache))

	// Render each art line with typing reveal
	for lineIdx, artLine := range art {
		runes := []rune(artLine)
		artRuneLen := len(runes)

		// Pad to artWidth
		for len(runes) < aw {
			runes = append(runes, ' ')
		}

		// Pre-computed style for this line
		normalStyle := cache.normal[lineIdx]

		var lineBuilder strings.Builder

		// Center padding
		leftPad := (width - aw) / 2
		if leftPad < 0 {
			leftPad = 0
		}
		if leftPad > 0 {
			lineBuilder.WriteString(cache.bg.Render(strings.Repeat(" ", leftPad)))
		}

		if artLine == "" {
			// Blank separator line
			lineBuilder.WriteString(cache.bg.Render(strings.Repeat(" ", aw)))
		} else {
			for col := 0; col < artRuneLen; col++ {
				ch := string(runes[col])
				if col >= revealedCols {
					// Not yet revealed — render as space
					lineBuilder.WriteString(cache.bg.Render(" "))
				} else if col >= revealedCols-revealColumns && frame < revealTotalTicks() {
					// Flash edge — just revealed this tick
					lineBuilder.WriteString(cache.flash.Render(ch))
				} else {
					lineBuilder.WriteString(normalStyle.Render(ch))
				}
			}
			// Pad remaining after art chars
			remaining := aw - artRuneLen
			if remaining > 0 {
				lineBuilder.WriteString(cache.bg.Render(strings.Repeat(" ", remaining)))
			}
		}

		// Right padding
		rightPad := width - leftPad - aw
		if rightPad > 0 {
			lineBuilder.WriteString(cache.bg.Render(strings.Repeat(" ", rightPad)))
		}

		lines = append(lines, lineBuilder.String())
	}

	// Bottom border
	lines = append(lines, renderAnimatedBorder(width, frame+width/2, cache))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
