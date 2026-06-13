// Package config holds harbrr's typed runtime configuration and the loader
// that assembles it from defaults, a config file, environment variables, and
// command-line flags. Structured config is always a typed struct — never a
// map[string]any (see AGENTS.md).
package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

// redactedMask is the placeholder substituted for secret values when a Config
// is rendered for logging.
const redactedMask = "***"

// Config is harbrr's complete runtime configuration. It starts intentionally
// small (Phase 0) and grows as consumers — the web server, database, and
// secrets store — are wired in later phases.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Log      LogConfig      `mapstructure:"log"`
	DataDir  string         `mapstructure:"data_dir"`
	Database DatabaseConfig `mapstructure:"database"`
	Secrets  SecretsConfig  `mapstructure:"secrets"`
}

// ServerConfig describes the management-API listener (not yet served).
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// LogConfig drives the zerolog logger built in internal/logger.
type LogConfig struct {
	Level  string `mapstructure:"level"`  // trace|debug|info|warn|error
	Format string `mapstructure:"format"` // console|json
}

// DatabaseConfig points at the SQLite database file. Postgres is deliberately
// not modeled yet (see AGENTS.md / dbinterface).
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// SecretsConfig selects the at-rest encryption key source. At most one of
// EncryptionKey (inline/env) or KeyFile (path) may be set; if neither is set,
// harbrr falls back to plaintext with a loud startup warning.
type SecretsConfig struct {
	EncryptionKey string `mapstructure:"encryption_key"`
	KeyFile       string `mapstructure:"key_file"`
}

// validLogLevels and validLogFormats are the accepted enum values for LogConfig.
var (
	validLogLevels  = map[string]struct{}{"trace": {}, "debug": {}, "info": {}, "warn": {}, "error": {}}
	validLogFormats = map[string]struct{}{"console": {}, "json": {}}
)

// Defaults returns the baseline configuration before any file, environment, or
// flag overrides are applied.
func Defaults() Config {
	return Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 7474,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "console",
		},
		DataDir:  "./data",
		Database: DatabaseConfig{Path: ""}, // derived under DataDir when empty
		Secrets:  SecretsConfig{},
	}
}

// DatabasePath returns the configured SQLite path, defaulting to
// "<DataDir>/harbrr.db" when unset.
func (c Config) DatabasePath() string {
	if c.Database.Path != "" {
		return c.Database.Path
	}
	return filepath.Join(c.DataDir, "harbrr.db")
}

// HasSecretKey reports whether an at-rest encryption key source is configured.
// When false, serve emits a loud plaintext warning.
func (c Config) HasSecretKey() bool {
	return c.Secrets.EncryptionKey != "" || c.Secrets.KeyFile != ""
}

// Validate checks the configuration for self-consistency, returning a wrapped
// error describing the first problem found.
func (c Config) Validate() error {
	if _, ok := validLogLevels[c.Log.Level]; !ok {
		return fmt.Errorf("config: invalid log.level %q", c.Log.Level)
	}
	if _, ok := validLogFormats[c.Log.Format]; !ok {
		return fmt.Errorf("config: invalid log.format %q", c.Log.Format)
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("config: server.port %d out of range 1-65535", c.Server.Port)
	}
	if c.DataDir == "" {
		return errors.New("config: data_dir must not be empty")
	}
	if c.Secrets.EncryptionKey != "" && c.Secrets.KeyFile != "" {
		return errors.New("config: set only one of secrets.encryption_key or secrets.key_file")
	}
	return nil
}

// Redacted returns a copy of the configuration with secret values masked, safe
// for logging. This is the redaction guarantee for the config surface.
func (c Config) Redacted() Config {
	redacted := c
	if redacted.Secrets.EncryptionKey != "" {
		redacted.Secrets.EncryptionKey = redactedMask
	}
	if redacted.Secrets.KeyFile != "" {
		redacted.Secrets.KeyFile = redactedMask
	}
	return redacted
}

// String renders the configuration with secrets redacted, so a Config is safe
// to interpolate into log lines.
func (c Config) String() string {
	r := c.Redacted()
	return fmt.Sprintf(
		"Config{server=%s:%d log=%s/%s data_dir=%s db=%s secrets.encryption_key=%q secrets.key_file=%q}",
		r.Server.Host, r.Server.Port, r.Log.Level, r.Log.Format,
		r.DataDir, r.DatabasePath(), r.Secrets.EncryptionKey, r.Secrets.KeyFile,
	)
}
