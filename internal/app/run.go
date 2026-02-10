package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"project-initiator/internal/config"
	"project-initiator/internal/domain"
	"project-initiator/internal/flags"
	"project-initiator/internal/scaffold"
	"project-initiator/internal/ui"
)

func Run(args []string) int {
	opts, err := flags.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "config error:", err)
		return 2
	}

	request, err := buildRequest(opts, cfg)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 2
	}

	plan, err := scaffold.DefaultPlanner().Plan(request)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if opts.DryRun {
		printPlan(plan)
		return 0
	}

	if plan.Generator != "" {
		if err := runGenerator(plan.Generator, plan.ProjectDir); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return 1
		}
	} else if err := scaffold.NewApplier().Apply(plan, false); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	gitOk := gitInit(plan.ProjectDir)

	if err := config.Save(opts.ConfigPath, config.Config{
		DefaultLanguage:  request.Language,
		DefaultFramework: request.Framework,
		DefaultDir:       request.Dir,
	}); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "config save error:", err)
	}

	printSuccess(request, plan, gitOk)
	return 0
}

func buildRequest(opts flags.Options, cfg config.Config) (scaffold.Request, error) {
	language := firstNonEmpty(opts.Language, cfg.DefaultLanguage)
	framework := firstNonEmpty(opts.Framework, cfg.DefaultFramework)
	name := opts.Name
	dir := firstNonEmpty(opts.Dir, cfg.DefaultDir)

	if opts.NoTUI {
		if name == "" {
			return scaffold.Request{}, errors.New("name is required when --no-tui is set")
		}
		return scaffold.Request{
			Language:  language,
			Framework: framework,
			Name:      name,
			Dir:       dir,
			DryRun:    opts.DryRun,
		}, nil
	}

	if name == "" || opts.Language == "" || opts.Framework == "" {
		wizard := ui.NewWizard(language, framework)
		program := tea.NewProgram(wizard, tea.WithAltScreen())
		finalModel, err := program.Run()
		if err != nil {
			return scaffold.Request{}, err
		}

		result, err := ui.ResultFromModel(finalModel)
		if err != nil {
			return scaffold.Request{}, err
		}

		if name == "" {
			name = result.Name
		}
		if opts.Language == "" {
			language = result.Language
		}
		if opts.Framework == "" {
			framework = result.Framework
		}
		libs := result.Libraries
		return scaffold.Request{
			Language:  language,
			Framework: framework,
			Name:      name,
			Dir:       dir,
			DryRun:    opts.DryRun,
			Libraries: libs,
		}, nil
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return scaffold.Request{}, errors.New("project name is required")
	}

	return scaffold.Request{
		Language:  language,
		Framework: framework,
		Name:      name,
		Dir:       dir,
		DryRun:    opts.DryRun,
		Libraries: nil,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func printPlan(plan domain.Plan) {
	_, _ = fmt.Fprintln(os.Stdout, "Plan:")
	_, _ = fmt.Fprintln(os.Stdout, "Project:", plan.ProjectDir)
	if plan.Generator != "" {
		_, _ = fmt.Fprintln(os.Stdout, "Generator:", plan.Generator)
	}
	for _, action := range plan.Actions {
		_, _ = fmt.Fprintln(os.Stdout, "-", action.Path)
	}
}

func printSuccess(request scaffold.Request, plan domain.Plan, gitOk bool) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.Green)
	labelStyle := lipgloss.NewStyle().Foreground(ui.Muted)
	valueStyle := lipgloss.NewStyle().Foreground(ui.Text)
	cmdStyle := lipgloss.NewStyle().Foreground(ui.Accent)
	hintStyle := lipgloss.NewStyle().Foreground(ui.Muted).Italic(true)

	lines := []string{
		"",
		titleStyle.Render("  Project created successfully!"),
		"",
		labelStyle.Render("  Path        ") + valueStyle.Render(plan.ProjectDir),
		labelStyle.Render("  Language    ") + valueStyle.Render(request.Language),
		labelStyle.Render("  Framework   ") + valueStyle.Render(request.Framework),
	}

	if len(request.Libraries) > 0 {
		lines = append(lines, labelStyle.Render("  Libraries   ")+valueStyle.Render(strings.Join(request.Libraries, ", ")))
	}

	fileCount := len(plan.Actions)
	noun := "files"
	if fileCount == 1 {
		noun = "file"
	}
	lines = append(lines, labelStyle.Render("  Files       ")+valueStyle.Render(fmt.Sprintf("%d %s created", fileCount, noun)))

	if gitOk {
		lines = append(lines, labelStyle.Render("  Git         ")+valueStyle.Render("initialized"))
	}

	lines = append(lines, "")
	lines = append(lines, hintStyle.Render("  Next steps:"))
	lines = append(lines, cmdStyle.Render("    cd "+plan.ProjectDir))

	nextCmd := nextStepCommand(request.Language)
	if nextCmd != "" {
		lines = append(lines, cmdStyle.Render("    "+nextCmd))
	}

	lines = append(lines, "")

	_, _ = fmt.Fprintln(os.Stdout, strings.Join(lines, "\n"))
}

func nextStepCommand(language string) string {
	switch strings.ToLower(language) {
	case "go":
		return "go mod tidy"
	case "node.js":
		return "npm install"
	case "bun":
		return "bun install"
	case "python":
		return "pip install -r requirements.txt"
	default:
		return ""
	}
}

func gitInit(projectDir string) bool {
	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func runGenerator(generator string, projectDir string) error {
	switch generator {
	case "composer-laravel":
		return runCommand("composer", []string{"create-project", "laravel/laravel", projectDir})
	default:
		return fmt.Errorf("unknown generator: %s", generator)
	}
}

func runCommand(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
