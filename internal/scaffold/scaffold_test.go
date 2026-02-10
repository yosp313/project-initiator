package scaffold

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
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
// render
// ---------------------------------------------------------------------------

func TestRender(t *testing.T) {
	t.Run("simple template", func(t *testing.T) {
		got, err := render("hello {{.Name}}", Data{Name: "world"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello world" {
			t.Errorf("got %q, want %q", got, "hello world")
		}
	})

	t.Run("multiple vars", func(t *testing.T) {
		src := "module {{.Module}} go {{.GoVersion}}"
		got, err := render(src, Data{Module: "mymod", GoVersion: "1.23"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "module mymod go 1.23"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("no vars", func(t *testing.T) {
		got, err := render("static content", Data{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "static content" {
			t.Errorf("got %q, want %q", got, "static content")
		}
	})

	t.Run("invalid template syntax", func(t *testing.T) {
		_, err := render("{{.Bad", Data{})
		if err == nil {
			t.Fatal("expected error for invalid template syntax, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// goVersionTag
// ---------------------------------------------------------------------------

func TestGoVersionTag(t *testing.T) {
	got := goVersionTag()

	// Must match a "major.minor" pattern like "1.23".
	matched, err := regexp.MatchString(`^\d+\.\d+$`, got)
	if err != nil {
		t.Fatalf("regexp error: %v", err)
	}
	if !matched {
		t.Errorf("goVersionTag() = %q, want X.Y semver format", got)
	}

	// Should be consistent with runtime.Version().
	rv := runtime.Version()
	if strings.HasPrefix(rv, "go") {
		parts := strings.SplitN(strings.TrimPrefix(rv, "go"), ".", 3)
		if len(parts) >= 2 {
			want := parts[0] + "." + parts[1]
			if got != want {
				t.Errorf("goVersionTag() = %q, want %q (from runtime %q)", got, want, rv)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// findOption
// ---------------------------------------------------------------------------

func TestFindOption(t *testing.T) {
	t.Run("valid combo", func(t *testing.T) {
		opt, err := findOption("Go", "Vanilla")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.EqualFold(opt.Language, "Go") || !strings.EqualFold(opt.Framework, "Vanilla") {
			t.Errorf("got %s/%s, want Go/Vanilla", opt.Language, opt.Framework)
		}
	})

	t.Run("invalid combo", func(t *testing.T) {
		_, err := findOption("Rust", "Tokio")
		if err == nil {
			t.Fatal("expected error for invalid combo, got nil")
		}
		if !strings.Contains(err.Error(), "no template") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("case insensitivity", func(t *testing.T) {
		opt, err := findOption("go", "vanilla")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opt.Language != "Go" {
			t.Errorf("got Language=%q, want Go", opt.Language)
		}
	})

	t.Run("whitespace trimming", func(t *testing.T) {
		opt, err := findOption("  Go  ", "  Vanilla  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opt.Language != "Go" {
			t.Errorf("got Language=%q, want Go", opt.Language)
		}
	})
}

// ---------------------------------------------------------------------------
// Plan
// ---------------------------------------------------------------------------

func TestPlan_GoVanilla(t *testing.T) {
	res, err := Plan(Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "My App",
		Dir:       "/tmp/test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantDir := filepath.Join("/tmp/test", "Go", "my-app")
	if res.ProjectDir != wantDir {
		t.Errorf("ProjectDir = %q, want %q", res.ProjectDir, wantDir)
	}

	paths := actionPaths(res.Actions)
	expectedFiles := []string{
		"main.go",
		"go.mod",
		"README.md",
		filepath.Join("internal", "app", "app.go"),
	}
	for _, f := range expectedFiles {
		full := filepath.Join(wantDir, f)
		if !paths[full] {
			t.Errorf("expected file %q in actions, not found", f)
		}
	}
}

func TestPlan_JSVanilla(t *testing.T) {
	res, err := Plan(Request{
		Language:  "JavaScript",
		Framework: "Vanilla",
		Name:      "jsapp",
		Dir:       "/tmp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := actionPaths(res.Actions)
	wantDir := filepath.Join("/tmp", "JavaScript", "jsapp")
	if res.ProjectDir != wantDir {
		t.Errorf("ProjectDir = %q, want %q", res.ProjectDir, wantDir)
	}

	for _, f := range []string{"package.json", filepath.Join("src", "index.js"), "README.md"} {
		full := filepath.Join(wantDir, f)
		if !paths[full] {
			t.Errorf("expected file %q in actions", f)
		}
	}
}

func TestPlan_EmptyNameError(t *testing.T) {
	_, err := Plan(Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "",
	})
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
	if !strings.Contains(err.Error(), "project name is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlan_InvalidLanguageFramework(t *testing.T) {
	_, err := Plan(Request{
		Language:  "Rust",
		Framework: "Tokio",
		Name:      "app",
	})
	if err == nil {
		t.Fatal("expected error for invalid language/framework, got nil")
	}
	if !strings.Contains(err.Error(), "no template") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlan_GoGinLibrary(t *testing.T) {
	res, err := Plan(Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "ginapp",
		Dir:       "/tmp",
		Libraries: []string{"gin"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := actionPaths(res.Actions)
	wantDir := res.ProjectDir

	// Gin-specific files should be present.
	for _, f := range []string{
		filepath.Join("internal", "http", "server.go"),
		filepath.Join("internal", "http", "routes.go"),
	} {
		full := filepath.Join(wantDir, f)
		if !paths[full] {
			t.Errorf("expected gin file %q in actions", f)
		}
	}

	// Verify {{.Name}} is rendered in routes.go (bug-fix check).
	routesPath := filepath.Join(wantDir, filepath.FromSlash("internal/http/routes.go"))
	for _, a := range res.Actions {
		if a.Path == routesPath {
			if strings.Contains(a.Content, "{{.Name}}") {
				t.Error("routes.go still contains unrendered {{.Name}} template tag")
			}
			if !strings.Contains(a.Content, "ginapp") {
				t.Error("routes.go does not contain the rendered project name 'ginapp'")
			}
			break
		}
	}

	// Library-generated main.go should mention gin's http package.
	mainPath := filepath.Join(wantDir, "main.go")
	for _, a := range res.Actions {
		if a.Path == mainPath {
			if !strings.Contains(a.Content, "http.NewServer") {
				t.Error("main.go should reference http.NewServer when gin is enabled")
			}
		}
	}
}

func TestPlan_GoGormLibrary(t *testing.T) {
	res, err := Plan(Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "gormapp",
		Dir:       "/tmp",
		Libraries: []string{"gorm"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := actionPaths(res.Actions)
	wantDir := res.ProjectDir

	for _, f := range []string{
		filepath.Join("internal", "db", "db.go"),
		filepath.Join("internal", "db", "models.go"),
	} {
		full := filepath.Join(wantDir, f)
		if !paths[full] {
			t.Errorf("expected gorm file %q in actions", f)
		}
	}
}

func TestPlan_GoAllLibraries(t *testing.T) {
	res, err := Plan(Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "fullapp",
		Dir:       "/tmp",
		Libraries: []string{"gin", "gorm", "sqlc"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := actionPaths(res.Actions)
	wantDir := res.ProjectDir

	expected := []string{
		filepath.Join("internal", "http", "server.go"),
		filepath.Join("internal", "http", "routes.go"),
		filepath.Join("internal", "db", "db.go"),
		filepath.Join("internal", "db", "models.go"),
		"sqlc.yaml",
		filepath.Join("db", "schema.sql"),
		filepath.Join("db", "query.sql"),
		filepath.Join("internal", "db", "README.md"),
		"main.go",
		"go.mod",
		"README.md",
	}

	for _, f := range expected {
		full := filepath.Join(wantDir, f)
		if !paths[full] {
			t.Errorf("expected file %q in actions with all libraries", f)
		}
	}
}

func TestPlan_GoCobraFramework(t *testing.T) {
	res, err := Plan(Request{
		Language:  "Go",
		Framework: "Cobra",
		Name:      "clitool",
		Dir:       "/tmp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := actionPaths(res.Actions)
	mainPath := filepath.Join(res.ProjectDir, "cmd", "clitool", "main.go")
	if !paths[mainPath] {
		t.Errorf("expected cobra main at %q, actions: %v", mainPath, pathList(res.Actions))
	}
}

func TestPlan_GoCobraWithLibraries(t *testing.T) {
	res, err := Plan(Request{
		Language:  "Go",
		Framework: "Cobra",
		Name:      "clitool",
		Dir:       "/tmp",
		Libraries: []string{"gin"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := actionPaths(res.Actions)
	// When cobra + libraries, the main.go should be at cmd/slug/main.go.
	mainPath := filepath.Join(res.ProjectDir, "cmd", "clitool", "main.go")
	if !paths[mainPath] {
		t.Errorf("expected cobra+lib main at %q, actions: %v", mainPath, pathList(res.Actions))
	}
}

func TestPlan_LaravelUsesGenerator(t *testing.T) {
	res, err := Plan(Request{
		Language:  "PHP",
		Framework: "Laravel",
		Name:      "laravelapp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Generator != "composer-laravel" {
		t.Errorf("Generator = %q, want %q", res.Generator, "composer-laravel")
	}
	if len(res.Actions) != 0 {
		t.Errorf("expected no template actions for Laravel, got %d", len(res.Actions))
	}
}

func TestPlan_DirDefaultsToDot(t *testing.T) {
	res, err := Plan(Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "app",
		Dir:       "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Dir "" should default to ".", so project dir starts with Go/app.
	want := filepath.Join("Go", "app")
	if res.ProjectDir != want {
		t.Errorf("ProjectDir = %q, want %q", res.ProjectDir, want)
	}
}

func TestPlan_GoVersionInGoMod(t *testing.T) {
	res, err := Plan(Request{
		Language:  "Go",
		Framework: "Vanilla",
		Name:      "app",
		Dir:       "/tmp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedVersion := goVersionTag()

	for _, a := range res.Actions {
		if strings.HasSuffix(a.Path, "go.mod") {
			if !strings.Contains(a.Content, "go "+expectedVersion) {
				t.Errorf("go.mod should contain 'go %s', got:\n%s", expectedVersion, a.Content)
			}
			if strings.Contains(a.Content, "go 1.22") && expectedVersion != "1.22" {
				t.Error("go.mod contains hardcoded 'go 1.22' instead of runtime version")
			}
			break
		}
	}
}

// ---------------------------------------------------------------------------
// Apply
// ---------------------------------------------------------------------------

func TestApply_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	actions := []Action{
		{Path: filepath.Join(dir, "main.go"), Content: "package main\n"},
		{Path: filepath.Join(dir, "sub", "app.go"), Content: "package sub\n"},
	}

	if err := Apply(actions, false); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	for _, a := range actions {
		data, err := os.ReadFile(a.Path)
		if err != nil {
			t.Errorf("file %q not created: %v", a.Path, err)
			continue
		}
		if string(data) != a.Content {
			t.Errorf("file %q content = %q, want %q", a.Path, string(data), a.Content)
		}
	}
}

func TestApply_DryRunNoFiles(t *testing.T) {
	dir := t.TempDir()
	actions := []Action{
		{Path: filepath.Join(dir, "should-not-exist.go"), Content: "package main\n"},
	}

	if err := Apply(actions, true); err != nil {
		t.Fatalf("Apply (dryRun) failed: %v", err)
	}

	if _, err := os.Stat(actions[0].Path); !os.IsNotExist(err) {
		t.Error("dryRun should not create files, but file exists")
	}
}

func TestApply_ErrorIfFileExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing.go")
	if err := os.WriteFile(existing, []byte("old"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	actions := []Action{
		{Path: existing, Content: "new"},
	}

	err := Apply(actions, false)
	if err == nil {
		t.Fatal("expected error when file already exists, got nil")
	}
	if !strings.Contains(err.Error(), "file already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// goLibrariesReadme
// ---------------------------------------------------------------------------

func TestGoLibrariesReadme(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		contains []string
		absent   []string
	}{
		{
			name:     "gin only",
			data:     Data{Name: "TestApp", UseGin: true},
			contains: []string{"# TestApp", "- Gin"},
			absent:   []string{"- Gorm", "- Sqlc"},
		},
		{
			name:     "gorm only",
			data:     Data{Name: "TestApp", UseGorm: true},
			contains: []string{"- Gorm"},
			absent:   []string{"- Gin", "- Sqlc"},
		},
		{
			name:     "sqlc only",
			data:     Data{Name: "TestApp", UseSqlc: true},
			contains: []string{"- Sqlc", "sqlc generate"},
			absent:   []string{"- Gin", "- Gorm"},
		},
		{
			name:     "all libraries",
			data:     Data{Name: "App", UseGin: true, UseGorm: true, UseSqlc: true},
			contains: []string{"# App", "- Gin", "- Gorm", "- Sqlc", "sqlc generate"},
		},
		{
			name:     "gin and gorm",
			data:     Data{Name: "App", UseGin: true, UseGorm: true},
			contains: []string{"- Gin", "- Gorm"},
			absent:   []string{"- Sqlc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goLibrariesReadme(tt.data)
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("expected readme to contain %q, got:\n%s", s, got)
				}
			}
			for _, s := range tt.absent {
				if strings.Contains(got, s) {
					t.Errorf("expected readme NOT to contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// goLibrariesMod
// ---------------------------------------------------------------------------

func TestGoLibrariesMod(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		contains []string
		absent   []string
	}{
		{
			name:     "gin only",
			data:     Data{Module: "mymod", GoVersion: "1.23", UseGin: true},
			contains: []string{"module mymod", "go 1.23", "gin-gonic/gin"},
			absent:   []string{"gorm.io"},
		},
		{
			name:     "gorm only",
			data:     Data{Module: "mymod", GoVersion: "1.23", UseGorm: true},
			contains: []string{"gorm.io/driver/sqlite", "gorm.io/gorm"},
			absent:   []string{"gin-gonic"},
		},
		{
			name:     "both",
			data:     Data{Module: "mymod", GoVersion: "1.23", UseGin: true, UseGorm: true},
			contains: []string{"gin-gonic/gin", "gorm.io/gorm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goLibrariesMod(tt.data)
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("expected go.mod to contain %q, got:\n%s", s, got)
				}
			}
			for _, s := range tt.absent {
				if strings.Contains(got, s) {
					t.Errorf("expected go.mod NOT to contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// goLibrariesMain
// ---------------------------------------------------------------------------

func TestGoLibrariesMain(t *testing.T) {
	tests := []struct {
		name      string
		data      Data
		framework string
		contains  []string
		absent    []string
	}{
		{
			name:      "gin only",
			data:      Data{Module: "mymod", UseGin: true},
			framework: "Vanilla",
			contains:  []string{"package main", `"mymod/internal/http"`, "http.NewServer()", "server.Run"},
			absent:    []string{`"mymod/internal/db"`},
		},
		{
			name:      "gorm only",
			data:      Data{Module: "mymod", UseGorm: true},
			framework: "Vanilla",
			contains:  []string{`"mymod/internal/db"`, "db.Open()", "db.AutoMigrate"},
			absent:    []string{`"mymod/internal/http"`},
		},
		{
			name:      "sqlc only",
			data:      Data{Module: "mymod", UseSqlc: true},
			framework: "Vanilla",
			contains:  []string{"sqlc generate"},
			absent:    []string{`"mymod/internal/http"`, `"mymod/internal/db"`},
		},
		{
			name:      "gin and gorm",
			data:      Data{Module: "mymod", UseGin: true, UseGorm: true},
			framework: "Vanilla",
			contains:  []string{`"mymod/internal/http"`, `"mymod/internal/db"`, "db.Open()", "http.NewServer()"},
		},
		{
			name:      "all three",
			data:      Data{Module: "mymod", UseGin: true, UseGorm: true, UseSqlc: true},
			framework: "Vanilla",
			contains:  []string{`"mymod/internal/http"`, `"mymod/internal/db"`, "sqlc generate"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goLibrariesMain(tt.data, tt.framework)
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("expected main to contain %q, got:\n%s", s, got)
				}
			}
			for _, s := range tt.absent {
				if strings.Contains(got, s) {
					t.Errorf("expected main NOT to contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// actionPaths returns a set of all paths in the given actions.
func actionPaths(actions []Action) map[string]bool {
	m := make(map[string]bool, len(actions))
	for _, a := range actions {
		m[a.Path] = true
	}
	return m
}

// pathList returns a slice of paths for error messages.
func pathList(actions []Action) []string {
	out := make([]string, len(actions))
	for i, a := range actions {
		out[i] = a.Path
	}
	return out
}
