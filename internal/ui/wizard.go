package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"project-initiator/internal/scaffold"
)

type Result struct {
	Language  string
	Framework string
	Name      string
	Libraries []string
}

type listItem struct {
	label       string
	description string
}

func (i listItem) Title() string       { return i.label }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.label }

type stage int

const (
	stageLanguage stage = iota
	stageFramework
	stageLibraries
	stageName
	stageConfirm
	stageDone
)

type model struct {
	stage        stage
	languages    list.Model
	framework    list.Model
	libraries    list.Model
	name         textinput.Model
	result       Result
	options      map[string][]string
	libOptions   map[string][]string
	selectedLibs map[string]bool
	err          error
	width        int
	height       int
	panelW       int
	panelH       int
	styles       styles
	spinner      spinner.Model
	titleFrame   int
}

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
	accent       lipgloss.Color
	muted        lipgloss.Color
	soft         lipgloss.Color
	background   lipgloss.Color
}

func defaultStyles() styles {
	accent := lipgloss.Color("#7aa2f7")
	muted := lipgloss.Color("#6b7280")
	soft := lipgloss.Color("#3b4261")
	background := lipgloss.Color("#1f2335")
	panelBg := lipgloss.Color("#24283b")
	text := lipgloss.Color("#c0caf5")
	textSoft := lipgloss.Color("#a9b1d6")
	return styles{
		frame:        lipgloss.NewStyle().Background(background),
		panel:        lipgloss.NewStyle().Padding(1, 3).BorderStyle(lipgloss.RoundedBorder()).BorderForeground(soft).Background(panelBg),
		header:       lipgloss.NewStyle().Bold(true).Foreground(text).Background(panelBg),
		subheader:    lipgloss.NewStyle().Foreground(muted).Background(panelBg),
		chip:         lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1b26")).Background(accent).Padding(0, 1),
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

	rowBg, ok := d.styles.panel.GetBackground().(lipgloss.Color)
	if !ok {
		rowBg = lipgloss.Color("#24283b")
	}

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

func NewWizard(defaultLanguage string, defaultFramework string) tea.Model {
	styles := defaultStyles()
	spin := spinner.New()
	spin.Spinner = spinner.MiniDot
	spin.Style = lipgloss.NewStyle().Foreground(styles.accent).Background(lipgloss.Color("#24283b"))
	options := map[string][]string{}
	libOptions := map[string][]string{}
	for _, opt := range scaffold.Options {
		options[opt.Language] = append(options[opt.Language], opt.Framework)
		if len(opt.Libraries) > 0 {
			key := opt.Language + "::" + opt.Framework
			for _, lib := range opt.Libraries {
				libOptions[key] = append(libOptions[key], lib.Name)
			}
		}
	}
	for lang, frameworks := range options {
		if !contains(frameworks, "Vanilla") {
			options[lang] = append([]string{"Vanilla"}, frameworks...)
		}
	}
	if defaultFramework == "" {
		defaultFramework = "Vanilla"
	}

	langNames := make([]string, 0, len(options))
	for lang := range options {
		langNames = append(langNames, lang)
	}
	sortStrings(langNames)

	langItems := make([]list.Item, 0, len(langNames))
	for _, lang := range langNames {
		frameworks := options[lang]
		noun := "templates"
		if len(frameworks) == 1 {
			noun = "template"
		}
		description := fmt.Sprintf("%d %s", len(frameworks), noun)
		langItems = append(langItems, listItem{label: lang, description: description})
	}

	langList := list.New(langItems, listDelegate{styles: styles}, 0, 0)
	langList.Title = ""
	langList.SetShowTitle(false)
	langList.SetShowFilter(false)
	langList.SetFilteringEnabled(false)
	langList.SetShowHelp(false)
	langList.SetShowStatusBar(false)
	langList.SetShowPagination(false)

	if defaultLanguage != "" {
		selectListItem(&langList, defaultLanguage)
	}

	frameworkList := list.New([]list.Item{}, listDelegate{styles: styles}, 0, 0)
	frameworkList.Title = ""
	frameworkList.SetShowTitle(false)
	frameworkList.SetShowFilter(false)
	frameworkList.SetFilteringEnabled(false)
	frameworkList.SetShowHelp(false)
	frameworkList.SetShowStatusBar(false)
	frameworkList.SetShowPagination(false)

	libraryList := list.New([]list.Item{}, listDelegate{styles: styles}, 0, 0)
	libraryList.Title = ""
	libraryList.SetShowTitle(false)
	libraryList.SetShowFilter(false)
	libraryList.SetFilteringEnabled(false)
	libraryList.SetShowHelp(false)
	libraryList.SetShowStatusBar(false)
	libraryList.SetShowPagination(false)

	nameInput := textinput.New()
	nameInput.Placeholder = "my-project"
	nameInput.Prompt = ""
	nameInput.Focus()
	nameInput.CharLimit = 64

	return model{
		stage:        stageLanguage,
		languages:    langList,
		framework:    frameworkList,
		libraries:    libraryList,
		name:         nameInput,
		options:      options,
		libOptions:   libOptions,
		selectedLibs: map[string]bool{},
		result:       Result{Language: defaultLanguage, Framework: defaultFramework},
		styles:       styles,
		spinner:      spin,
	}
}

func ResultFromModel(m tea.Model) (Result, error) {
	modelValue, ok := m.(model)
	if !ok {
		return Result{}, fmt.Errorf("unexpected model")
	}

	if modelValue.err != nil {
		return Result{}, modelValue.err
	}

	return modelValue.result, nil
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.err = fmt.Errorf("cancelled")
			return m, tea.Quit
		case "b", "left", "backspace":
			if m.stage != stageName {
				m = m.back()
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.panelW = clamp(int(float64(m.width)*0.80), 64, m.width-4)
		m.panelH = clamp(int(float64(m.height)*0.80), 28, m.height-4)
		listWidth := clamp(m.panelW-8, 56, 100)
		listHeight := m.listHeightFixed()
		m.languages.SetSize(listWidth, listHeight)
		m.framework.SetSize(listWidth, listHeight)
		m.libraries.SetSize(listWidth, listHeight)
		m.name.Width = clamp(m.panelW-14, 24, 72)
	}

	var spinCmd tea.Cmd
	m.spinner, spinCmd = m.spinner.Update(msg)
	if _, ok := msg.(spinner.TickMsg); ok {
		m.titleFrame++
	}

	switch m.stage {
	case stageLanguage:
		modelValue, cmd := m.updateLanguage(msg)
		return modelValue, tea.Batch(cmd, spinCmd)
	case stageFramework:
		modelValue, cmd := m.updateFramework(msg)
		return modelValue, tea.Batch(cmd, spinCmd)
	case stageLibraries:
		modelValue, cmd := m.updateLibraries(msg)
		return modelValue, tea.Batch(cmd, spinCmd)
	case stageName:
		modelValue, cmd := m.updateName(msg)
		return modelValue, tea.Batch(cmd, spinCmd)
	case stageConfirm:
		modelValue, cmd := m.updateConfirm(msg)
		return modelValue, tea.Batch(cmd, spinCmd)
	case stageDone:
		return m, tea.Quit
	default:
		return m, spinCmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("error: %v\n", m.err)
	}

	switch m.stage {
	case stageLanguage:
		return m.renderFrame(m.languages.View(), m.stepLabel())
	case stageFramework:
		return m.renderFrame(m.framework.View(), m.stepLabel())
	case stageLibraries:
		return m.renderFrame(m.libraries.View(), m.stepLabel())
	case stageName:
		return m.renderFrame(m.renderNameInput(), m.stepLabel())
	case stageConfirm:
		return m.renderFrame(m.renderConfirmation(), m.stepLabel())
	case stageDone:
		return "done\n"
	default:
		return ""
	}
}

func (m model) updateLanguage(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.languages, cmd = m.languages.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" {
			item, ok := m.languages.SelectedItem().(listItem)
			if !ok {
				m.err = fmt.Errorf("no language selected")
				return m, tea.Quit
			}
			m.result.Language = item.label
			m.framework = buildFrameworkList(m.result.Language, m.options, m.result.Framework, m.styles)
			m.framework.SetSize(m.languages.Width(), m.listHeightFixed())
			m.stage = stageFramework
		}
	}

	return m, cmd
}

