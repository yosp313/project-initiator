package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type listItem struct {
	label       string
	description string
}

func (i listItem) Title() string       { return i.label }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.label }

type styles struct {
	frame        lipgloss.Style
	panel        lipgloss.Style
	header       lipgloss.Style
	subheader    lipgloss.Style
	chip         lipgloss.Style
	chipGhost    lipgloss.Style
	listTitle    lipgloss.Style
	listSelected lipgloss.Style
	listNormal   lipgloss.Style
	listDesc     lipgloss.Style
	marker       lipgloss.Style
	inputLabel   lipgloss.Style
	inputBox     lipgloss.Style
	inputFocused lipgloss.Style
	help         lipgloss.Style
	status       lipgloss.Style
	accent       lipgloss.AdaptiveColor
	muted        lipgloss.AdaptiveColor
	soft         lipgloss.AdaptiveColor
	background   lipgloss.AdaptiveColor
	panelBg      lipgloss.AdaptiveColor
}

// Exported color constants used by the wizard UI and post-run output.
var (
	Accent = lipgloss.AdaptiveColor{Light: "#2e7de9", Dark: "#7aa2f7"}
	Muted  = lipgloss.AdaptiveColor{Light: "#8c8c8c", Dark: "#6b7280"}
	Text   = lipgloss.AdaptiveColor{Light: "#3760bf", Dark: "#c0caf5"}
	Green  = lipgloss.AdaptiveColor{Light: "#587539", Dark: "#9ece6a"}
)

func defaultStyles() styles {
	accent := Accent
	muted := Muted
	soft := lipgloss.AdaptiveColor{Light: "#c4c8da", Dark: "#3b4261"}
	background := lipgloss.AdaptiveColor{Light: "#d5d6db", Dark: "#1f2335"}
	panelBg := lipgloss.AdaptiveColor{Light: "#e1e2e7", Dark: "#24283b"}
	text := Text
	textSoft := lipgloss.AdaptiveColor{Light: "#6172b0", Dark: "#a9b1d6"}
	return styles{
		frame:        lipgloss.NewStyle().Background(background),
		panel:        lipgloss.NewStyle().Padding(1, 3).BorderStyle(lipgloss.RoundedBorder()).BorderForeground(soft).Background(panelBg),
		header:       lipgloss.NewStyle().Bold(true).Foreground(text).Background(panelBg),
		subheader:    lipgloss.NewStyle().Foreground(muted).Background(panelBg),
		chip:         lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#1a1b26"}).Background(accent).Padding(0, 1),
		chipGhost:    lipgloss.NewStyle().Foreground(textSoft).Background(soft).Padding(0, 1),
		listTitle:    lipgloss.NewStyle().Bold(true).Foreground(textSoft).Background(panelBg),
		listSelected: lipgloss.NewStyle().Foreground(text).Bold(true).Background(panelBg),
		listNormal:   lipgloss.NewStyle().Foreground(textSoft).Background(panelBg),
		listDesc:     lipgloss.NewStyle().Foreground(muted).Background(panelBg),
		marker:       lipgloss.NewStyle().Foreground(accent).Bold(true).Background(panelBg),
		inputLabel:   lipgloss.NewStyle().Foreground(muted).Background(panelBg),
		inputBox:     lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(soft).Padding(0, 1).Background(panelBg),
		inputFocused: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(accent).Padding(0, 1).Background(panelBg),
		help:         lipgloss.NewStyle().Foreground(muted).Background(panelBg),
		status:       lipgloss.NewStyle().Foreground(muted).Background(panelBg),
		accent:       accent,
		muted:        muted,
		soft:         soft,
		background:   background,
		panelBg:      panelBg,
	}
}

type listDelegate struct {
	styles styles
}

func (d listDelegate) Height() int  { return 2 }
func (d listDelegate) Spacing() int { return 0 }
func (d listDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d listDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(listItem)
	if !ok {
		return
	}

	rowBg := d.styles.panelBg

	isSelected := index == m.Index()
	nameStyle := d.styles.listNormal
	if isSelected {
		nameStyle = d.styles.listSelected
	}

	marker := d.styles.listNormal.Render("  ")
	if isSelected {
		marker = d.styles.marker.Render("› ")
	}
	nameLine := marker + nameStyle.Render(i.label)
	descLine := d.styles.listDesc.Render(i.description)
	rowStyle := lipgloss.NewStyle().Width(m.Width()).Background(rowBg)
	_, _ = fmt.Fprintln(w, rowStyle.Render(nameLine))
	if i.description != "" {
		indent := d.styles.listDesc.Render("  ")
		_, _ = fmt.Fprintln(w, rowStyle.Render(indent+descLine))
	}
}
