package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.DefaultLanguage != "Go" {
		t.Errorf("DefaultLanguage = %q, want %q", cfg.DefaultLanguage, "Go")
	}
	if cfg.DefaultFramework != "Cobra" {
		t.Errorf("DefaultFramework = %q, want %q", cfg.DefaultFramework, "Cobra")
	}

	// DefaultDir should be based on the user's home directory, not hardcoded.
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot determine home dir: %v", err)
	}
	wantDir := filepath.Join(home, "Projects")
	if cfg.DefaultDir != wantDir {
		t.Errorf("DefaultDir = %q, want %q", cfg.DefaultDir, wantDir)
	}
}

func TestLoad(t *testing.T) {
	defaults := Default()

	t.Run("non-existent file returns defaults", func(t *testing.T) {
		cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.json"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != defaults {
			t.Errorf("got %+v, want %+v", cfg, defaults)
		}
	})

	t.Run("valid JSON file is loaded correctly", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		want := Config{
			DefaultLanguage:  "Rust",
			DefaultFramework: "Actix",
			DefaultDir:       "/home/user/projects",
		}
		writeJSON(t, path, want)

		got, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("partial JSON gets defaults applied", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		// Only set the language; framework and dir should get defaults.
		partial := map[string]string{"defaultLanguage": "Python"}
		writeJSON(t, path, partial)

		got, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.DefaultLanguage != "Python" {
			t.Errorf("DefaultLanguage = %q, want %q", got.DefaultLanguage, "Python")
		}
		if got.DefaultFramework != defaults.DefaultFramework {
			t.Errorf("DefaultFramework = %q, want default %q", got.DefaultFramework, defaults.DefaultFramework)
		}
		if got.DefaultDir != defaults.DefaultDir {
			t.Errorf("DefaultDir = %q, want default %q", got.DefaultDir, defaults.DefaultDir)
		}
	})

	t.Run("empty JSON object returns all defaults", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		if err := os.WriteFile(path, []byte(`{}`), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		got, err := Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != defaults {
			t.Errorf("got %+v, want %+v", got, defaults)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		if err := os.WriteFile(path, []byte(`{not valid json`), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		_, err := Load(path)
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})

	t.Run("empty path does not panic", func(t *testing.T) {
		// An empty path falls back to defaultConfigPath(). The file may or may
		// not exist on the host, but the call must not panic.
		_, _ = Load("")
	})
}

func TestSave(t *testing.T) {
	t.Run("saves to file and reads back correctly", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		want := Config{
			DefaultLanguage:  "TypeScript",
			DefaultFramework: "Express",
			DefaultDir:       "/tmp/projects",
		}

		if err := Save(path, want); err != nil {
			t.Fatalf("Save() error: %v", err)
		}

		got, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error after Save: %v", err)
		}
		if got != want {
			t.Errorf("round-trip failed: got %+v, want %+v", got, want)
		}
	})

	t.Run("creates parent directories if needed", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a", "b", "c", "config.json")

		cfg := Default()
		if err := Save(path, cfg); err != nil {
			t.Fatalf("Save() error: %v", err)
		}

		if _, err := os.Stat(path); err != nil {
			t.Fatalf("file was not created: %v", err)
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		original := Config{
			DefaultLanguage:  "Java",
			DefaultFramework: "Spring",
			DefaultDir:       "/opt/projects",
		}
		if err := Save(path, original); err != nil {
			t.Fatalf("first Save() error: %v", err)
		}

		updated := Config{
			DefaultLanguage:  "Kotlin",
			DefaultFramework: "Ktor",
			DefaultDir:       "/home/dev",
		}
		if err := Save(path, updated); err != nil {
			t.Fatalf("second Save() error: %v", err)
		}

		got, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error: %v", err)
		}
		if got != updated {
			t.Errorf("got %+v, want %+v", got, updated)
		}
	})
}

func TestApplyDefaults(t *testing.T) {
	defaults := Default()

	tests := []struct {
		name string
		in   Config
		want Config
	}{
		{
			name: "all fields empty gets all defaults",
			in:   Config{},
			want: defaults,
		},
		{
			name: "all fields set stays unchanged",
			in: Config{
				DefaultLanguage:  "Ruby",
				DefaultFramework: "Rails",
				DefaultDir:       "/srv/apps",
			},
			want: Config{
				DefaultLanguage:  "Ruby",
				DefaultFramework: "Rails",
				DefaultDir:       "/srv/apps",
			},
		},
		{
			name: "only language set",
			in:   Config{DefaultLanguage: "Elixir"},
			want: Config{
				DefaultLanguage:  "Elixir",
				DefaultFramework: defaults.DefaultFramework,
				DefaultDir:       defaults.DefaultDir,
			},
		},
		{
			name: "only framework set",
			in:   Config{DefaultFramework: "Gin"},
			want: Config{
				DefaultLanguage:  defaults.DefaultLanguage,
				DefaultFramework: "Gin",
				DefaultDir:       defaults.DefaultDir,
			},
		},
		{
			name: "only dir set",
			in:   Config{DefaultDir: "/custom/dir"},
			want: Config{
				DefaultLanguage:  defaults.DefaultLanguage,
				DefaultFramework: defaults.DefaultFramework,
				DefaultDir:       "/custom/dir",
			},
		},
		{
			name: "language and dir set, framework defaults",
			in: Config{
				DefaultLanguage: "C#",
				DefaultDir:      "/dotnet",
			},
			want: Config{
				DefaultLanguage:  "C#",
				DefaultFramework: defaults.DefaultFramework,
				DefaultDir:       "/dotnet",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyDefaults(tt.in)
			if got != tt.want {
				t.Errorf("applyDefaults(%+v) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}

// writeJSON is a test helper that marshals v to JSON and writes it to path.
func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}
