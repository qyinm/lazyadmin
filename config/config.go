package config

import (
	"fmt"
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
	Label    string     `yaml:"label"`
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
	ProjectName string           `yaml:"project_name"`
	Database    DatabaseConfig   `yaml:"database"` // Deprecated: used for backward compatibility
	Connections []DatabaseConfig `yaml:"connections"`
	Views       []View           `yaml:"views"`
}

// Load reads and parses the configuration file at the given path.
// It handles backward compatibility for single database configurations
// and validates all database connections, setting defaults where needed.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Backward compatibility: If Connections is empty but Database is present, use it.
	if len(cfg.Connections) == 0 && cfg.Database.Driver != "" {
		if cfg.Database.Label == "" {
			cfg.Database.Label = "Default"
		}
		cfg.Connections = append(cfg.Connections, cfg.Database)
	}

	if len(cfg.Connections) == 0 {
		return nil, fmt.Errorf("no database connections defined")
	}

	// Validate and set defaults for all connections
	for i := range cfg.Connections {
		c := &cfg.Connections[i]

		switch c.Driver {
		case "sqlite", "sqlite3", "postgres", "postgresql", "mysql":
		case "":
			return nil, fmt.Errorf("connection %d: database driver is required", i)
		default:
			return nil, fmt.Errorf("connection %d: unsupported database driver %q", i, c.Driver)
		}

		if c.Port == 0 {
			switch c.Driver {
			case "postgres", "postgresql":
				c.Port = 5432
			case "mysql":
				c.Port = 3306
			}
		}

		if c.SSH != nil && c.SSH.Port == 0 {
			c.SSH.Port = 22
		}

		if c.Label == "" {
			c.Label = fmt.Sprintf("Connection %d", i+1)
		}
	}

	// Ensure the deprecated field matches the first connection for any legacy code access
	if len(cfg.Connections) > 0 {
		cfg.Database = cfg.Connections[0]
	}

	return &cfg, nil
}

// Save writes the configuration to the given file path.
// It serializes the Config struct to YAML format.
func Save(path string, cfg *Config) error {
	type configToSave struct {
		ProjectName string           `yaml:"project_name"`
		Connections []DatabaseConfig `yaml:"connections"`
		Views       []View           `yaml:"views"`
	}

	toSave := configToSave{
		ProjectName: cfg.ProjectName,
		Connections: cfg.Connections,
		Views:       cfg.Views,
	}

	data, err := yaml.Marshal(&toSave)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
