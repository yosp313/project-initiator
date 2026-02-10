package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"project-initiator/internal/config"
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

	plan, err := scaffold.Plan(request)
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
	} else if err := scaffold.Apply(plan.Actions, opts.DryRun); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := config.Save(opts.ConfigPath, config.Config{
		DefaultLanguage:  request.Language,
		DefaultFramework: request.Framework,
		DefaultDir:       request.Dir,
	}); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "config save error:", err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "Created", plan.ProjectDir)
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
		program := tea.NewProgram(wizard)
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

func printPlan(plan scaffold.PlanResult) {
	_, _ = fmt.Fprintln(os.Stdout, "Plan:")
	_, _ = fmt.Fprintln(os.Stdout, "Project:", plan.ProjectDir)
	if plan.Generator != "" {
		_, _ = fmt.Fprintln(os.Stdout, "Generator:", plan.Generator)
	}
	for _, action := range plan.Actions {
		_, _ = fmt.Fprintln(os.Stdout, "-", action.Path)
	}
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
