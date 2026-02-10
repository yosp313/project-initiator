package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
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

type keyMap struct {
	Quit  key.Binding
	Back  key.Binding
	Enter key.Binding
	Space key.Binding
}

// ShortHelp returns bindings for the compact help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Space, k.Back, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

var keys = keyMap{
	Quit:  key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("esc", "cancel")),
	Back:  key.NewBinding(key.WithKeys("b", "left", "backspace"), key.WithHelp("b", "back")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "continue")),
	Space: key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
}

type model struct {
	stage         stage
	languages     list.Model
	framework     list.Model
	libraries     list.Model
	name          textinput.Model
	help          help.Model
	progress      progress.Model
	result        Result
	options       map[string][]string
	libOptions    map[string][]string
	selectedLibs  map[string]bool
	err           error
	width         int
	height        int
	panelW        int
	panelH        int
	styles        styles
	animCache     animCache
	titleFrame    int
	animationDone bool
	nameErr       string

	// Spring-animated panel entrance.
	panelSpring harmonica.Spring
	panelScale  float64
	panelVel    float64
	panelReady  bool // true once panel has reached full size

	// Spring-animated stage transitions.
	transSpring harmonica.Spring
	transOffset float64 // horizontal offset in columns
	transVel    float64
	transActive bool
}

// NewWizard creates the Bubble Tea model for the project wizard.
func NewWizard(defaultLanguage string, defaultFramework string) tea.Model {
	s := defaultStyles()
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

	// Help model styled to match the status bar.
	h := help.New()
	h.ShortSeparator = "  •  "
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(s.accent).Background(s.panelBg)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(s.muted).Background(s.panelBg)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(s.muted).Background(s.panelBg)

	// Progress bar for step indicator.
	p := progress.New(
		progress.WithGradient(string(Accent.Dark), string(Muted.Dark)),
		progress.WithWidth(20),
		progress.WithoutPercentage(),
	)

	// Spring for panel entrance — slightly under-damped for a subtle bounce.
	panelSpring := harmonica.NewSpring(harmonica.FPS(60), 5.0, 0.7)
	// Spring for stage transitions — fast, minimal overshoot.
	transSpring := harmonica.NewSpring(harmonica.FPS(60), 8.0, 0.85)

	return model{
		stage:        stageLanguage,
		languages:    langList,
		framework:    frameworkList,
		libraries:    libraryList,
		name:         nameInput,
		help:         h,
		progress:     p,
		options:      options,
		libOptions:   libOptions,
		selectedLibs: map[string]bool{},
		result:       Result{Language: defaultLanguage, Framework: defaultFramework},
		styles:       s,
		animCache:    buildAnimCache(s),
		panelSpring:  panelSpring,
		panelScale:   0.0,
		transSpring:  transSpring,
	}
}

// ResultFromModel extracts the wizard result from the final Bubble Tea model.
func ResultFromModel(m tea.Model) (Result, error) {
	modelValue, ok := m.(model)
	if !ok {
		return Result{}, errors.New("unexpected model")
	}

	if modelValue.err != nil {
		return Result{}, modelValue.err
	}

	return modelValue.result, nil
}

// animationTickMsg drives the title animation at a fixed interval.
type animationTickMsg time.Time

// smoothTickMsg drives spring animations at 60fps.
type smoothTickMsg time.Time

func tickAnimation() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return animationTickMsg(t)
	})
}

func tickSmooth() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return smoothTickMsg(t)
	})
}

