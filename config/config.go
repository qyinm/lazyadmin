package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type SSHConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	PrivateKey string `yaml:"private_key"`
}

type DatabaseConfig struct {
	Driver   string     `yaml:"driver"`
	Host     string     `yaml:"host"`
	Port     int        `yaml:"port"`
	User     string     `yaml:"user"`
	Password string     `yaml:"password"`
	Name     string     `yaml:"name"`
	SSLMode  string     `yaml:"ssl_mode"`
	Path     string     `yaml:"path"`
	SSH      *SSHConfig `yaml:"ssh"`
}

type View struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Query       string `yaml:"query"`
}

type Config struct {
	ProjectName string         `yaml:"project_name"`
	Database    DatabaseConfig `yaml:"database"`
	Views       []View         `yaml:"views"`
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

	if cfg.Database.Port == 0 {
		switch cfg.Database.Driver {
		case "postgres":
			cfg.Database.Port = 5432
		case "mysql":
			cfg.Database.Port = 3306
		}
	}

	if cfg.Database.SSH != nil && cfg.Database.SSH.Port == 0 {
		cfg.Database.SSH.Port = 22
	}

	return &cfg, nil
}
