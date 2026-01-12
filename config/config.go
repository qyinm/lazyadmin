package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type View struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Query       string `yaml:"query"`
}

type Config struct {
	ProjectName string `yaml:"project_name"`
	Database    string `yaml:"database"`
	Views       []View `yaml:"views"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
