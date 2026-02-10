# AGENTS.md

This file guides agentic coding assistants working in this repository.

Project summary
- Go CLI tool that scaffolds projects via a Bubble Tea TUI.
- Entry point: `cmd/project-initiator/main.go`.
- Core logic: `internal/app`, `internal/scaffold`, `internal/ui`.
- Config stored in `~/.project-initiator.json`.

Important repository rules
- No Cursor or Copilot instruction files are present.
- Follow Go standard formatting and conventions unless specified here.

Build commands
- Build CLI: `go build ./cmd/project-initiator`
- Build release binary: `go build -o bin/project-initiator ./cmd/project-initiator`
- Build with custom name: `go build -o scaffold-wizard ./cmd/project-initiator`

Run commands
- Run TUI: `go run ./cmd/project-initiator`
- Non-TUI mode: `go run ./cmd/project-initiator --no-tui --lang Go --framework Vanilla --name myapp`

Test commands
- Run all tests: `go test ./...`
- Run a single package: `go test ./internal/scaffold`
- Run a single test by name (package): `go test ./internal/scaffold -run TestName`
- Run all tests with name match (repo): `go test ./... -run TestName`
- Re-run without cache: `go test ./internal/scaffold -run TestName -count=1`
- Run tests with verbose output: `go test ./... -v`
- Run tests with coverage: `go test ./... -cover`

Lint / static analysis
- Go formatting: `gofmt -w <files>`
- Vet: `go vet ./...`
- No additional linters are configured in this repo.

Project generation notes
- Default output path: `~/Projects/{language}/{project_name}`.
- Language folder keeps the original casing but is sanitized for path safety.
- Laravel uses Composer generator: `composer create-project laravel/laravel <projectDir>`.

Code style guidelines

Formatting
- Always run `gofmt` on changed Go files before committing.
- Keep line length reasonable; prefer small helper functions over long blocks.
- Use 4-space indentation (standard Go style).

Imports
- Use `gofmt` to group/format imports automatically.
- Standard library imports first (e.g., `fmt`, `os`, `path/filepath`).
- Third-party imports second (e.g., `github.com/charmbracelet/bubbletea`).
- Local project imports last (e.g., `project-initiator/internal/scaffold`).
- Group imports with blank lines between groups.

Naming conventions
- Exported identifiers: `CamelCase` (e.g., `PlanResult`, `FindOption`).
- Unexported identifiers: `camelCase` (e.g., `slugify`, `cleanLanguageDir`).
- Avoid abbreviations unless commonly accepted (e.g., `ctx`, `cfg`, `req`).
- Prefer explicit names: `projectDir`, `frameworkList`, `panelHeight`.
- Test functions: `TestCamelCase` or `TestPackage_FunctionName`.
- Test tables: use `tests` or `tt` for the slice, `tc` or `test` for the element.

Types and structures
- Prefer small structs with clear, single responsibilities.
- Pass config and options via struct values instead of global state.
- Keep `internal/*` packages cohesive and focused.
- Use `iota` for enumerated stage constants.
- Embed interfaces where appropriate (e.g., `tea.Model` in wizard).

Error handling
- Return errors early; avoid deep nesting.
- Use `fmt.Errorf` to add context; do not swallow errors.
- For user-facing errors, print to `os.Stderr` and exit with non-zero status.
- Check for specific error types when possible.
- Validate inputs at function boundaries.

Logging / output
- Use `fmt.Fprintln(os.Stdout, ...)` for user success output.
- Use `fmt.Fprintln(os.Stderr, ...)` for errors.
- Keep CLI output concise and actionable.
- Avoid debug print statements in production code.

CLI behavior
- Flags are parsed in `internal/flags`.
- `--no-tui` requires `--name` flag.
- Keep TUI non-destructive and cancelable.
- Handle `ctrl+c` and `esc` to exit gracefully.

TUI conventions (Bubble Tea)
- Use `list` for selectable menus.
- Use `textinput` for user text entry.
- Do not bind destructive actions to Backspace in the name input step.
- Use `lipgloss` for styling with consistent color scheme.
- Keep animations smooth with frame-based rendering.

Scaffolding conventions
- Templates are stored in `internal/scaffold/frameworks.go`.
- Each option should include a `Language` and `Framework`.
- For template-based scaffolds, provide minimal runnable starters.
- For generator-based scaffolds (e.g., Laravel), use `Generator` field and skip templates.

File paths and IO
- Write files using `os.WriteFile` with `0o644`.
- Create directories using `os.MkdirAll` with `0o755`.
- Do not overwrite existing files; return a clear error.
- Use `filepath.Join` for path construction (cross-platform).

Adding new templates
- Add a new `Framework` entry in `internal/scaffold/frameworks.go`.
- Keep template files minimal; prefer direct `main` or `app` entrypoints.
- Include a short `README.md` for each template.
- If dependency tooling is required, document it in the README content.
- Update tests for new templates in `scaffold_test.go`.

Go module considerations
- Module name: `project-initiator` (see `go.mod`).
- Go version: `1.25.4`.
- Dependencies: `bubbletea`, `bubbles`, `lipgloss` (charmbracelet stack).

Platform assumptions
- Default project directory uses `~/Projects` (computed via `os.UserHomeDir()`).
- If adding platform-specific behavior, guard it clearly and document it.

Testing guidelines
- Use table-driven tests with `tests := []struct{...}{}`.
- Name test cases descriptively: `name: "empty name returns error"`.
- Use `t.Run(tc.name, func(t *testing.T){...})` for subtests.
- Test error cases explicitly, not just happy paths.
- Clean up test files: `defer os.RemoveAll(tempDir)`.
- Use `t.Fatalf` for setup failures, `t.Errorf` for assertion failures.
- Group related tests with section comments: `// slugify`.

Git conventions
- Do not commit the `bin/` directory (it's in `.gitignore`).
- Do not commit `scaffold-wizard` or other build artifacts.
- Keep commits focused and atomic.

When in doubt
- Prefer consistency with existing code patterns.
- Keep user-facing messages succinct.
- Avoid introducing new dependencies unless necessary.
- Run `go test ./...` before committing changes.
