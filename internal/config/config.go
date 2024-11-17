package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultConfigFile is the default configuration filename
	DefaultConfigFile = "ai-digest.json"
)

// Config represents the application configuration
type Config struct {
	DefaultIgnores []string `json:"defaultIgnores"`
	IgnoreFile     string   `json:"ignoreFile"`
}

// Manager handles configuration file operations
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	if configPath == "" {
		// Use current working directory by default
		cwd, err := os.Getwd()
		if err != nil {
			// Fallback to relative path if can't get CWD
			configPath = DefaultConfigFile
		} else {
			configPath = filepath.Join(cwd, DefaultConfigFile)
		}
	}
	return &Manager{
		configPath: configPath,
	}
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() Config {
	return Config{
		DefaultIgnores: []string{
			"node_modules",
			".git",
			"*.log",
			"*.swp",
			".DS_Store",
			"Thumbs.db",
			"*.tmp",
			"*.temp",
			".idea",
			".vscode",
		},
		IgnoreFile: ".aidigestignore",
	}
}

// Init initializes a new configuration file in the current directory
func (m *Manager) Init() error {
	if _, err := os.Stat(m.configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", m.configPath)
	}

	cfg := GetDefaultConfig()
	return m.Save(cfg)
}

// Load reads and parses the configuration file
func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			cfg := GetDefaultConfig()
			return &cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to file
func (m *Manager) Save(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Show returns the configuration as a formatted JSON string
func (m *Manager) Show() (string, error) {
	cfg, err := m.Load()
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	return string(data), nil
}

// GetConfigPath returns the current configuration file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// Exists checks if the configuration file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}
