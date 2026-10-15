package cliconfig

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	baseURLEnv = "DNSLOG_BASE_URL"
	tokenEnv   = "DNSLOG_TOKEN"
	configEnv  = "DNSLOG_CONFIG"
)

// Config contains the connection settings required by the CLI.
type Config struct {
	BaseURL string `yaml:"base_url"`
	Token   string `yaml:"token"`
}

// Updates contains optional values to persist in the user configuration file.
type Updates struct {
	BaseURL *string
	Token   *string
}

// DefaultPath returns the default per-user CLI configuration path.
func DefaultPath() (string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(homeDirectory, ".config", "dnslog-go", "config.yaml"), nil
}

// Path returns DNSLOG_CONFIG when configured, otherwise the default path.
func Path() (string, error) {
	if configuredPath := os.Getenv(configEnv); configuredPath != "" {
		return configuredPath, nil
	}
	return DefaultPath()
}

// Load reads the user configuration and applies environment variable overrides.
func Load() (Config, error) {
	configPath, err := Path()
	if err != nil {
		return Config{}, err
	}

	configuration, err := loadFile(configPath)
	if err != nil {
		return Config{}, err
	}
	applyEnvironment(&configuration)
	if err := validate(configuration); err != nil {
		return Config{}, fmt.Errorf("invalid CLI configuration: %w", err)
	}
	configuration.BaseURL = strings.TrimRight(configuration.BaseURL, "/")
	return configuration, nil
}

// Save updates the user configuration file without importing environment values.
func Save(updates Updates) (string, error) {
	configPath, err := Path()
	if err != nil {
		return "", err
	}

	configuration, err := loadFile(configPath)
	if err != nil {
		return "", err
	}
	if updates.BaseURL != nil {
		configuration.BaseURL = *updates.BaseURL
	}
	if updates.Token != nil {
		configuration.Token = *updates.Token
	}
	if err := validate(configuration); err != nil {
		return "", fmt.Errorf("invalid CLI configuration: %w", err)
	}
	configuration.BaseURL = strings.TrimRight(configuration.BaseURL, "/")

	contents, err := yaml.Marshal(configuration)
	if err != nil {
		return "", fmt.Errorf("encode CLI configuration: %w", err)
	}
	if err := writeFile(configPath, contents); err != nil {
		return "", err
	}
	return configPath, nil
}

func loadFile(path string) (Config, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	var configuration Config
	if err := yaml.UnmarshalStrict(contents, &configuration); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return configuration, nil
}

func writeFile(path string, contents []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create %s: %w", directory, err)
	}

	temporaryFile, err := os.CreateTemp(directory, ".config-*.yaml")
	if err != nil {
		return fmt.Errorf("create temporary configuration: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	defer os.Remove(temporaryPath)

	if err := temporaryFile.Chmod(0o600); err != nil {
		temporaryFile.Close()
		return fmt.Errorf("secure temporary configuration: %w", err)
	}
	if _, err := temporaryFile.Write(contents); err != nil {
		temporaryFile.Close()
		return fmt.Errorf("write temporary configuration: %w", err)
	}
	if err := temporaryFile.Sync(); err != nil {
		temporaryFile.Close()
		return fmt.Errorf("sync temporary configuration: %w", err)
	}
	if err := temporaryFile.Close(); err != nil {
		return fmt.Errorf("close temporary configuration: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("secure %s: %w", path, err)
	}
	return nil
}

func applyEnvironment(configuration *Config) {
	if baseURL := os.Getenv(baseURLEnv); baseURL != "" {
		configuration.BaseURL = baseURL
	}
	if token := os.Getenv(tokenEnv); token != "" {
		configuration.Token = token
	}
}

func validate(configuration Config) error {
	if configuration.BaseURL == "" {
		return errors.New("DNSLOG_BASE_URL or base_url is required")
	}
	if configuration.Token == "" {
		return errors.New("DNSLOG_TOKEN or token is required")
	}

	parsedURL, err := url.Parse(configuration.BaseURL)
	if err != nil || parsedURL.Host == "" {
		return errors.New("base URL must be an absolute HTTP or HTTPS URL")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("base URL must use HTTP or HTTPS")
	}
	return nil
}
