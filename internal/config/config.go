package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider string        `yaml:"provider"`
	Model    string        `yaml:"model"`
	Endpoint string        `yaml:"endpoint"`
	Timeout  time.Duration `yaml:"timeout"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "shellai", "config.yaml")
}

func defaults() *Config {
	return &Config{
		Provider: "ollama",
		Endpoint: "http://localhost:11434",
		Timeout:  60 * time.Second,
	}
}

func Load(path string) (*Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// first run — write defaults
		if err := Save(cfg, path); err != nil {
			return cfg, nil // still usable even if we can't write
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// fill in zero values with defaults
	if cfg.Provider == "" {
		cfg.Provider = "ollama"
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:11434"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}

	return cfg, nil
}

func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// SetField updates a single key in the config file
func SetField(path, key, val string) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}
	switch key {
	case "provider":
		cfg.Provider = val
	case "model":
		cfg.Model = val
	case "endpoint":
		cfg.Endpoint = val
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return Save(cfg, path)
}
