package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var nameSlug = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)

type Data struct {
	Name        string
	PackageName string
	Module      string
	Framework   string
	UseGin      bool
	UseGorm     bool
	UseSqlc     bool
}

type Request struct {
	Language  string
	Framework string
	Name      string
	Dir       string
	DryRun    bool
	Libraries []string
}

type Action struct {
	Path    string
	Content string
}

type PlanResult struct {
	ProjectDir string
	Actions    []Action
	Generator  string
}

func Plan(req Request) (PlanResult, error) {
	opt, err := findOption(req.Language, req.Framework)
	if err != nil {
		return PlanResult{}, err
	}

	projectName := strings.TrimSpace(req.Name)
	if projectName == "" {
		return PlanResult{}, errors.New("project name is required")
	}

	projectSlug := slugify(projectName)
	selected := map[string]bool{}
	for _, lib := range req.Libraries {
		selected[strings.ToLower(strings.TrimSpace(lib))] = true
	}
	data := Data{
		Name:        projectName,
		PackageName: projectSlug,
		Module:      projectSlug,
		Framework:   opt.Framework,
		UseGin:      selected["gin"],
		UseGorm:     selected["gorm"],
		UseSqlc:     selected["sqlc"],
	}

	rootDir := strings.TrimSpace(req.Dir)
	if rootDir == "" {
		rootDir = "."
	}
	rootDir = filepath.Clean(rootDir)
	languageDir := cleanLanguageDir(req.Language)
	projectDir := filepath.Join(rootDir, languageDir, projectSlug)

	actions := make([]Action, 0, len(opt.Templates))
	for _, tmpl := range opt.Templates {
		content, err := render(tmpl.Content, data)
		if err != nil {
			return PlanResult{}, err
		}
		path := filepath.Join(projectDir, filepath.FromSlash(tmpl.RelativePath))
		actions = append(actions, Action{Path: path, Content: content})
	}

	if strings.EqualFold(opt.Language, "go") {
		if data.UseGin || data.UseGorm || data.UseSqlc {
			mainPath := filepath.Join(projectDir, "main.go")
			if strings.EqualFold(opt.Framework, "cobra") {
				mainPath = filepath.Join(projectDir, "cmd", projectSlug, "main.go")
			}
			actions = append(actions, Action{Path: mainPath, Content: goLibrariesMain(data, opt.Framework)})
			actions = append(actions, Action{Path: filepath.Join(projectDir, "go.mod"), Content: goLibrariesMod(data)})
			actions = append(actions, Action{Path: filepath.Join(projectDir, "README.md"), Content: goLibrariesReadme(data)})
		}
		if data.UseGin {
			actions = append(actions, Action{Path: filepath.Join(projectDir, "internal/http/server.go"), Content: goGinServer})
			actions = append(actions, Action{Path: filepath.Join(projectDir, "internal/http/routes.go"), Content: goGinRoutes})
		}
		if data.UseGorm {
			actions = append(actions, Action{Path: filepath.Join(projectDir, "internal/db/db.go"), Content: goGormDB})
			actions = append(actions, Action{Path: filepath.Join(projectDir, "internal/db/models.go"), Content: goGormModels})
		}
		if data.UseSqlc {
			actions = append(actions, Action{Path: filepath.Join(projectDir, "sqlc.yaml"), Content: goSqlcConfig})
			actions = append(actions, Action{Path: filepath.Join(projectDir, "db/schema.sql"), Content: goSqlcSchema})
			actions = append(actions, Action{Path: filepath.Join(projectDir, "db/query.sql"), Content: goSqlcQuery})
			actions = append(actions, Action{Path: filepath.Join(projectDir, "internal/db/README.md"), Content: goSqlcReadme})
		}
	}

	return PlanResult{ProjectDir: projectDir, Actions: actions, Generator: opt.Generator}, nil
}

