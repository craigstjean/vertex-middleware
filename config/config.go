package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port string `yaml:"port"`
}

type KeyConfig struct {
	CredentialFile string `yaml:"credential_file"`
	ProjectID      string `yaml:"project_id"`
	Location       string `yaml:"location"`
	DefaultModel   string `yaml:"default_model"`
}

type Config struct {
	Server  ServerConfig         `yaml:"server"`
	APIKeys map[string]KeyConfig `yaml:"api_keys"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	for key, kc := range c.APIKeys {
		if kc.CredentialFile == "" {
			return fmt.Errorf("api key %q missing credential_file", key)
		}
		if kc.ProjectID == "" {
			return fmt.Errorf("api key %q missing project_id", key)
		}
		if kc.Location == "" {
			return fmt.Errorf("api key %q missing location", key)
		}
	}
	return nil
}
