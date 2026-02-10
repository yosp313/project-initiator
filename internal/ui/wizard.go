package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"project-initiator/internal/scaffold"
)

// Result holds the user's selections from the wizard.
type Result struct {
	Language  string
	Framework string
	Name      string
	Libraries []string
}

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

// NewWizard creates the Bubble Tea model for the project wizard.
func NewWizard(defaultLanguage string, defaultFramework string) tea.Model {
	s := defaultStyles()
	spin := spinner.New()
	spin.Spinner = spinner.MiniDot
	spin.Style = lipgloss.NewStyle().Foreground(s.accent).Background(lipgloss.Color("#24283b"))
	options := map[string][]string{}
	libOptions := map[string][]string{}
	for _, opt := range scaffold.Frameworks {
		options[opt.Language] = append(options[opt.Language], opt.Name)
		if len(opt.Libraries) > 0 {
			key := opt.Language + "::" + opt.Name
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

	langList := newCleanList(langItems, listDelegate{styles: s}, 0, 0)

	if defaultLanguage != "" {
		selectListItem(&langList, defaultLanguage)
	}

	frameworkList := newCleanList([]list.Item{}, listDelegate{styles: s}, 0, 0)
	libraryList := newCleanList([]list.Item{}, listDelegate{styles: s}, 0, 0)

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
		styles:       s,
		spinner:      spin,
	}
}

// ResultFromModel extracts the wizard result from the final Bubble Tea model.
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

// ---------------------------------------------------------------------------
// Stage update handlers
// ---------------------------------------------------------------------------

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
