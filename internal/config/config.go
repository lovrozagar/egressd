package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the on-disk home-exit configuration.
type Config struct {
	Listen ListenConfig `yaml:"listen"`
	Users  UsersConfig  `yaml:"users"`
}

// ListenConfig holds proxy bind addresses.
type ListenConfig struct {
	SOCKS string `yaml:"socks"`
	HTTP  string `yaml:"http"`
}

// UsersConfig points at the local users database file.
type UsersConfig struct {
	File string `yaml:"file"`
}

// Default returns a localhost-only config suitable for v1.
func Default() Config {
	return Config{
		Listen: ListenConfig{
			SOCKS: "127.0.0.1:1080",
			HTTP:  "127.0.0.1:8080",
		},
		Users: UsersConfig{
			File: "users.json",
		},
	}
}

// FindPath resolves the config file path.
// Precedence:
//  1. path if non-empty (explicit -config / HOME_EXIT_CONFIG)
//  2. ./home-exit.yaml
//  3. ~/.config/home-exit/config.yaml
func FindPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if env := os.Getenv("HOME_EXIT_CONFIG"); env != "" {
		return env, nil
	}
	cwd := "home-exit.yaml"
	if _, err := os.Stat(cwd); err == nil {
		return cwd, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: home dir: %w", err)
	}
	return filepath.Join(home, ".config", "home-exit", "config.yaml"), nil
}

// Load reads YAML from path. Missing file returns Default with the path noted.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("config: read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: parse %s: %w", path, err)
	}
	if cfg.Listen.SOCKS == "" {
		cfg.Listen.SOCKS = Default().Listen.SOCKS
	}
	if cfg.Listen.HTTP == "" {
		cfg.Listen.HTTP = Default().Listen.HTTP
	}
	if cfg.Users.File == "" {
		cfg.Users.File = Default().Users.File
	}
	return cfg, nil
}

// UsersFilePath resolves the users DB path relative to the config file directory
// when the users file is not absolute.
func (c Config) UsersFilePath(configPath string) string {
	if filepath.IsAbs(c.Users.File) {
		return c.Users.File
	}
	dir := filepath.Dir(configPath)
	if dir == "." || dir == "" {
		return c.Users.File
	}
	// If config came from ~/.config/home-exit/config.yaml, keep users next to it.
	return filepath.Join(dir, c.Users.File)
}

// Summary returns a human-readable config summary for `status` / `up`.
func (c Config) Summary(configPath, usersPath string) string {
	return fmt.Sprintf(
		"config:  %s\nusers:   %s\nsocks:   %s\nhttp:    %s\n",
		configPath, usersPath, c.Listen.SOCKS, c.Listen.HTTP,
	)
}
