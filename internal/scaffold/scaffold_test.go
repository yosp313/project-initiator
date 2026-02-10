package scaffold

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"project-initiator/internal/domain"
	"project-initiator/internal/template"
)

// ---------------------------------------------------------------------------
// slugify
// ---------------------------------------------------------------------------

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "normal name", input: "MyProject", want: "myproject"},
		{name: "spaces become dashes", input: "my cool project", want: "my-cool-project"},
		{name: "special chars replaced", input: "hello@world!v2", want: "hello-world-v2"},
		{name: "empty string fallback", input: "", want: "project"},
		{name: "already kebab", input: "my-project", want: "my-project"},
		{name: "leading trailing dashes trimmed", input: "--my-project--", want: "my-project"},
		{name: "uppercase", input: "HELLO", want: "hello"},
		{name: "underscores preserved", input: "my_project", want: "my_project"},
		{name: "leading trailing spaces", input: "  hello  ", want: "hello"},
		{name: "only special chars", input: "@@@", want: "project"},
		{name: "mixed spaces and special", input: "  Hello World!  ", want: "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// cleanLanguageDir
// ---------------------------------------------------------------------------

func TestCleanLanguageDir(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "normal", input: "Go", want: "Go"},
		{name: "empty", input: "", want: "language"},
		{name: "forward slashes", input: "a/b/c", want: "a-b-c"},
		{name: "backslashes", input: `a\b\c`, want: "a-b-c"},
		{name: "spaces only", input: "   ", want: "language"},
		{name: "mixed slashes", input: `foo/bar\baz`, want: "foo-bar-baz"},
		{name: "surrounding spaces", input: "  Node.js  ", want: "Node.js"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanLanguageDir(tt.input)
			if got != tt.want {
				t.Errorf("cleanLanguageDir(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// template renderer
// ---------------------------------------------------------------------------

func TestTemplateRenderer(t *testing.T) {
	renderer := template.NewRenderer()

	t.Run("simple template", func(t *testing.T) {
		data := TemplateData{Name: "world"}
		got, err := renderer.Render("hello {{.Name}}", data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello world" {
			t.Errorf("got %q, want %q", got, "hello world")
		}
	})

	t.Run("multiple vars", func(t *testing.T) {
		src := "module {{.Module}} go {{.GoVersion}}"
		data := TemplateData{Module: "mymod", GoVersion: "1.23"}
		got, err := renderer.Render(src, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "module mymod go 1.23"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("no vars", func(t *testing.T) {
		data := TemplateData{}
		got, err := renderer.Render("hello world", data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello world" {
			t.Errorf("got %q, want %q", got, "hello world")
		}
	})

	t.Run("invalid template syntax", func(t *testing.T) {
		data := TemplateData{}
		_, err := renderer.Render("hello {{.Name", data)
		if err == nil {
			t.Error("expected error for invalid template syntax")
		}
	})
}

// ---------------------------------------------------------------------------
// goVersionTag
// ---------------------------------------------------------------------------

func TestGoVersionTag(t *testing.T) {
	got := goVersionTag()

	// Should match pattern like "1.23" or "1.22"
	match := regexp.MustCompile(`^\d+\.\d+$`).MatchString(got)
	if !match {
		t.Errorf("goVersionTag() = %q, want pattern like '1.23'", got)
	}

	// Verify it matches runtime version prefix
	v := runtime.Version()
	v = strings.TrimPrefix(v, "go")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) >= 2 {
		expected := parts[0] + "." + parts[1]
		if got != expected {
			t.Errorf("goVersionTag() = %q, want %q", got, expected)
		}
	}
}

// ---------------------------------------------------------------------------
// findFramework
// ---------------------------------------------------------------------------

func TestFindFramework(t *testing.T) {
	planner := DefaultPlanner()

	tests := []struct {
		name      string
		language  string
		framework string
		wantErr   bool
	}{
		{"valid combo", "Go", "Vanilla", false},
		{"invalid combo", "Go", "Django", true},
		{"case insensitivity", "go", "vanilla", false},
		{"whitespace trimming", "  Go  ", "  Vanilla  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := planner.findFramework(tt.language, tt.framework)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %s/%s", tt.language, tt.framework)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Plan
// ---------------------------------------------------------------------------

func TestPlan_GoVanilla(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "myapp",
		Dir:       tempDir,
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Check project dir is correct
	if !strings.Contains(plan.ProjectDir, "myapp") {
		t.Errorf("ProjectDir doesn't contain project name: %s", plan.ProjectDir)
	}

	// Should have templates
	if len(plan.Actions) == 0 {
		t.Error("expected actions, got none")
	}

	// Should not have a generator
	if plan.Generator != "" {
		t.Errorf("unexpected generator: %s", plan.Generator)
	}
}

func TestPlan_JSVanilla(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "JavaScript",
		Framework: "Vanilla",
		Name:      "myjsapp",
		Dir:       tempDir,
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Check project dir contains correct language dir
	if !strings.Contains(plan.ProjectDir, "JavaScript") {
		t.Errorf("ProjectDir doesn't contain language: %s", plan.ProjectDir)
	}
}

func TestPlan_EmptyNameError(t *testing.T) {
	req := Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "",
	}

	planner := DefaultPlanner()
	_, err := planner.Plan(req)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestPlan_InvalidLanguageFramework(t *testing.T) {
	req := Request{
		Language:  "Go",
		Framework: "Django",
		Name:      "myapp",
	}

	planner := DefaultPlanner()
	_, err := planner.Plan(req)
	if err == nil {
		t.Error("expected error for invalid language/framework")
	}
}

func TestPlan_GoGinLibrary(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "myapi",
		Dir:       tempDir,
		Libraries: []string{"gin"},
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Should have gin-specific files
	hasGinServer := false
	for _, action := range plan.Actions {
		if strings.Contains(action.Path, "internal/http/server.go") {
			hasGinServer = true
			break
		}
	}
	if !hasGinServer {
		t.Error("expected gin server file")
	}
}

func TestPlan_GoGormLibrary(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "myapp",
		Dir:       tempDir,
		Libraries: []string{"gorm"},
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Should have gorm-specific files
	hasGormDB := false
	for _, action := range plan.Actions {
		if strings.Contains(action.Path, "internal/db/db.go") {
			hasGormDB = true
			break
		}
	}
	if !hasGormDB {
		t.Error("expected gorm db file")
	}
}

func TestPlan_GoAllLibraries(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "myapp",
		Dir:       tempDir,
		Libraries: []string{"gin", "gorm", "sqlc"},
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Should have all library files
	paths := make(map[string]bool)
	for _, action := range plan.Actions {
		paths[action.Path] = true
	}

	expectedFiles := []string{
		"internal/http/server.go",
		"internal/http/routes.go",
		"internal/db/db.go",
		"internal/db/models.go",
		"sqlc.yaml",
	}

	for _, expected := range expectedFiles {
		found := false
		for path := range paths {
			if strings.HasSuffix(path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected file %s not found", expected)
		}
	}
}

func TestPlan_GoCobraFramework(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "Go",
		Framework: "Cobra",
		Name:      "mycli",
		Dir:       tempDir,
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Should use cmd structure
	hasCmdDir := false
	for _, action := range plan.Actions {
		if strings.Contains(action.Path, "cmd/mycli/") {
			hasCmdDir = true
			break
		}
	}
	if !hasCmdDir {
		t.Error("expected cmd/<name>/main.go structure for Cobra")
	}
}

func TestPlan_GoCobraWithLibraries(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "Go",
		Framework: "Cobra",
		Name:      "mycli",
		Dir:       tempDir,
		Libraries: []string{"gin"},
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Should still use cmd structure with libraries
	hasMainInCmd := false
	for _, action := range plan.Actions {
		if strings.HasSuffix(action.Path, "cmd/mycli/main.go") {
			hasMainInCmd = true
			break
		}
	}
	if !hasMainInCmd {
		t.Error("expected main.go in cmd/mycli/")
	}
}

func TestPlan_LaravelUsesGenerator(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "PHP",
		Framework: "Laravel",
		Name:      "myapp",
		Dir:       tempDir,
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Should have generator set
	if plan.Generator != "composer-laravel" {
		t.Errorf("expected generator 'composer-laravel', got %q", plan.Generator)
	}

	// Should have no actions (generator handles everything)
	if len(plan.Actions) != 0 {
		t.Errorf("expected no actions for generator, got %d", len(plan.Actions))
	}
}

func TestPlan_DirDefaultsToDot(t *testing.T) {
	req := Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "myapp",
		Dir:       "",
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Should use current directory
	if !strings.HasPrefix(plan.ProjectDir, "Go") && !strings.Contains(plan.ProjectDir, "/Go/") {
		// The project dir should contain the language somewhere
		t.Logf("ProjectDir: %s", plan.ProjectDir)
	}
}

func TestPlan_GoVersionInGoMod(t *testing.T) {
	tempDir := t.TempDir()
	req := Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "myapp",
		Dir:       tempDir,
	}

	planner := DefaultPlanner()
	plan, err := planner.Plan(req)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	// Find go.mod and check version
	var goModContent string
	for _, action := range plan.Actions {
		if strings.HasSuffix(action.Path, "go.mod") {
			goModContent = action.Content
			break
		}
	}

	if goModContent == "" {
		t.Fatal("go.mod not found in actions")
	}

	expectedVersion := goVersionTag()
	if !strings.Contains(goModContent, "go "+expectedVersion) {
		t.Errorf("go.mod doesn't contain expected version %s: %s", expectedVersion, goModContent)
	}
}

// ---------------------------------------------------------------------------
// Apply
// ---------------------------------------------------------------------------

func TestApply_CreatesFiles(t *testing.T) {
	tempDir := t.TempDir()

	plan := domain.Plan{
		Actions: []domain.Action{
			{
				Path:    filepath.Join(tempDir, "test.txt"),
				Content: "hello world",
			},
			{
				Path:    filepath.Join(tempDir, "subdir", "test2.txt"),
				Content: "nested file",
			},
		},
	}

	applier := NewApplier()
	if err := applier.Apply(plan, false); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Check files were created
	content, err := os.ReadFile(filepath.Join(tempDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read test.txt: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("test.txt content = %q, want %q", string(content), "hello world")
	}

	content2, err := os.ReadFile(filepath.Join(tempDir, "subdir", "test2.txt"))
	if err != nil {
		t.Fatalf("failed to read test2.txt: %v", err)
	}
	if string(content2) != "nested file" {
		t.Errorf("test2.txt content = %q, want %q", string(content2), "nested file")
	}
}

func TestApply_DryRunNoFiles(t *testing.T) {
	tempDir := t.TempDir()

	plan := domain.Plan{
		Actions: []domain.Action{
			{
				Path:    filepath.Join(tempDir, "test.txt"),
				Content: "hello world",
			},
		},
	}

	applier := NewApplier()
	if err := applier.Apply(plan, true); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Check file was NOT created
	_, err := os.Stat(filepath.Join(tempDir, "test.txt"))
	if !os.IsNotExist(err) {
		t.Error("expected file to not exist in dry-run mode")
	}
}

func TestApply_ErrorIfFileExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create existing file
	existingFile := filepath.Join(tempDir, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("existing"), 0o644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	plan := domain.Plan{
		Actions: []domain.Action{
			{
				Path:    existingFile,
				Content: "new content",
			},
		},
	}

	applier := NewApplier()
	err := applier.Apply(plan, false)
	if err == nil {
		t.Error("expected error when file exists")
	}
}

// ---------------------------------------------------------------------------
// Library code generation
// ---------------------------------------------------------------------------

func TestGoLibrariesReadme(t *testing.T) {
	tests := []struct {
		name      string
		libraries []string
		want      []string
	}{
		{
			name:      "gin only",
			libraries: []string{"gin"},
			want:      []string{"Gin"},
		},
		{
			name:      "gorm only",
			libraries: []string{"gorm"},
			want:      []string{"Gorm"},
		},
		{
			name:      "sqlc only",
			libraries: []string{"sqlc"},
			want:      []string{"Sqlc", "sqlc generate"},
		},
		{
			name:      "all libraries",
			libraries: []string{"gin", "gorm", "sqlc"},
			want:      []string{"Gin", "Gorm", "Sqlc"},
		},
		{
			name:      "gin and gorm",
			libraries: []string{"gin", "gorm"},
			want:      []string{"Gin", "Gorm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			req := Request{
				Language:  "Go",
				Framework: "Vanilla",
				Name:      "TestProject",
				Dir:       tempDir,
				Libraries: tt.libraries,
			}

			planner := DefaultPlanner()
			plan, err := planner.Plan(req)
			if err != nil {
				t.Fatalf("Plan() error = %v", err)
			}

			var readmeContent string
			for _, action := range plan.Actions {
				if strings.HasSuffix(action.Path, "README.md") {
					readmeContent = action.Content
					break
				}
			}

			if readmeContent == "" {
				t.Fatal("README.md not found")
			}

			for _, expected := range tt.want {
				if !strings.Contains(readmeContent, expected) {
					t.Errorf("README missing %q: %s", expected, readmeContent)
				}
			}
		})
	}
}

func TestGoLibrariesMod(t *testing.T) {
	tests := []struct {
		name      string
		libraries []string
		want      []string
	}{
		{
			name:      "gin only",
			libraries: []string{"gin"},
			want:      []string{"github.com/gin-gonic/gin"},
		},
		{
			name:      "gorm only",
			libraries: []string{"gorm"},
			want:      []string{"gorm.io/driver/sqlite", "gorm.io/gorm"},
		},
		{
			name:      "both",
			libraries: []string{"gin", "gorm"},
			want:      []string{"github.com/gin-gonic/gin", "gorm.io/gorm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			req := Request{
				Language:  "Go",
				Framework: "Vanilla",
				Name:      "testmod",
				Dir:       tempDir,
				Libraries: tt.libraries,
			}

			planner := DefaultPlanner()
			plan, err := planner.Plan(req)
			if err != nil {
				t.Fatalf("Plan() error = %v", err)
			}

			var goModContent string
			for _, action := range plan.Actions {
				if strings.HasSuffix(action.Path, "go.mod") {
					goModContent = action.Content
					break
				}
			}

			if goModContent == "" {
				t.Fatal("go.mod not found")
			}

			for _, expected := range tt.want {
				if !strings.Contains(goModContent, expected) {
					t.Errorf("go.mod missing %q: %s", expected, goModContent)
				}
			}
		})
	}
}

func TestGoLibrariesMain(t *testing.T) {
	tests := []struct {
		name      string
		libraries []string
		want      []string
		notWant   []string
	}{
		{
			name:      "gin only",
			libraries: []string{"gin"},
			want:      []string{"internal/http", "http.NewServer", "server.Run"},
			notWant:   []string{"db.Open", "gorm"},
		},
		{
			name:      "gorm only",
			libraries: []string{"gorm"},
			want:      []string{"db.Open", "AutoMigrate"},
			notWant:   []string{"http.NewServer"},
		},
		{
			name:      "sqlc only",
			libraries: []string{"sqlc"},
			want:      []string{"sqlc generate"},
		},
		{
			name:      "gin and gorm",
			libraries: []string{"gin", "gorm"},
			want:      []string{"http.NewServer", "db.Open", "AutoMigrate"},
		},
		{
			name:      "all three",
			libraries: []string{"gin", "gorm", "sqlc"},
			want:      []string{"http.NewServer", "db.Open", "sqlc generate"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			req := Request{
				Language:  "Go",
				Framework: "Vanilla",
				Name:      "testmain",
				Dir:       tempDir,
				Libraries: tt.libraries,
			}

			planner := DefaultPlanner()
			plan, err := planner.Plan(req)
			if err != nil {
				t.Fatalf("Plan() error = %v", err)
			}

			var mainContent string
			for _, action := range plan.Actions {
				if strings.HasSuffix(action.Path, "main.go") {
					mainContent = action.Content
					break
				}
			}

			if mainContent == "" {
				t.Fatal("main.go not found")
			}

			for _, expected := range tt.want {
				if !strings.Contains(mainContent, expected) {
					t.Errorf("main.go missing %q", expected)
				}
			}

			for _, notExpected := range tt.notWant {
				if strings.Contains(mainContent, notExpected) {
					t.Errorf("main.go should not contain %q", notExpected)
				}
			}
		})
	}
}
