package ui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// newCleanList creates a list.Model with all chrome (title, filter, help,
// status bar, pagination) disabled — the standard configuration used by
// every list in the wizard.
func newCleanList(items []list.Item, delegate list.ItemDelegate, w, h int) list.Model {
	l := list.New(items, delegate, w, h)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	return l
}

func buildFrameworkList(language string, options map[string][]string, defaultFramework string, s styles) list.Model {
	frameworks := options[language]
	if len(frameworks) == 0 {
		frameworks = []string{"Vanilla"}
	}
	frameworks = uniqueStrings(frameworks)
	sortStrings(frameworks)
	items := make([]list.Item, 0, len(frameworks))
	for _, framework := range frameworks {
		description := frameworkDescription(language, framework)
		items = append(items, listItem{label: framework, description: description})
	}

	model := newCleanList(items, listDelegate{styles: s}, 0, 0)

	if defaultFramework != "" {
		selectListItem(&model, defaultFramework)
	}

	return model
}

func buildLibraryItems(language string, framework string, options map[string][]string, selected map[string]bool) []list.Item {
	key := language + "::" + framework
	libraries := uniqueStrings(options[key])
	sortStrings(libraries)
	items := make([]list.Item, 0, len(libraries))
	for _, lib := range libraries {
		label := "[ ] " + lib
		if selected[lib] {
			label = "[x] " + lib
		}
		items = append(items, listItem{label: label, description: "optional package"})
	}
	return items
}

func buildLibrariesList(language string, framework string, options map[string][]string, selected map[string]bool, s styles) list.Model {
	items := buildLibraryItems(language, framework, options, selected)
	return newCleanList(items, listDelegate{styles: s}, 0, 0)
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func sortStrings(values []string) {
	slices.SortFunc(values, func(a, b string) int {
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	})
}

func frameworkDescription(language string, framework string) string {
	switch strings.ToLower(framework) {
	case "vanilla":
		return "minimal starter"
	case "cobra":
		return "CLI app structure"
	case "express":
		return "Node.js web server"
	case "hono":
		return "lightweight web framework"
	case "nestjs":
		return "typed Node framework"
	case "bun":
		return "Bun runtime server"
	case "fastapi":
		return "Python API server"
	case "laravel":
		return "PHP web framework"
	default:
		return fmt.Sprintf("%s template", language)
	}
}

func selectListItem(model *list.Model, label string) {
	for i, item := range model.Items() {
		if candidate, ok := item.(listItem); ok {
			if strings.EqualFold(candidate.label, label) {
				model.Select(i)
				return
			}
		}
	}
}

func selectedLibraries(selected map[string]bool) []string {
	values := make([]string, 0, len(selected))
	for name, isSelected := range selected {
		if isSelected {
			values = append(values, name)
		}
	}
	sortStrings(values)
	return values
}

func (m model) listHeightFixed() int {
	reservedRows := 14
	return clamp(m.panelH-reservedRows, 6, 30)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}

	return false
}