func (m model) updateFramework(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.framework, cmd = m.framework.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" {
			item, ok := m.framework.SelectedItem().(listItem)
			if !ok {
				m.err = fmt.Errorf("no framework selected")
				return m, tea.Quit
			}
			m.result.Framework = item.label
			m.selectedLibs = map[string]bool{}
			m.libraries = buildLibrariesList(m.result.Language, m.result.Framework, m.libOptions, m.selectedLibs, m.styles)
			m.libraries.SetSize(m.framework.Width(), m.listHeightFixed())
			if len(m.libraries.Items()) == 0 {
				m.stage = stageName
			} else {
				m.stage = stageLibraries
			}
		}
	}

	return m, cmd
}

func (m model) updateLibraries(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.libraries, cmd = m.libraries.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case " ":
			idx := m.libraries.Index()
			item, ok := m.libraries.SelectedItem().(listItem)
			if ok {
				name := strings.TrimPrefix(item.label, "[x] ")
				name = strings.TrimPrefix(name, "[ ] ")
				m.selectedLibs[name] = !m.selectedLibs[name]
				m.libraries = buildLibrariesList(m.result.Language, m.result.Framework, m.libOptions, m.selectedLibs, m.styles)
				m.libraries.SetSize(m.framework.Width(), m.listHeightFixed())
				if idx < len(m.libraries.Items()) {
					m.libraries.Select(idx)
				}
			}
		case "enter":
			m.stage = stageName
		}
	}

	return m, cmd
}

