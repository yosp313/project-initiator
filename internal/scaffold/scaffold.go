// Package scaffold provides project scaffolding functionality.
package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"project-initiator/internal/domain"
	apperrors "project-initiator/internal/errors"
	"project-initiator/internal/library"
	"project-initiator/internal/template"
)

var nameSlug = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)

// Request represents a scaffolding request.
type Request struct {
	Language  string
	Framework string
	Name      string
	Dir       string
	DryRun    bool
	Libraries []string
}

// Planner handles project planning.
type Planner struct {
	renderer *template.Renderer
	options  []domain.Framework
}

// NewPlanner creates a new planner with the given options.
func NewPlanner(options []domain.Framework) *Planner {
	return &Planner{
		renderer: template.NewRenderer(),
		options:  options,
	}
}

// DefaultPlanner creates a planner with the default options.
func DefaultPlanner() *Planner {
	return NewPlanner(Frameworks)
}

// Plan creates a scaffolding plan for the given request.
func (p *Planner) Plan(req Request) (domain.Plan, error) {
	framework, err := p.findFramework(req.Language, req.Framework)
	if err != nil {
		return domain.Plan{}, err
	}

	project, err := p.buildProject(req, framework)
	if err != nil {
		return domain.Plan{}, err
	}

	return p.generatePlan(project, framework)
}

func (p *Planner) buildProject(req Request, framework domain.Framework) (domain.Project, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.Project{}, apperrors.NewValidationError("name", "project name is required")
	}

	dir := strings.TrimSpace(req.Dir)
	if dir == "" {
		dir = "."
	}

	slug := slugify(name)
	languageDir := cleanLanguageDir(framework.Language)
	projectDir := filepath.Join(filepath.Clean(dir), languageDir, slug)

	return domain.Project{
		Language:  framework.Language,
		Framework: framework.Name,
		Name:      name,
		Slug:      slug,
		Module:    slug,
		Dir:       projectDir,
		Libraries: req.Libraries,
	}, nil
}

func (p *Planner) generatePlan(project domain.Project, framework domain.Framework) (domain.Plan, error) {
	actions, err := p.generateActions(project, framework)
	if err != nil {
		return domain.Plan{}, apperrors.NewScaffoldError("generate actions", err)
	}

	return domain.Plan{
		ProjectDir: project.Dir,
		Actions:    actions,
		Generator:  framework.Generator,
	}, nil
}

func (p *Planner) generateActions(project domain.Project, framework domain.Framework) ([]domain.Action, error) {
	data := p.buildTemplateData(project)
	actions := make([]domain.Action, 0)

	// Generate base template actions
	for _, tmpl := range framework.Templates {
		content, err := p.renderer.Render(tmpl.Content, data)
		if err != nil {
			return nil, fmt.Errorf("render template content: %w", err)
		}

		relPath, err := p.renderer.Render(tmpl.RelativePath, data)
		if err != nil {
			return nil, fmt.Errorf("render template path: %w", err)
		}

		path := filepath.Join(project.Dir, filepath.FromSlash(relPath))
		actions = append(actions, domain.Action{Path: path, Content: content})
	}

	// Apply library-specific modifications for Go projects
	if strings.EqualFold(project.Language, "go") {
		actions = p.applyGoLibraries(actions, project)
	}

	return actions, nil
}

func (p *Planner) buildTemplateData(project domain.Project) TemplateData {
	selectedLibs := make(map[string]bool)
	for _, lib := range project.Libraries {
		selectedLibs[strings.ToLower(strings.TrimSpace(lib))] = true
	}

	return TemplateData{
		Name:        project.Name,
		PackageName: project.Slug,
		Module:      project.Module,
		Framework:   project.Framework,
		GoVersion:   goVersionTag(),
		UseGin:      selectedLibs["gin"],
		UseGorm:     selectedLibs["gorm"],
		UseSqlc:     selectedLibs["sqlc"],
	}
}

