package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const defaultConfigFilename = ".project-initiator.json"

type Config struct {
	DefaultLanguage  string `json:"defaultLanguage"`
	DefaultFramework string `json:"defaultFramework"`
	DefaultDir       string `json:"defaultDir"`
}

func Default() Config {
	return Config{
		DefaultLanguage:  "Go",
		DefaultFramework: "Cobra",
		DefaultDir:       "/mnt/Dev/Projects",
	}
}

func Load(path string) (Config, error) {
	if path == "" {
		path = defaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return applyDefaults(cfg), nil
}

func Save(path string, cfg Config) error {
	if path == "" {
		path = defaultConfigPath()
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfigFilename
	}

	return filepath.Join(home, defaultConfigFilename)
}

func applyDefaults(cfg Config) Config {
	defaults := Default()
	if cfg.DefaultLanguage == "" {
		cfg.DefaultLanguage = defaults.DefaultLanguage
	}
	if cfg.DefaultFramework == "" {
		cfg.DefaultFramework = defaults.DefaultFramework
	}
	if cfg.DefaultDir == "" {
		cfg.DefaultDir = defaults.DefaultDir
	}

	return cfg
}