func clamp(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func stageTitle(s stage) string {
	switch s {
	case stageLanguage:
		return "Choose a language"
	case stageFramework:
		return "Choose a framework"
	case stageLibraries:
		return "Choose libraries"
	case stageName:
		return "Name your project"
	case stageConfirm:
		return "Confirm your selections"
	default:
		return ""
	}
}

func stageSubtitle(s stage) string {
	switch s {
	case stageLanguage:
		return "Pick the main language for the starter"
	case stageFramework:
		return "Select the starter template"
	case stageLibraries:
		return "Select optional packages (space to toggle)"
	case stageName:
		return "This will create the folder name"
	case stageConfirm:
		return "Review before creating the project"
	default:
		return ""
	}
}

func (m model) stageProgress() float64 {
	hasLibs := len(m.libraries.Items()) > 0
	totalSteps := 3
	if hasLibs {
		totalSteps = 4
	}
	switch m.stage {
	case stageLanguage:
		return 0.0
	case stageFramework:
		return 1.0 / float64(totalSteps)
	case stageLibraries:
		return 2.0 / float64(totalSteps)
	case stageName:
		if hasLibs {
			return 3.0 / float64(totalSteps)
		}
		return 2.0 / float64(totalSteps)
	case stageConfirm:
		return 1.0
	default:
		return 0.0
	}
}

func (m model) stepLabel() string {
	hasLibs := len(m.libraries.Items()) > 0
	switch m.stage {
	case stageLanguage:
		return "Step 1"
	case stageFramework:
		return "Step 2"
	case stageLibraries:
		return "Step 3/4"
	case stageName:
		if hasLibs {
			return "Step 4/4"
		}
		return "Step 3/3"
	case stageConfirm:
		return "Review"
	default:
		return ""
	}
}

// ---------------------------------------------------------------------------
// View rendering helpers
// ---------------------------------------------------------------------------

func (m model) renderFrame(content string, step string) string {
	if m.width == 0 {
		m.width = 96
	}
	if m.height == 0 {
		m.height = 36
	}
	if m.panelW == 0 {
		m.panelW = 88
	}
	if m.panelH == 0 {
		m.panelH = 32
	}

	// Panel entrance animation — scale dimensions during spring approach.
	pw := m.panelW
	ph := m.panelH
	if !m.panelReady {
		pw = int(float64(m.panelW) * m.panelScale)
		ph = int(float64(m.panelH) * m.panelScale)
		if pw < 1 {
			pw = 1
		}
		if ph < 1 {
			ph = 1
		}
	}

	contentWidth := pw - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	titleBlock := m.renderAnimatedTitle(contentWidth)

	// Status bar: step label + progress bar + help bindings.
	prog := m.progress.ViewAs(m.stageProgress())
	helpView := m.help.ShortHelpView(keys.ShortHelp())
	status := m.styles.status.Render(step + "  " + prog + "  •  " + helpView)

	stageTitleLine := m.styles.listTitle.Render(stageTitle(m.stage))
	stageSubtitleLine := m.styles.subheader.Render(stageSubtitle(m.stage))
	contentBlock := m.renderContentBlock(content, contentWidth)

	// Stage transition — apply horizontal offset by padding/clipping content.
	if m.transActive {
		offset := int(m.transOffset)
		if offset != 0 {
			stageTitleLine = applyHorizontalOffset(stageTitleLine, offset, contentWidth)
			stageSubtitleLine = applyHorizontalOffset(stageSubtitleLine, offset, contentWidth)
			contentBlock = applyHorizontalOffset(contentBlock, offset, contentWidth)
		}
	}

	rowBg := m.styles.panelBg
	rowStyle := lipgloss.NewStyle().Width(contentWidth).Background(rowBg)
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		rowStyle.Render(titleBlock),
		rowStyle.Render(stageTitleLine),
		rowStyle.Render(stageSubtitleLine),
		rowStyle.Render(contentBlock),
		rowStyle.Render(status),
	)
	innerHeight := ph - 4
	if innerHeight < 1 {
		innerHeight = 1
	}
	body = lipgloss.NewStyle().Width(contentWidth).Height(innerHeight).Background(rowBg).Render(body)
	panel := m.styles.panel.Width(pw).Height(ph).Render(body)
	return m.styles.frame.Width(m.width).Height(m.height).Align(lipgloss.Center, lipgloss.Center).Render(panel)
}

// applyHorizontalOffset shifts rendered text by the given column offset.
// Positive offset shifts right (content slides in from right); negative shifts left.
func applyHorizontalOffset(text string, offset int, maxWidth int) string {
	if offset == 0 || maxWidth <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		runes := []rune(line)
		if offset > 0 {
			// Shift right: prepend spaces, truncate to width.
			pad := strings.Repeat(" ", offset)
			shifted := pad + string(runes)
			sr := []rune(shifted)
			if len(sr) > maxWidth {
				sr = sr[:maxWidth]
			}
			result = append(result, string(sr))
		} else {
			// Shift left: drop leading chars, pad end.
			drop := -offset
			if drop >= len(runes) {
				result = append(result, strings.Repeat(" ", maxWidth))
			} else {
				shifted := string(runes[drop:])
				sr := []rune(shifted)
				if len(sr) < maxWidth {
					shifted += strings.Repeat(" ", maxWidth-len(sr))
				}
				result = append(result, shifted)
			}
		}
	}
	return strings.Join(result, "\n")
}

func (m model) renderContentBlock(content string, width int) string {
	rowBg := m.styles.panelBg
	height := m.listHeightFixed()
	return lipgloss.NewStyle().Width(width).Height(height).Background(rowBg).Render(content)
}

func (m model) renderNameInput() string {
	rowBg := m.styles.panelBg
	blankLine := lipgloss.NewStyle().Background(rowBg).Render(" ")
	label := m.styles.inputLabel.Render("Project name")
	box := m.styles.inputFocused.Render(m.name.View())
	help := m.styles.help.Render("Tip: Use a short, kebab-case name")

	if m.nameErr != "" {
		errStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#f52a65", Dark: "#f7768e"}).
			Background(rowBg)
		errLine := errStyle.Render("  " + m.nameErr)
		return lipgloss.JoinVertical(lipgloss.Left, label, blankLine, box, errLine, blankLine, help)
	}

	return lipgloss.JoinVertical(lipgloss.Left, label, blankLine, box, blankLine, help)
}

func (m model) renderConfirmation() string {
	rowBg := m.styles.panelBg
	blankLine := lipgloss.NewStyle().Background(rowBg).Render(" ")

	labelStyle := m.styles.inputLabel
	valueStyle := m.styles.listSelected

	lines := []string{
		labelStyle.Render("Language    ") + valueStyle.Render(m.result.Language),
		labelStyle.Render("Framework   ") + valueStyle.Render(m.result.Framework),
	}

	if len(m.result.Libraries) > 0 {
		lines = append(lines, labelStyle.Render("Libraries   ")+valueStyle.Render(strings.Join(m.result.Libraries, ", ")))
	}

	lines = append(lines, labelStyle.Render("Name        ")+valueStyle.Render(m.result.Name))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	hint := m.styles.help.Render("Press Enter to create project")
	return lipgloss.JoinVertical(lipgloss.Left, content, blankLine, hint)
}
