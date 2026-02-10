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

	langItems := make([]list.Item, 0, len(options))
	for lang, frameworks := range options {
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
		m.panelH = clamp(int(float64(m.height)*0.70), 18, m.height-4)
		listWidth := clamp(m.panelW-8, 56, 100)
		listHeight := m.listHeightFixed()
		m.languages.SetSize(listWidth, listHeight)
		m.framework.SetSize(listWidth, listHeight)
		m.libraries.SetSize(listWidth, listHeight)
		m.name.Width = clamp(m.panelW-14, 24, 72)
	}

	var spinCmd tea.Cmd
	m.spinner, spinCmd = m.spinner.Update(msg)

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
		return m.renderFrame(m.languages.View(), "Step 1/3")
	case stageFramework:
		return m.renderFrame(m.framework.View(), "Step 2/4")
	case stageLibraries:
		return m.renderFrame(m.libraries.View(), "Step 3/4")
	case stageName:
		return m.renderFrame(m.renderNameInput(), "Step 4/4")
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
			m.stage = stageDone
			return m, tea.Quit
		}
	}

	return m, cmd
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
		m.height = 28
	}
	if m.panelW == 0 {
		m.panelW = 88
	}
	if m.panelH == 0 {
		m.panelH = 26
	}
	contentWidth := m.panelW - 6
	title := m.spinner.View() + " ▸ Scaffold Wizard ◂ " + m.spinner.View()
	header := m.styles.header.Copy().Width(contentWidth).Align(lipgloss.Center).Render(title)
	meta := m.styles.subheader.Copy().Width(contentWidth).Align(lipgloss.Center).Render("Choose a template and name")
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
		rowStyle.Render(header),
		rowStyle.Render(meta),
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

func backChip(s stage, styles styles) string {
	if s == stageName || s == stageLanguage {
		return ""
	}
	return " " + styles.chipGhost.Render("B Back")
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
	reservedRows := 5
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

func (m model) back() model {
	switch m.stage {
	case stageFramework:
		m.stage = stageLanguage
	case stageName:
		m.stage = stageFramework
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
