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

Run commands
- Run TUI: `go run ./cmd/project-initiator`
- Non-TUI mode: `go run ./cmd/project-initiator --no-tui --lang Go --framework Vanilla --name myapp`

Test commands
- Run all tests: `go test ./...`
- Run a single package: `go test ./internal/scaffold`
- Run a single test by name (package): `go test ./internal/scaffold -run TestName`
- Run all tests with name match (repo): `go test ./... -run TestName`
- Re-run without cache: `go test ./internal/scaffold -run TestName -count=1`

Lint / static analysis
- Go formatting: `gofmt -w <files>`
- Vet: `go vet ./...`
- No additional linters are configured in this repo.

Project generation notes
- Default output path: `/mnt/Dev/Projects/{language}/{project_name}`.
- Language folder keeps the original casing but is sanitized for path safety.
- Laravel uses Composer generator: `composer create-project laravel/laravel <projectDir>`.

Code style guidelines

Formatting
- Always run `gofmt` on changed Go files.
- Keep line length reasonable; prefer small helper functions over long blocks.

Imports
- Use `gofmt` to group/format imports.
- Standard lib first, then third-party, then local packages.

Naming
- Exported identifiers: `CamelCase`.
- Unexported identifiers: `camelCase`.
- Avoid abbreviations unless commonly accepted (e.g., `ctx`, `cfg`).
- Prefer explicit names: `projectDir`, `frameworkList`, `panelHeight`.

Types and structures
- Prefer small structs with clear responsibilities.
- Pass config and options via struct values instead of global state.
- Keep `internal/*` packages cohesive and focused.

Error handling
- Return errors early; avoid deep nesting.
- Use `fmt.Errorf` to add context; do not swallow errors.
- For user-facing errors, print to stderr and exit with non-zero status.

Logging / output
- Use `fmt.Fprintln(os.Stdout, ...)` for user success output.
- Use `fmt.Fprintln(os.Stderr, ...)` for errors.
- Keep CLI output concise and actionable.

CLI behavior
- Flags are parsed in `internal/flags`.
- `--no-tui` requires `--name`.
- Keep TUI non-destructive and cancelable.

TUI conventions (Bubble Tea)
- Use `list` for selectable menus.
- Use `textinput` for user text entry.
- Handle `ctrl+c` and `esc` to exit.
- Do not bind destructive actions to Backspace in the name input step.

Scaffolding conventions
- Templates are stored in `internal/scaffold/templates.go`.
- Each option should include a `Language` and `Framework`.
- For template-based scaffolds, provide minimal runnable starters.
- For generator-based scaffolds (e.g., Laravel), use `Generator` and skip templates.

File paths and IO
- Write files using `os.WriteFile` with `0o644`.
- Create directories using `os.MkdirAll` with `0o755`.
- Do not overwrite existing files; return a clear error.

Adding new templates
- Add a new `Option` entry in `internal/scaffold/templates.go`.
- Keep template files minimal; prefer direct `main` or `app` entrypoints.
- Include a short `README.md` for each template.
- If dependency tooling is required, document it in the README content.

Go module considerations
- Module name: `project-initiator` (see `go.mod`).
- Go version: `1.25.4`.

Platform assumptions
- Default paths assume a Linux environment under `/mnt/Dev/Projects`.
- If adding platform-specific behavior, guard it clearly and document it.

When in doubt
- Prefer consistency with existing code patterns.
- Keep user-facing messages succinct.
- Avoid introducing new dependencies unless necessary.