func (m model) updateName(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.name, cmd = m.name.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" {
			value := strings.TrimSpace(m.name.Value())
			if value == "" {
				m.err = fmt.Errorf("project name is required")
				return m, tea.Quit
			}
			m.result.Name = value
			m.result.Libraries = selectedLibraries(m.selectedLibs)
			m.stage = stageConfirm
		}
	}

	return m, cmd
}

func (m model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" {
			m.stage = stageDone
			return m, tea.Quit
		}
	}
	return m, nil
}

func buildFrameworkList(language string, options map[string][]string, defaultFramework string, styles styles) list.Model {
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

	model := list.New(items, listDelegate{styles: styles}, 0, 0)
	model.Title = ""
	model.SetShowTitle(false)
	model.SetShowFilter(false)
	model.SetFilteringEnabled(false)
	model.SetShowHelp(false)
	model.SetShowStatusBar(false)
	model.SetShowPagination(false)

	if defaultFramework != "" {
		selectListItem(&model, defaultFramework)
	}

	return model
}

func buildLibrariesList(language string, framework string, options map[string][]string, selected map[string]bool, styles styles) list.Model {
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

	model := list.New(items, listDelegate{styles: styles}, 0, 0)
	model.Title = ""
	model.SetShowTitle(false)
	model.SetShowFilter(false)
	model.SetFilteringEnabled(false)
	model.SetShowHelp(false)
	model.SetShowStatusBar(false)
	model.SetShowPagination(false)
	return model
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
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if strings.ToLower(values[j]) < strings.ToLower(values[i]) {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

func frameworkDescription(language string, framework string) string {
	key := strings.ToLower(framework)
	switch key {
	case "vanilla":
		return "minimal starter"
	case "cobra":
		return "CLI app structure"
	case "gin":
		return "HTTP API server"
	case "gorm":
		return "ORM database layer"
	case "sqlc":
		return "SQL code generation"
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
	contentWidth := m.panelW - 6
	titleBlock := m.renderAnimatedTitle(contentWidth)
	status := m.styles.status.Render(step + "  •  Enter to continue" + backHint(m.stage) + "  •  Esc to cancel")
	stageTitleLine := m.styles.listTitle.Render(stageTitle(m.stage))
	stageSubtitleLine := m.styles.subheader.Render(stageSubtitle(m.stage))
	contentBlock := m.renderContentBlock(content, contentWidth)

	rowBg, ok := m.styles.panel.GetBackground().(lipgloss.Color)
	if !ok {
		rowBg = lipgloss.Color("#24283b")
	}
	rowStyle := lipgloss.NewStyle().Width(contentWidth).Background(rowBg)
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		rowStyle.Render(titleBlock),
		rowStyle.Render(stageTitleLine),
		rowStyle.Render(stageSubtitleLine),
		rowStyle.Render(contentBlock),
		rowStyle.Render(status),
	)
	innerHeight := m.panelH - 4
	if innerHeight < 1 {
		innerHeight = 1
	}
	body = lipgloss.NewStyle().Width(contentWidth).Height(innerHeight).Background(rowBg).Render(body)
	panel := m.styles.panel.Width(m.panelW).Height(m.panelH).Render(body)
	return m.styles.frame.Width(m.width).Height(m.height).Align(lipgloss.Center, lipgloss.Center).Render(panel)
}

func (m model) renderContentBlock(content string, width int) string {
	rowBg, ok := m.styles.panel.GetBackground().(lipgloss.Color)
	if !ok {
		rowBg = lipgloss.Color("#24283b")
	}
	height := m.listHeightFixed()
	return lipgloss.NewStyle().Width(width).Height(height).Background(rowBg).Render(content)
}

func (m model) renderNameInput() string {
	rowBg, ok := m.styles.panel.GetBackground().(lipgloss.Color)
	if !ok {
		rowBg = lipgloss.Color("#24283b")
	}
	blankLine := lipgloss.NewStyle().Background(rowBg).Render(" ")
	label := m.styles.inputLabel.Render("Project name")
	box := m.styles.inputFocused.Render(m.name.View())
	help := m.styles.help.Render("Tip: Use a short, kebab-case name")
	return lipgloss.JoinVertical(lipgloss.Left, label, blankLine, box, blankLine, help)
}

func (m model) renderConfirmation() string {
	rowBg, ok := m.styles.panel.GetBackground().(lipgloss.Color)
	if !ok {
		rowBg = lipgloss.Color("#24283b")
	}
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

func backHint(s stage) string {
	if s == stageName || s == stageLanguage {
		return ""
	}
	return "  •  B to go back"
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

// borderChars used for the animated border line.
var borderSegments = []rune{'═', '═', '═', '═', '═', '═', '═', '═'}

// renderAnimatedBorder returns a single styled border line with a traveling spark.
// width is the total character width. frame drives the spark position.
func renderAnimatedBorder(width int, frame int, s styles) string {
	if width < 2 {
		return ""
	}

	panelBg, ok := s.panel.GetBackground().(lipgloss.Color)
	if !ok {
		panelBg = lipgloss.Color("#24283b")
	}

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
	panelBg, ok := m.styles.panel.GetBackground().(lipgloss.Color)
	if !ok {
		panelBg = lipgloss.Color("#24283b")
	}

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

func (m model) back() model {
	switch m.stage {
	case stageFramework:
		m.stage = stageLanguage
	case stageLibraries:
		m.stage = stageFramework
	case stageName:
		if len(m.libraries.Items()) > 0 {
			m.stage = stageLibraries
		} else {
			m.stage = stageFramework
		}
	case stageConfirm:
		m.stage = stageName
	}

	return m
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
