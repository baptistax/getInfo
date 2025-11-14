package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// File names are kept in Portuguese for local operators/users.
const (
	CsvName    = "inventario.csv"
	ErrLogName = "inventario_erros.txt"
)

// AnyDeskPassword is the default unattended access password to set via CLI.
// IMPORTANT:
//   - Change this before using in any real environment.
//   - For public distributions, prefer loading this from a config file or
//     environment variable instead of hardcoding it here.
const AnyDeskPassword = "CHANGE_ME"

// defaultConfigJSON holds the default configuration file contents.
// It is written only if config/config.json does not exist yet.
var defaultConfigJSON = []byte(`{
  "collection": {
    "hardware": true,
    "gpu": true,
    "ram": true,
    "osInfo": true,
    "installedSoftware": false,
    "remoteDesktop": true,
    "antivirus": true,
    "network": false
  },
  "output": {
    "excel": true,
    "csv": false,
    "json": false,
    "showSummaryInConsole": true
  },
  "ui": {
    "interactiveIfNoArgs": true
  }
}
`)

// ensureStructure creates required directories and default files if they do not exist.
// It never deletes or overwrites existing files.
func ensureStructure() error {
	// Directories to ensure (can be used for future features/output layout).
	dirs := []string{
		"config",
		"output",
		"logs",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	// Config file (only created if missing)
	configPath := filepath.Join("config", "config.json")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(configPath, defaultConfigJSON, 0600); err != nil {
			return fmt.Errorf("write default config: %w", err)
		}
	}

	// Optional: .gitignore for output and logs (to avoid committing generated files)
	_ = ensureGitignore(filepath.Join("output", ".gitignore"))
	_ = ensureGitignore(filepath.Join("logs", ".gitignore"))

	return nil
}

// ensureGitignore creates a simple .gitignore file if it does not exist.
// It never overwrites existing files.
func ensureGitignore(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	content := []byte("*\n!.gitignore\n")
	return os.WriteFile(path, content, 0644)
}