// updateBindings enables or disables key bindings based on the current stage.
func (m *model) updateBindings() {
	keys.Back.SetEnabled(m.stage != stageLanguage && m.stage != stageName)
	keys.Space.SetEnabled(m.stage == stageLibraries)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickAnimation(), tickSmooth(), m.name.Cursor.SetMode(cursor.CursorBlink))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.err = errors.New("cancelled")
			return m, tea.Quit
		case key.Matches(msg, keys.Back) && m.stage != stageName:
			prevStage := m.stage
			m = m.back()
			if m.stage != prevStage {
				m.triggerTransition(false)
			}
			m.updateBindings()
			return m, tickSmooth()
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
		m.help.Width = m.panelW - 6
	}

	var animCmd tea.Cmd
	if _, ok := msg.(animationTickMsg); ok {
		m.titleFrame++
		if !m.animationDone {
			contentWidth := 82 // default panelW(88) - 6
			if m.panelW > 0 {
				contentWidth = m.panelW - 6
			}
			sparkCycle := contentWidth + 2
			totalTicks := revealTotalTicks() + sparkCycle*2
			if m.titleFrame >= totalTicks {
				m.animationDone = true
			} else {
				animCmd = tickAnimation()
			}
		}
	}

	// Smooth tick: update spring animations (panel entrance + transitions).
	var smoothCmd tea.Cmd
	if _, ok := msg.(smoothTickMsg); ok {
		needsMore := false

		// Panel entrance spring.
		if !m.panelReady {
			m.panelScale, m.panelVel = m.panelSpring.Update(m.panelScale, m.panelVel, 1.0)
			if m.panelScale > 0.999 && absF(m.panelVel) < 0.001 {
				m.panelScale = 1.0
				m.panelVel = 0
				m.panelReady = true
			} else {
				needsMore = true
			}
		}

		// Stage transition spring.
		if m.transActive {
			m.transOffset, m.transVel = m.transSpring.Update(m.transOffset, m.transVel, 0.0)
			if absF(m.transOffset) < 0.5 && absF(m.transVel) < 0.5 {
				m.transOffset = 0
				m.transVel = 0
				m.transActive = false
			} else {
				needsMore = true
			}
		}

		if needsMore {
			smoothCmd = tickSmooth()
		}
	}

	switch m.stage {
	case stageLanguage:
		modelValue, cmd := m.updateLanguage(msg)
		return modelValue, tea.Batch(cmd, animCmd, smoothCmd)
	case stageFramework:
		modelValue, cmd := m.updateFramework(msg)
		return modelValue, tea.Batch(cmd, animCmd, smoothCmd)
	case stageLibraries:
		modelValue, cmd := m.updateLibraries(msg)
		return modelValue, tea.Batch(cmd, animCmd, smoothCmd)
	case stageName:
		modelValue, cmd := m.updateName(msg)
		return modelValue, tea.Batch(cmd, animCmd, smoothCmd)
	case stageConfirm:
		modelValue, cmd := m.updateConfirm(msg)
		return modelValue, tea.Batch(cmd, animCmd, smoothCmd)
	case stageDone:
		return m, tea.Quit
	default:
		return m, tea.Batch(animCmd, smoothCmd)
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
		if key.Matches(keyMsg, keys.Enter) {
			item, ok := m.languages.SelectedItem().(listItem)
			if !ok {
				m.err = errors.New("no language selected")
				return m, tea.Quit
			}
			m.result.Language = item.label
			m.framework = buildFrameworkList(m.result.Language, m.options, m.result.Framework, m.styles)
			m.framework.SetSize(m.languages.Width(), m.listHeightFixed())
			m.stage = stageFramework
			m.triggerTransition(true)
			m.updateBindings()
			return m, tea.Batch(cmd, tickSmooth())
		}
	}

	return m, cmd
}

func (m model) updateFramework(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.framework, cmd = m.framework.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, keys.Enter) {
			item, ok := m.framework.SelectedItem().(listItem)
			if !ok {
				m.err = errors.New("no framework selected")
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
			m.triggerTransition(true)
			m.updateBindings()
			return m, tea.Batch(cmd, tickSmooth())
		}
	}

	return m, cmd
}

func (m model) updateLibraries(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.libraries, cmd = m.libraries.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, keys.Space):
			idx := m.libraries.Index()
			item, ok := m.libraries.SelectedItem().(listItem)
			if ok {
				name := strings.TrimPrefix(item.label, "[x] ")
				name = strings.TrimPrefix(name, "[ ] ")
				m.selectedLibs[name] = !m.selectedLibs[name]
				m.libraries.SetItems(buildLibraryItems(m.result.Language, m.result.Framework, m.libOptions, m.selectedLibs))
				if idx < len(m.libraries.Items()) {
					m.libraries.Select(idx)
				}
			}
		case key.Matches(keyMsg, keys.Enter):
			m.stage = stageName
			m.triggerTransition(true)
			m.updateBindings()
			return m, tea.Batch(cmd, tickSmooth())
		}
	}

	return m, cmd
}

func (m model) updateName(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.name, cmd = m.name.Update(msg)

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, keys.Enter) {
			value := strings.TrimSpace(m.name.Value())
			if value == "" {
				m.nameErr = "Name is required"
				return m, cmd
			}
			m.nameErr = ""
			m.result.Name = value
			m.result.Libraries = selectedLibraries(m.selectedLibs)
			m.stage = stageConfirm
			m.triggerTransition(true)
			m.updateBindings()
			return m, tea.Batch(cmd, tickSmooth())
		}
	}

	return m, cmd
}

func (m model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, keys.Enter) {
			m.stage = stageDone
			return m, tea.Quit
		}
	}
	return m, nil
}

// triggerTransition sets up a horizontal slide animation.
// forward=true slides content in from the right; false from the left.
func (m *model) triggerTransition(forward bool) {
	contentWidth := 82 // default panelW(88) - 6
	if m.panelW > 0 {
		contentWidth = m.panelW - 6
	}
	if forward {
		m.transOffset = float64(contentWidth)
	} else {
		m.transOffset = float64(-contentWidth)
	}
	m.transVel = 0
	m.transActive = true
}

func absF(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
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
