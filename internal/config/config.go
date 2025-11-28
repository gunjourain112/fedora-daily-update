package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

// CustomTask represents a user-defined command.
type CustomTask struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// Config holds the application configuration.
type Config struct {
	CustomTasks []CustomTask `json:"custom_tasks"`
}

// Manager handles loading and saving the configuration.
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager.
// It resolves the configuration path to ~/.config/my-updater/config.json
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "my-updater")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &Manager{
		configPath: filepath.Join(configDir, "config.json"),
	}, nil
}

// Load reads the configuration from the file.
// If the file does not exist, it returns a default configuration.
func (m *Manager) Load() (*Config, error) {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return &Config{CustomTasks: []CustomTask{}}, nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Warn("Failed to parse config file, using empty config", "err", err)
		return &Config{CustomTasks: []CustomTask{}}, nil
	}

	return &cfg, nil
}

// Save writes the configuration to the file.
func (m *Manager) Save(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}
