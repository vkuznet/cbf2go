package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host string `json:"host" yaml:"host"`
		Port int    `json:"port" yaml:"port"`
	} `json:"server" yaml:"server"`

	Qdrant struct {
		URL        string `json:"url" yaml:"url"`
		Collection string `json:"collection" yaml:"collection"`
		Verbose    int    `json:"verbose" yaml:"verbose"`
	} `json:"qdrant" yaml:"qdrant"`
}

// LoadConfig reads JSON or YAML file based on extension
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}

	switch {
	case len(path) > 5 && path[len(path)-5:] == ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case len(path) > 4 && (path[len(path)-4:] == ".yml" || path[len(path)-5:] == ".yaml"):
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file type: %s", path)
	}

	return cfg, nil
}
