# project-initiator

A terminal-based project scaffolding tool built with Go. Select a language, framework, and optional libraries through an animated TUI wizard, and get a ready-to-go starter project with `git init` already done.

## Features

- **Interactive TUI** powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea) with animated ASCII art title, spring-animated panel entrance, and smooth stage transitions
- **6 languages, 12 framework templates** covering Go, JavaScript, Node.js, Bun, Python, and PHP
- **Go library add-ons** &mdash; optionally layer in Gin, Gorm, and/or Sqlc on any Go template
- **Non-interactive mode** for CI/scripting via `--no-tui` and CLI flags
- **Dry-run mode** to preview the plan without writing files
- **Persistent config** remembers your last language, framework, and output directory
- **Adaptive colors** &mdash; light and dark terminal themes supported via `lipgloss.AdaptiveColor`

## Supported Languages & Frameworks

| Language   | Frameworks             |
|------------|------------------------|
| Go         | Vanilla, Cobra         |
| JavaScript | Vanilla                |
| Node.js    | Express, Hono, NestJS  |
| Bun        | Vanilla, Bun (server)  |
| Python     | Vanilla, FastAPI       |
| PHP        | Vanilla, Laravel*      |

\* Laravel uses `composer create-project` under the hood.

### Go Library Add-ons

When scaffolding a Go project, you can optionally include:

| Library | What it adds |
|---------|-------------|
| **Gin** | HTTP server with router, health endpoint, and route registration (`internal/http/`) |
| **Gorm** | SQLite database layer with auto-migration and a sample model (`internal/db/`) |
| **Sqlc** | SQL schema, queries, and `sqlc.yaml` config for type-safe SQL (`db/`, `internal/db/`) |

Libraries can be combined freely. When any library is selected, the generated `main.go`, `go.mod`, and `README.md` are replaced with library-aware versions.

## Installation

Requires Go 1.23 or later.

**From source:**

```bash
go install github.com/your-username/project-initiator/cmd/project-initiator@latest
```

**Or clone and build:**

```bash
git clone https://github.com/your-username/project-initiator.git
cd project-initiator
go build -o project-initiator ./cmd/project-initiator
```

## Usage

### TUI Mode (default)

Simply run the binary with no arguments to launch the interactive wizard:

```bash
./project-initiator
```

The wizard walks you through:

1. **Language** &mdash; pick from the supported list
2. **Framework** &mdash; choose a framework/template for that language
3. **Libraries** &mdash; (Go only) optionally add Gin, Gorm, Sqlc
4. **Project name** &mdash; enter the name for your new project
5. **Confirm** &mdash; review your choices and scaffold

### CLI Mode (non-interactive)

Pass all required values as flags to skip the TUI entirely:

```bash
./project-initiator --no-tui --lang Go --framework Cobra --name my-app
```

If only some flags are provided (and `--no-tui` is not set), the TUI opens pre-filled with those values.

### Dry Run

Preview what files would be created without writing anything:

```bash
./project-initiator --no-tui --lang Go --framework Vanilla --name demo --dry-run
```

### CLI Flags

| Flag          | Description                              | Default          |
|---------------|------------------------------------------|------------------|
| `--lang`      | Language to scaffold                     | From config      |
| `--framework` | Framework template to use                | From config      |
| `--name`      | Project name                             | _(interactive)_  |
| `--dir`       | Base directory for the new project       | From config      |
| `--config`    | Path to config file                      | `~/.project-initiator.json` |
| `--dry-run`   | Print planned actions without writing    | `false`          |
| `--no-tui`    | Disable TUI; requires `--name`           | `false`          |

## Configuration

Settings are stored in `~/.project-initiator.json` and automatically updated after each run:

```json
{
  "defaultLanguage": "Go",
  "defaultFramework": "Cobra",
  "defaultDir": "~/Projects"
}
```

If the config file doesn't exist, defaults are used:

- **Language:** Go
- **Framework:** Cobra
- **Directory:** `~/Projects` (resolved via `$HOME`)

## Project Structure

```
project-initiator/
├── cmd/project-initiator/
│   └── main.go                  # Entry point
└── internal/
    ├── app/run.go               # Orchestration: parse flags, run TUI or CLI, scaffold, git init
    ├── config/
    │   ├── config.go            # Load/save JSON config with defaults
    │   └── config_test.go
    ├── domain/models.go         # Shared types: Framework, Template, Library, Plan, Action, Project
    ├── errors/errors.go         # Sentinel errors
    ├── flags/
    │   ├── flags.go             # CLI flag parsing
    │   └── flags_test.go
    ├── library/manager.go       # Go library code generation (Gin, Gorm, Sqlc)
    ├── scaffold/
    │   ├── frameworks.go        # All 12 framework template definitions
    │   ├── scaffold.go          # Planner and Applier: resolve templates, write files
    │   └── scaffold_test.go
    ├── template/renderer.go     # Go text/template wrapper
    └── ui/
        ├── animation.go         # ASCII art title, animated border with gradient glow spark
        ├── helpers.go           # List builders, stage progress, rendering helpers, transitions
        ├── styles.go            # Lipgloss styles, list delegate, adaptive color palette
        ├── wizard.go            # Bubble Tea model: Init, Update, View, stage handlers, springs
        └── wizard_test.go
```

## Adding a New Template

1. Open `internal/scaffold/frameworks.go`
2. Add a new `domain.Framework` entry to the `Frameworks` slice:

```go
{
    Language: "Ruby",
    Name:     "Sinatra",
    Templates: []domain.Template{
        {
            RelativePath: "app.rb",
            Content:      "require 'sinatra'\n\nget '/' do\n  'Hello from {{.Name}}'\nend\n",
        },
        {
            RelativePath: "Gemfile",
            Content:      "source 'https://rubygems.org'\ngem 'sinatra'\n",
        },
    },
},
```

Templates use Go `text/template` syntax. Available variables:

| Variable       | Description                                    |
|----------------|------------------------------------------------|
| `{{.Name}}`    | Project name as entered by the user            |
| `{{.PackageName}}` | URL/package-safe slug of the name         |
| `{{.Module}}`  | Go module path (Go projects only)              |
| `{{.GoVersion}}` | Current Go version (Go projects only)        |

The new language/framework will automatically appear in the TUI wizard.

## Development

**Build:**

```bash
go build -o project-initiator ./cmd/project-initiator
```

**Run tests:**

```bash
go test ./...
```

**Format & vet:**

```bash
gofmt -w .
go vet ./...
```

## Dependencies

- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) &mdash; Terminal UI framework
- [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) &mdash; TUI components (list, text input, help, progress, key bindings)
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) &mdash; Terminal styling and layout
- [charmbracelet/harmonica](https://github.com/charmbracelet/harmonica) &mdash; Spring-based animation physics

## License

MIT