func Apply(actions []Action, dryRun bool) error {
	for _, action := range actions {
		if _, err := os.Stat(action.Path); err == nil {
			return fmt.Errorf("file already exists: %s", action.Path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	for _, action := range actions {
		if dryRun {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(action.Path), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(action.Path, []byte(action.Content), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func render(source string, data Data) (string, error) {
	tmpl, err := template.New("template").Parse(source)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
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

func findOption(lang string, framework string) (Option, error) {
	lang = strings.TrimSpace(lang)
	framework = strings.TrimSpace(framework)

	for _, opt := range Options {
		if strings.EqualFold(opt.Language, lang) && strings.EqualFold(opt.Framework, framework) {
			return opt, nil
		}
	}

	return Option{}, fmt.Errorf("no template for %s / %s", lang, framework)
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

const goGinServer = "package http\n\nimport (\n\t\"net/http\"\n\n\t\"github.com/gin-gonic/gin\"\n)\n\nfunc NewServer() *gin.Engine {\n\trouter := gin.New()\n\trouter.Use(gin.Recovery())\n\n\tRegisterRoutes(router)\n\n\trouter.GET(\"/health\", func(c *gin.Context) {\n\t\tc.JSON(http.StatusOK, gin.H{\"status\": \"ok\"})\n\t})\n\n\treturn router\n}\n"

const goGinRoutes = "package http\n\nimport (\n\t\"net/http\"\n\n\t\"github.com/gin-gonic/gin\"\n)\n\nfunc RegisterRoutes(router *gin.Engine) {\n\trouter.GET(\"/\", func(c *gin.Context) {\n\t\tc.JSON(http.StatusOK, gin.H{\"message\": \"hello from {{.Name}}\"})\n\t})\n}\n"

const goGormDB = "package db\n\nimport (\n\t\"gorm.io/driver/sqlite\"\n\t\"gorm.io/gorm\"\n)\n\nfunc Open() (*gorm.DB, error) {\n\treturn gorm.Open(sqlite.Open(\"app.db\"), &gorm.Config{})\n}\n"

const goGormModels = "package db\n\nimport \"gorm.io/gorm\"\n\ntype User struct {\n\tID   uint\n\tName string\n}\n\nfunc AutoMigrate(db *gorm.DB) error {\n\treturn db.AutoMigrate(&User{})\n}\n"

const goSqlcConfig = "version: \"2\"\nsql:\n  - engine: \"sqlite\"\n    schema: \"db/schema.sql\"\n    queries: \"db/query.sql\"\n    gen:\n      go:\n        package: \"db\"\n        out: \"internal/db\"\n"

const goSqlcSchema = "CREATE TABLE users (\n  id INTEGER PRIMARY KEY,\n  name TEXT NOT NULL\n);\n"

const goSqlcQuery = "-- name: ListUsers :many\nSELECT id, name FROM users;\n\n-- name: CreateUser :exec\nINSERT INTO users (name) VALUES (?);\n"

const goSqlcReadme = "# SQLC\n\nRun `sqlc generate` to generate Go code into internal/db.\n"

func goLibrariesReadme(data Data) string {
	lines := []string{
		"# " + data.Name,
		"",
		"Generated by project-initiator.",
		"",
		"Included libraries:",
	}
	if data.UseGin {
		lines = append(lines, "- Gin")
	}
	if data.UseGorm {
		lines = append(lines, "- Gorm")
	}
	if data.UseSqlc {
		lines = append(lines, "- Sqlc")
		lines = append(lines, "", "Run: `sqlc generate`")
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func goLibrariesMod(data Data) string {
	lines := []string{"module " + data.Module, "", "go 1.22", "", "require ("}
	if data.UseGin {
		lines = append(lines, "\tgithub.com/gin-gonic/gin v1.10.0")
	}
	if data.UseGorm {
		lines = append(lines, "\tgorm.io/driver/sqlite v1.5.7")
		lines = append(lines, "\tgorm.io/gorm v1.25.12")
	}
	lines = append(lines, ")")
	return strings.Join(lines, "\n") + "\n"
}

func goLibrariesMain(data Data, framework string) string {
	usesGin := data.UseGin
	usesGorm := data.UseGorm
	usesSqlc := data.UseSqlc

	imports := []string{"\"fmt\""}
	if usesGin {
		imports = append(imports, "\""+data.Module+"/internal/http\"")
	}
	if usesGorm {
		imports = append(imports, "\""+data.Module+"/internal/db\"")
	}

	body := []string{}
	body = append(body, "func run() error {")
	body = append(body, "\tfmt.Println(\"starting\")")
	if usesGorm {
		body = append(body, "\tdbConn, err := db.Open()")
		body = append(body, "\tif err != nil {\n\t\treturn err\n\t}")
		body = append(body, "\tif err := db.AutoMigrate(dbConn); err != nil {\n\t\treturn err\n\t}")
	}
	if usesSqlc {
		body = append(body, "\t// Run: sqlc generate")
	}
	if usesGin {
		body = append(body, "\tserver := http.NewServer()")
		if usesGorm {
			body = append(body, "\t_ = dbConn")
		}
		body = append(body, "\treturn server.Run(\":3000\")")
	} else {
		body = append(body, "\treturn nil")
	}
	body = append(body, "}")

	mainBody := []string{"func main() {", "\tif err := run(); err != nil {", "\t\tfmt.Println(\"error:\", err)", "\t}", "}"}

	code := []string{"package main", "", "import ("}
	for _, imp := range imports {
		code = append(code, "\t"+imp)
	}
	code = append(code, ")", "", strings.Join(body, "\n"), "", strings.Join(mainBody, "\n"), "")

	return strings.Join(code, "\n")
}