func (p *Planner) applyGoLibraries(actions []domain.Action, project domain.Project) []domain.Action {
	libMgr := library.NewManager(project)

	// Check if any libraries are enabled
	if !libMgr.HasLibrary("gin") && !libMgr.HasLibrary("gorm") && !libMgr.HasLibrary("sqlc") {
		return actions
	}

	// Remove replaced files
	replaced := libMgr.ReplacedFiles(project.Slug)
	filtered := make([]domain.Action, 0, len(actions))
	for _, action := range actions {
		relPath, err := filepath.Rel(project.Dir, action.Path)
		if err != nil {
			relPath = filepath.Base(action.Path)
		}
		relPath = filepath.ToSlash(relPath)
		if !replaced[relPath] {
			filtered = append(filtered, action)
		}
	}
	actions = filtered

	goVersion := goVersionTag()

	// Add library-specific files
	if libMgr.HasLibrary("gin") || libMgr.HasLibrary("gorm") || libMgr.HasLibrary("sqlc") {
		// Determine main file path based on framework
		mainPath := filepath.Join(project.Dir, "main.go")
		if strings.EqualFold(project.Framework, "cobra") {
			mainPath = filepath.Join(project.Dir, "cmd", project.Slug, "main.go")
		}

		actions = append(actions, domain.Action{
			Path:    mainPath,
			Content: libMgr.GenerateMain(project.Framework),
		})
		actions = append(actions, domain.Action{
			Path:    filepath.Join(project.Dir, "go.mod"),
			Content: libMgr.GenerateGoMod(goVersion),
		})
		actions = append(actions, domain.Action{
			Path:    filepath.Join(project.Dir, "README.md"),
			Content: libMgr.GenerateReadme(),
		})
	}

	// Add library-specific file templates
	for path, content := range libMgr.FileTemplates() {
		fullPath := filepath.Join(project.Dir, filepath.FromSlash(path))
		actions = append(actions, domain.Action{Path: fullPath, Content: content})
	}

	return actions
}

func (p *Planner) findFramework(lang, framework string) (domain.Framework, error) {
	lang = strings.TrimSpace(lang)
	framework = strings.TrimSpace(framework)

	for _, opt := range p.options {
		if strings.EqualFold(opt.Language, lang) && strings.EqualFold(opt.Name, framework) {
			return opt, nil
		}
	}

	return domain.Framework{}, fmt.Errorf("no template for %s / %s", lang, framework)
}

// TemplateData holds data for template rendering.
type TemplateData struct {
	Name        string
	PackageName string
	Module      string
	Framework   string
	GoVersion   string
	UseGin      bool
	UseGorm     bool
	UseSqlc     bool
}

// Applier handles applying scaffold plans.
type Applier struct{}

// NewApplier creates a new applier.
func NewApplier() *Applier {
	return &Applier{}
}

// Apply executes the plan by writing files to disk.
func (a *Applier) Apply(plan domain.Plan, dryRun bool) error {
	// Check for existing files first
	for _, action := range plan.Actions {
		if _, err := os.Stat(action.Path); err == nil {
			return fmt.Errorf("%w: %s", apperrors.ErrProjectExists, action.Path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("check file existence: %w", err)
		}
	}

	// Apply actions
	for _, action := range plan.Actions {
		if dryRun {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(action.Path), 0o755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}

		if err := os.WriteFile(action.Path, []byte(action.Content), 0o644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}

	return nil
}

func slugify(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, " ", "-")
	value = nameSlug.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-_")
	if value == "" {
		return "project"
	}
	return value
}

func cleanLanguageDir(language string) string {
	value := strings.TrimSpace(language)
	if value == "" {
		return "language"
	}

	replacer := func(r rune) rune {
		switch r {
		case '/', '\\':
			return '-'
		default:
			return r
		}
	}

	value = strings.Map(replacer, value)
	value = strings.TrimSpace(value)
	if value == "" {
		return "language"
	}
	return value
}

func goVersionTag() string {
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return "1.22"
}
